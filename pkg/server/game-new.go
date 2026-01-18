package server

import (
	"avalon/pkg/constants"
	"avalon/pkg/dto"
	"avalon/pkg/presets"
	"avalon/pkg/store"
	"context"
	"database/sql"
	"encoding/json"
	"math/rand"
	"net/http"
)

type GameHandler struct {
	DB       *sql.DB
	Agent    dto.Agent
	Ctx      context.Context
	TtsAgent dto.TtsAgent
	MediaDir string
}

func (h *GameHandler) CreateGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx := h.Ctx

	missions := presets.GetMissionsV2()

	players := presets.GetPlayersV2()

	leaderPosition := rand.Intn(len(players)) + 1

	game := &dto.GameV2{
		MissionPriority: 1,
		LeaderPosition:  leaderPosition,
		SpeakerPosition: leaderPosition,
		SkipsCount:      0,
		Wins:            0,
		Fails:           0,
		GameState:       constants.STATE_DISCUSSION,
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
