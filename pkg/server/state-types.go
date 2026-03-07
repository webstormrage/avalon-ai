package server

import "avalon/pkg/dto"

type GameState struct {
	Game           dto.GameV2      `json:"game"`
	Prompt         *dto.Prompt     `json:"prompt,omitempty"`
	Players        []dto.PlayerV2  `json:"players,omitempty"`
	CurrentEvent   string          `json:"currentEvent,omitempty"`
	RequiredAction *RequiredAction `json:"requiredAction,omitempty"`
}

type RequiredAction struct {
	PlayerID  int               `json:"playerId"`
	Name      string            `json:"name"`
	ParamsDef map[string]string `json:"paramsDef"`
}
