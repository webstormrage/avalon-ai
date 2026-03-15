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

	leaderPosition := rand.Intn(len(players)-1) + 1
	orderedPlayers := make([]dto.PlayerV2, 0, len(players))
	for _, p := range players {
		if p == nil {
			continue
		}
		orderedPlayers = append(orderedPlayers, *p)
	}
	initialTurnsOrder, err := orderedAllFromLeader(orderedPlayers, leaderPosition)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, m := range missions {
		if m == nil {
			continue
		}
		if m.Priority == 1 {
			m.Squad = []int{leaderPosition}
			break
		}
	}

	game := &dto.GameV2{
		MissionPriority: 1,
		LeaderPosition:  leaderPosition,
		SpeakerPosition: leaderPosition,
		TurnsOrder:      initialTurnsOrder,
		SkipsCount:      0,
		Wins:            0,
		Fails:           0,
		Phase:           constants.STATE_DISCUSSION,
	}

	gameID, err := store.CreateGameTransaction(ctx, h.DB, game, missions, players)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	createdGame, err := store.GetGame(ctx, h.DB, gameID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if createdGame == nil {
		http.Error(w, "game not found after create", http.StatusInternalServerError)
		return
	}
	if err := persistGameWithTurnsOrder(h, h.DB, createdGame); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(map[string]int{
		"gameId": int(gameID),
	})
}
