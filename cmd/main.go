package main

import (
	"avalon/pkg/domain"
	"avalon/pkg/presets"
	"context"
	"github.com/joho/godotenv"
	"os"
)

func main() {
	ctx := context.Background()
	_ = godotenv.Load()
	apiKey := os.Getenv("GEMINI_API_KEY")

	missions := presets.Missions5
	players := domain.GenerateActors(ctx, apiKey, presets.Roles5, missions)
	state := domain.GetInitialState(missions, players)
	domain.RunGame(state)
}
