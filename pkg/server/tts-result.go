package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

func (h *GameHandler) TtsResult(w http.ResponseWriter, r *http.Request) {
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

	fileName := fmt.Sprintf("%d.mp3", id)

	fullPath := filepath.Join(h.MediaDir, fileName)

	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("Content-Length", strconv.FormatInt(stat.Size(), 10))
	w.Header().Set("Content-Disposition", `inline; filename="`+stat.Name()+`"`)

	http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
}
