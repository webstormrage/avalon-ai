package server

import (
	"avalon/pkg/presets"
	"avalon/pkg/prompts"
	"avalon/pkg/store"
	"encoding/json"
	"net/http"
)

type renderSystemPromptRequest struct {
	GameState GameState `json:"gameState"`
	PlayerID  int       `json:"playerId"`
}

type renderSystemPromptResponse struct {
	SystemPrompt string `json:"systemPrompt"`
}

func (h *GameHandler) RenderSystemPrompt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req renderSystemPromptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.PlayerID == 0 {
		http.Error(w, "missing playerId", http.StatusBadRequest)
		return
	}
	if req.GameState.Game.ID == 0 {
		http.Error(w, "missing gameState.game.id", http.StatusBadRequest)
		return
	}

	player, err := store.GetPlayerByID(h.Ctx, h.DB, req.PlayerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if player == nil {
		http.Error(w, "player not found", http.StatusNotFound)
		return
	}
	if player.GameID != req.GameState.Game.ID {
		http.Error(w, "player does not belong to game", http.StatusForbidden)
		return
	}

	FilterStateByRequester(&req.GameState, *player)

	missions, err := store.GetMissionsByGameID(h.Ctx, h.DB, req.GameState.Game.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	systemPrompt := prompts.GetSystemPrompt(prompts.SystemPromptProps{
		Name:     player.Name,
		Players:  req.GameState.Players,
		Roles:    presets.Roles5,
		Role:     player.Role,
		Missions: missions,
	})

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(renderSystemPromptResponse{
		SystemPrompt: systemPrompt,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
