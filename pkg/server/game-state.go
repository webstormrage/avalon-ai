package server

import (
	"avalon/pkg/store"
	"encoding/json"
	"net/http"
	"strconv"
)

func (h *GameHandler) GetGameState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	gameIDStr := r.URL.Query().Get("gameId")
	if gameIDStr == "" {
		http.Error(w, "missing gameId query param", http.StatusBadRequest)
		return
	}
	playerIDStr := r.URL.Query().Get("playerId")
	if playerIDStr == "" {
		http.Error(w, "missing playerId query param", http.StatusBadRequest)
		return
	}

	gameID, err := strconv.Atoi(gameIDStr)
	if err != nil {
		http.Error(w, "invalid gameId", http.StatusBadRequest)
		return
	}
	playerID, err := strconv.Atoi(playerIDStr)
	if err != nil {
		http.Error(w, "invalid playerId", http.StatusBadRequest)
		return
	}

	player, err := store.GetPlayerByID(h.Ctx, h.DB, playerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if player == nil {
		http.Error(w, "player not found", http.StatusNotFound)
		return
	}
	if player.GameID != gameID {
		http.Error(w, "player does not belong to game", http.StatusForbidden)
		return
	}

	state, err := getState(h, h.DB, gameID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	state.YourPosition = player.Position

	FilterStateByRequester(state, *player)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(state); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
