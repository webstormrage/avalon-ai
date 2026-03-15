package server

import (
	"avalon/pkg/dto"
	"avalon/pkg/server/actionhandlers"
	"avalon/pkg/store"
	"encoding/json"
	"net/http"
)

type GameActionParams = actionhandlers.GameActionParams
type GameAction = actionhandlers.GameAction

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
	case "propose_squad", "rate_squad", "announce_squad", "vote_squad", "mission_action", "propose_assassination", "rate_assassination":
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
		err = actionhandlers.ApplyProposeSquad(h.Ctx, tx, game.ID, action)
	case "rate_squad":
		err = actionhandlers.ApplyRateSquad(h.Ctx, tx, game.ID, action)
	case "announce_squad":
		err = actionhandlers.ApplyAnnounceSquad(h.Ctx, tx, game.ID, action)
	case "vote_squad":
		err = actionhandlers.ApplyVoteSquad(h.Ctx, tx, game.ID, action)
	case "mission_action":
		err = actionhandlers.ApplyMissionAction(h.Ctx, tx, game.ID, action)
	case "announce_assassination":
		err = actionhandlers.ApplyAnnounceAssassination(h.Ctx, tx, game.ID, action)
	case "propose_assassination":
		err = actionhandlers.ApplyProposeAssassination(h.Ctx, tx, game.ID, action)
	case "rate_assassination":
		err = actionhandlers.ApplyRateAssassination(h.Ctx, tx, game.ID, action)
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
