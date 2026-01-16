package server

import (
	"avalon/pkg/dto"
	"avalon/pkg/presets"
	"avalon/pkg/store"
	"context"
	"database/sql"
	"encoding/json"
	"math/rand"
	"net/http"
	"time"
)

type GameHandler struct {
	DB *sql.DB
}

func (h *GameHandler) CreateGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	missions := presets.GetMissionsV2()

	players := presets.GetPlayersV2()

	game := &dto.GameV2{
		MissionPriority: 1,
		LeaderPosition:  rand.Intn(len(players)),
		SkipsCount:      0,
		Wins:            0,
		Fails:           0,
	}

	gameID, err := store.CreateGameTransaction(ctx, h.DB, game, missions, players)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(map[string]int{
		"gameId": int(gameID),
	})
}
