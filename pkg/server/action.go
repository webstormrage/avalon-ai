package server

import (
	"avalon/pkg/constants"
	"avalon/pkg/dto"
	"avalon/pkg/prompts"
	"avalon/pkg/store"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type gameAction struct {
	PlayerID int            `json:"playerId"`
	Name     string         `json:"name"`
	Params   map[string]any `json:"params"`
}

type gameActionRequest struct {
	Action   gameAction     `json:"action"`
	PlayerID int            `json:"playerId"`
	Name     string         `json:"name"`
	Params   map[string]any `json:"params"`
}

func normalizeAction(req gameActionRequest) gameAction {
	if req.Action.Name != "" {
		return req.Action
	}
	return gameAction{PlayerID: req.PlayerID, Name: req.Name, Params: req.Params}
}

func (h *GameHandler) ApplyGameAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req gameActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	action := normalizeAction(req)
	if action.PlayerID == 0 || action.Name == "" {
		http.Error(w, "missing action.playerId or action.name", http.StatusBadRequest)
		return
	}

	tx, err := h.DB.BeginTx(h.Ctx, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	player, err := store.GetPlayerByID(h.Ctx, tx, action.PlayerID)
	if err != nil || player == nil {
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			http.Error(w, "player not found", http.StatusNotFound)
		}
		return
	}
	game, err := store.GetGame(h.Ctx, tx, player.GameID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	speaker, err := store.GetPlayerByPosition(h.Ctx, tx, game.ID, game.SpeakerPosition)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if speaker == nil || speaker.ID != action.PlayerID {
		http.Error(w, "player is not current speaker", http.StatusForbidden)
		return
	}

	if action.Name == "propose_assassination" || action.Name == "rate_assassination" {
		if game.GameState != constants.STATE_ASSASSIONATION_DISCUSSION && game.GameState != constants.STATE_ASSASSINATION_DISCUSSION {
			http.Error(w, "assassination discussion action allowed only in ASSASSIONATION_DISCUSSION", http.StatusConflict)
			return
		}
		var actionErr error
		if action.Name == "propose_assassination" {
			actionErr = applyProposeAssassinationAction(h, tx, game, speaker, action.Params)
		} else {
			actionErr = applyRateAssassinationAction(h, tx, game, speaker, action.Params)
		}
		if actionErr != nil {
			http.Error(w, actionErr.Error(), http.StatusBadRequest)
			return
		}
		state, err := getState(h, tx, game.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := tx.Commit(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(state)
		return
	}

	pendingPrompts, err := store.GetPromptsNotCompletedByGameID(h.Ctx, tx, game.ID)
	if err != nil || len(pendingPrompts) == 0 {
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			http.Error(w, "no active prompt", http.StatusConflict)
		}
		return
	}

	var responseText string
	switch action.Name {
	case "propose_squad":
		if game.GameState != constants.STATE_DISCUSSION || game.SpeakerPosition != game.LeaderPosition {
			http.Error(w, "propose_squad allowed only for leader in DISCUSSION", http.StatusConflict)
			return
		}
		responseText, err = buildProposeSquadResponse(action.Params, game.ID, tx, h)
	case "rate_squad":
		if game.GameState != constants.STATE_DISCUSSION || game.SpeakerPosition == game.LeaderPosition {
			http.Error(w, "rate_squad allowed only for non-leader in DISCUSSION", http.StatusConflict)
			return
		}
		responseText, err = buildRateSquadResponse(action.Params)
	case "announce_squad":
		if game.GameState != constants.STATE_VOTING || game.SpeakerPosition != game.LeaderPosition {
			http.Error(w, "announce_squad allowed only for leader in VOTING", http.StatusConflict)
			return
		}
		responseText, err = buildProposeSquadResponse(action.Params, game.ID, tx, h)
	case "vote_squad":
		if game.GameState != constants.STATE_VOTING || game.SpeakerPosition == game.LeaderPosition {
			http.Error(w, "vote_squad allowed only for non-leader in VOTING", http.StatusConflict)
			return
		}
		responseText, err = buildVoteSquadResponse(action.Params)
	case "mission_action":
		if game.GameState != constants.STATE_MISSION {
			http.Error(w, "mission_action allowed only in MISSION", http.StatusConflict)
			return
		}
		responseText, err = buildMissionActionResponse(action.Params)
	case "announce_assassination":
		if game.GameState != constants.STATE_ASSASSIONATION && game.GameState != constants.STATE_ASSASSINATION {
			http.Error(w, "announce_assassination allowed only in ASSASSINATION", http.StatusConflict)
			return
		}
		responseText, err = buildAnnounceAssassinationResponse(action.Params, game.ID, tx, h)
	default:
		http.Error(w, "unsupported action name", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	prompt := pendingPrompts[0]
	prompt.Response = responseText
	prompt.Status = constants.STATUS_HAS_RESPONSE
	if err := store.UpdatePrompt(h.Ctx, tx, &prompt); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch action.Name {
	case "propose_squad":
		err = applyLeaderDiscussionPrompt(h, tx, game.ID)
	case "rate_squad":
		err = applySpeakerDiscussionPrompt(h, tx, game.ID)
	case "announce_squad":
		err = applyLeaderVotingPrompt(h, tx, game.ID)
	case "vote_squad":
		err = applySpeakerVotingPrompt(h, tx, game.ID)
	case "mission_action":
		err = applyMissionPrompt(h, tx, game.ID)
	case "announce_assassination":
		err = applyAssassinationPrompt(h, tx, game.ID)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	state, err := getState(h, tx, game.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(state)
}

func buildProposeSquadResponse(params map[string]any, gameID int, tx store.QueryRower, h *GameHandler) (string, error) {
	if params == nil {
		return "", fmt.Errorf("missing action.params")
	}
	message, ok := params["message"].(string)
	if !ok {
		return "", fmt.Errorf("params.message must be string")
	}
	squadRaw, ok := params["squad"]
	if !ok {
		return "", fmt.Errorf("missing params.squad")
	}
	squadPositions, err := parseSquadPositions(squadRaw)
	if err != nil {
		return "", err
	}
	players, err := store.GetPlayersByGameID(h.Ctx, tx, gameID)
	if err != nil {
		return "", err
	}
	nameByPos := map[int]string{}
	for _, p := range players {
		nameByPos[p.Position] = p.Name
	}
	squadNames := make([]string, 0, len(squadPositions))
	for _, pos := range squadPositions {
		name, ok := nameByPos[pos]
		if !ok {
			return "", fmt.Errorf("unknown squad position: %d", pos)
		}
		squadNames = append(squadNames, name)
	}
	response := strings.TrimSpace(message)
	line := "Выставить: " + strings.Join(squadNames, ", ")
	if response == "" {
		return line, nil
	}
	return response + "\n" + line, nil
}

func buildRateSquadResponse(params map[string]any) (string, error) {
	if params == nil {
		return "", fmt.Errorf("missing action.params")
	}
	message, ok := params["message"].(string)
	if !ok {
		return "", fmt.Errorf("params.message must be string")
	}
	message = strings.TrimSpace(message)
	if message == "" {
		return "", fmt.Errorf("params.message must not be empty")
	}
	return message, nil
}

func buildVoteSquadResponse(params map[string]any) (string, error) {
	if params == nil {
		return "", fmt.Errorf("missing action.params")
	}
	approve, ok := params["approve"].(bool)
	if !ok {
		return "", fmt.Errorf("params.approve must be boolean")
	}
	vote := "ПРОТИВ"
	if approve {
		vote = "ЗА"
	}
	response := "Голосовать: " + vote
	if prompts.ExtractVote(response) == "" {
		return "", fmt.Errorf("failed to build vote response")
	}
	return response, nil
}

func buildMissionActionResponse(params map[string]any) (string, error) {
	if params == nil {
		return "", fmt.Errorf("missing action.params")
	}
	success, ok := params["success"].(bool)
	if !ok {
		return "", fmt.Errorf("params.success must be boolean")
	}
	res := "ПРОВАЛ"
	if success {
		res = "УСПЕХ"
	}
	response := "Совершить: " + res
	if prompts.ExtractMissionResult(response) != res {
		return "", fmt.Errorf("failed to build mission response")
	}
	return response, nil
}

func buildAnnounceAssassinationResponse(params map[string]any, gameID int, tx store.QueryRower, h *GameHandler) (string, error) {
	if params == nil {
		return "", fmt.Errorf("missing action.params")
	}
	message, ok := params["message"].(string)
	if !ok {
		return "", fmt.Errorf("params.message must be string")
	}
	targetRaw, ok := params["target"]
	if !ok {
		return "", fmt.Errorf("missing params.target")
	}
	targetPos, err := parseIntLike(targetRaw)
	if err != nil {
		return "", fmt.Errorf("params.target must be integer")
	}
	targetPlayer, err := store.GetPlayerByPosition(h.Ctx, tx, gameID, targetPos)
	if err != nil {
		return "", err
	}
	if targetPlayer == nil {
		return "", fmt.Errorf("unknown target position: %d", targetPos)
	}
	response := strings.TrimSpace(message)
	line := "Выбрать: " + targetPlayer.Name
	if response == "" {
		return line, nil
	}
	return response + "\n" + line, nil
}

func parseIntLike(v any) (int, error) {
	switch t := v.(type) {
	case float64:
		return int(t), nil
	case int:
		return t, nil
	case string:
		return strconv.Atoi(strings.TrimSpace(t))
	default:
		return 0, fmt.Errorf("not int-like")
	}
}

func applyProposeAssassinationAction(h *GameHandler, tx store.QueryRower, game *dto.GameV2, speaker *dto.PlayerV2, params map[string]any) error {
	if params == nil {
		return fmt.Errorf("missing action.params")
	}
	message, ok := params["message"].(string)
	if !ok {
		return fmt.Errorf("params.message must be string")
	}
	targetRaw, ok := params["target"]
	if !ok {
		return fmt.Errorf("missing params.target")
	}
	targetPos, err := parseIntLike(targetRaw)
	if err != nil {
		return fmt.Errorf("params.target must be integer")
	}
	targetPlayer, err := store.GetPlayerByPosition(h.Ctx, tx, game.ID, targetPos)
	if err != nil {
		return err
	}
	if targetPlayer == nil {
		return fmt.Errorf("unknown target position: %d", targetPos)
	}
	content := strings.TrimSpace(message)
	line := "Выставить: " + targetPlayer.Name
	if content == "" {
		content = line
	} else {
		content += "\n" + line
	}
	if err := store.CreateEvent(h.Ctx, tx, &dto.Event{GameID: game.ID, PlayerID: speaker.ID, Type: constants.EVENT_PLAYER_SPEECH, Content: content}); err != nil {
		return err
	}
	nextPos, err := nextRedSpeakerPosition(h, tx, game.ID, game.SpeakerPosition)
	if err != nil {
		return err
	}
	game.SpeakerPosition = nextPos
	return store.UpdateGame(h.Ctx, tx, game)
}

func applyRateAssassinationAction(h *GameHandler, tx store.QueryRower, game *dto.GameV2, speaker *dto.PlayerV2, params map[string]any) error {
	if params == nil {
		return fmt.Errorf("missing action.params")
	}
	message, ok := params["message"].(string)
	if !ok {
		return fmt.Errorf("params.message must be string")
	}
	content := strings.TrimSpace(message)
	if content == "" {
		return fmt.Errorf("params.message must not be empty")
	}
	if err := store.CreateEvent(h.Ctx, tx, &dto.Event{GameID: game.ID, PlayerID: speaker.ID, Type: constants.EVENT_PLAYER_SPEECH, Content: content}); err != nil {
		return err
	}
	nextPos, err := nextRedSpeakerPosition(h, tx, game.ID, game.SpeakerPosition)
	if err != nil {
		return err
	}
	nextSpeaker, err := store.GetPlayerByPosition(h.Ctx, tx, game.ID, nextPos)
	if err != nil {
		return err
	}
	if nextSpeaker == nil {
		return fmt.Errorf("next red speaker not found")
	}
	game.SpeakerPosition = nextPos
	if nextSpeaker.Role == constants.ROLE_ASSASSIN {
		game.GameState = constants.STATE_ASSASSINATION
	}
	return store.UpdateGame(h.Ctx, tx, game)
}

func nextRedSpeakerPosition(h *GameHandler, tx store.QueryRower, gameID int, currentPos int) (int, error) {
	players, err := store.GetPlayersByGameID(h.Ctx, tx, gameID)
	if err != nil {
		return 0, err
	}
	reds := make([]int, 0)
	for _, p := range players {
		if p.Role == constants.ROLE_ASSASSIN || p.Role == constants.ROLE_MORDRED_MINION {
			reds = append(reds, p.Position)
		}
	}
	if len(reds) == 0 {
		return 0, fmt.Errorf("no red players in game")
	}
	next := 0
	minPos := reds[0]
	for _, pos := range reds {
		if pos < minPos {
			minPos = pos
		}
		if pos > currentPos && (next == 0 || pos < next) {
			next = pos
		}
	}
	if next == 0 {
		next = minPos
	}
	return next, nil
}

func parseSquadPositions(raw any) ([]int, error) {
	items, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("params.squad must be array")
	}
	out := make([]int, 0, len(items))
	for _, item := range items {
		switch v := item.(type) {
		case float64:
			out = append(out, int(v))
		case string:
			n, err := strconv.Atoi(strings.TrimSpace(v))
			if err != nil {
				return nil, fmt.Errorf("invalid squad value: %v", item)
			}
			out = append(out, n)
		default:
			return nil, fmt.Errorf("invalid squad value type")
		}
	}
	return out, nil
}
