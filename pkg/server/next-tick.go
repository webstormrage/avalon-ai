package server

import (
	"avalon/pkg/store"
	"encoding/json"
	"net/http"
	"strconv"
)

func (h *GameHandler) NextTick(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// получить gameId из query params
	gameIDStr := r.URL.Query().Get("gameId")
	if gameIDStr == "" {
		http.Error(w, "missing gameId query param", http.StatusBadRequest)
		return
	}

	gameID, err := strconv.Atoi(gameIDStr)
	if err != nil {
		http.Error(w, "invalid gameId", http.StatusBadRequest)
		return
	}

	ctx := h.Ctx

	game, err := store.GetGame(ctx, h.DB, gameID)

	if err != nil {
		http.Error(w, "Failed to find game", http.StatusInternalServerError)
		return
	}

	prompts, err := store.GetPromptsNotCompletedByGameID(ctx, h.DB, gameID)

	if err != nil || len(prompts) > 1 {
		http.Error(w, "too many uncompleted prompts", http.StatusInternalServerError)
		return
	}

	err = h.stateMachine(game, prompts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]int{
		"gameId": gameID,
	})
}
