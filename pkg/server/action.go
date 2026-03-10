package server

import (
	"avalon/pkg/constants"
	"avalon/pkg/dto"
	"avalon/pkg/store"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type GameActionParams struct {
	Message      *string `json:"message,omitempty"`
	Success      *bool   `json:"success,omitempty"`
	Approve      *bool   `json:"approve,omitempty"`
	SquadNumbers []int   `json:"squadNumbers,omitempty"`
	Target       *int    `json:"target,omitempty"`
}

type GameAction struct {
	PlayerID int              `json:"playerId"`
	Name     string           `json:"name"`
	Params   GameActionParams `json:"params"`
}

type gameActionRequest struct {
	Action   GameAction       `json:"action"`
	PlayerID int              `json:"playerId"`
	Name     string           `json:"name"`
	Params   GameActionParams `json:"params"`
}

func normalizeAction(req gameActionRequest) GameAction {
	if req.Action.Name != "" {
		return req.Action
	}
	return GameAction{PlayerID: req.PlayerID, Name: req.Name, Params: req.Params}
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
	var speaker *dto.PlayerV2
	switch action.Name {
	case "propose_squad", "rate_squad":
		if len(game.TurnsOrder) == 0 {
			http.Error(w, "turnsOrder is empty", http.StatusForbidden)
			return
		}
		currentTurnPos := game.TurnsOrder[0]
		speaker, err = store.GetPlayerByPosition(h.Ctx, tx, game.ID, currentTurnPos)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if speaker == nil || speaker.ID != action.PlayerID {
			http.Error(w, "player is not current speaker", http.StatusForbidden)
			return
		}
	default:
		speaker, err = store.GetPlayerByPosition(h.Ctx, tx, game.ID, game.SpeakerPosition)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if speaker == nil || speaker.ID != action.PlayerID {
			http.Error(w, "player is not current speaker", http.StatusForbidden)
			return
		}
	}

	switch action.Name {
	case "propose_squad":
		err = applyProposeSquad(h, tx, game.ID, action)
	case "rate_squad":
		err = applyRateSquad(h, tx, game.ID, action)
	case "announce_squad":
		err = applyAnnounceSquad(h, tx, game.ID, action)
	case "vote_squad":
		err = applyVoteSquad(h, tx, game.ID, action)
	case "mission_action":
		err = applyMissionAction(h, tx, game.ID, action)
	case "announce_assassination":
		err = applyAnnounceAssassination(h, tx, game.ID, action)
	case "propose_assassination":
		err = applyProposeAssassinationAction(h, tx, game, speaker, action.Params)
	case "rate_assassination":
		err = applyRateAssassinationAction(h, tx, game, speaker, action.Params)
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

func applyProposeAssassinationAction(h *GameHandler, tx store.QueryRower, game *dto.GameV2, speaker *dto.PlayerV2, params GameActionParams) error {
	if game.GameState != constants.STATE_ASSASSIONATION_DISCUSSION && game.GameState != constants.STATE_ASSASSINATION_DISCUSSION {
		return fmt.Errorf("assassination discussion action allowed only in ASSASSIONATION_DISCUSSION")
	}
	if params.Target == nil {
		return fmt.Errorf("missing params.target")
	}
	targetPos := *params.Target
	targetPlayer, err := store.GetPlayerByPosition(h.Ctx, tx, game.ID, targetPos)
	if err != nil {
		return err
	}
	if targetPlayer == nil {
		return fmt.Errorf("unknown target position: %d", targetPos)
	}
	content := strings.TrimSpace(getActionMessage(params))
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
	return persistGameWithTurnsOrder(h, tx, game)
}

func applyRateAssassinationAction(h *GameHandler, tx store.QueryRower, game *dto.GameV2, speaker *dto.PlayerV2, params GameActionParams) error {
	if game.GameState != constants.STATE_ASSASSIONATION_DISCUSSION && game.GameState != constants.STATE_ASSASSINATION_DISCUSSION {
		return fmt.Errorf("assassination discussion action allowed only in ASSASSIONATION_DISCUSSION")
	}
	content := strings.TrimSpace(getActionMessage(params))
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
	return persistGameWithTurnsOrder(h, tx, game)
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
