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

	idRaw := r.URL.Query().Get("id")
	if idRaw == "" {
		http.Error(w, "missing id query param", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idRaw)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	prompt, err := store.GetPromptByID(h.Ctx, h.DB, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sIdRaw := r.URL.Query().Get("sid")
	if sIdRaw == "" {
		http.Error(w, "missing sId query param", http.StatusBadRequest)
		return
	}

	sId, err := strconv.Atoi(sIdRaw)
	if err != nil {
		http.Error(w, "invalid sId", http.StatusBadRequest)
		return
	}

	speaker, err := store.GetPlayerByID(h.Ctx, h.DB, sId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	audio, err := h.TtsAgent.Send(
		*speaker,
		prompt.Response,
		&speaker.VoiceStyle,
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

	fileName := fmt.Sprintf("%d.mp3", prompt.ID)

	fullPath := filepath.Join(h.MediaDir, fileName)

	if err := os.WriteFile(fullPath, audio, 0644); err != nil {
		http.Error(w, "cannot save file", http.StatusInternalServerError)
		return
	}

	resp := map[string]string{
		"status": "ok",
		"name":   fileName,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
