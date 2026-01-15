package main

import (
	"avalon/pkg/action"
	"avalon/pkg/domain"
	"avalon/pkg/dto"
	"avalon/pkg/presets"
	"context"
	"github.com/joho/godotenv"
	"math/rand"
	"os"
)

func main() {
	ctx := context.Background()
	_ = godotenv.Load()
	apiKey := os.Getenv("GEMINI_API_KEY")

	missions := presets.Missions5
	players := domain.GenerateActors(ctx, apiKey, presets.Roles5, missions)
	state := &dto.GameState{
		Missions:     missions,
		Players:      players,
		MissionIndex: 0,
		LeaderIndex:  rand.Intn(len(players)),
		SkipsCount:   0,
		Wins:         0,
		Fails:        0,
		Logs:         []action.Action{},
	}
	domain.RunGame(state)
}
