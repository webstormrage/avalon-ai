package server

import (
	"fmt"
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

	state, err := h.stateMachine(gameID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, err = fmt.Fprint(w, state)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
