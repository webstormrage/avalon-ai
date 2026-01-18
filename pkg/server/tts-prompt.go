package server

import (
	"avalon/pkg/store"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

func (h *GameHandler) TtsPrompt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	game, err := store.GetGame(h.Ctx, h.DB, gameID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	speaker, err := store.GetPlayerByPosition(h.Ctx, h.DB, gameID, game.SpeakerPosition)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pendingPrompts, err := store.GetPromptsNotCompletedByGameID(h.Ctx, h.DB, gameID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	prompt := pendingPrompts[0]

	if len(pendingPrompts) == 0 || len(pendingPrompts[0].Response) == 0 {
		http.Error(w, "No active llm response found", http.StatusBadRequest)
		return
	}

	audio, err := h.TtsAgent.Send(
		"gemini-2.5-flash-preview-tts", // TODO: store at player
		speaker.Voice,
		prompt.Response,
		nil, //TODO: extract from model response
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// гарантируем наличие media/
	if err := os.MkdirAll(h.MediaDir, 0755); err != nil {
		http.Error(w, "cannot create media dir", http.StatusInternalServerError)
		return
	}

	fileName := fmt.Sprintf("[Game-%d] prompt-%d %s (%s).wav", gameID, prompt.ID, speaker.Name, speaker.Voice)

	fullPath := filepath.Join(h.MediaDir, fileName)

	if err := os.WriteFile(fullPath, audio, 0644); err != nil {
		http.Error(w, "cannot save file", http.StatusInternalServerError)
		return
	}

	resp := map[string]string{
		"status": "ok",
		"path":   fullPath,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
