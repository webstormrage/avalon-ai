package presets

import (
	"avalon/pkg/dto"
)

func GetPlayersV2() []*dto.PlayerV2 {
	roles := GenRolesOrder(Roles5)
	return []*dto.PlayerV2{
		&dto.PlayerV2{
			Name:          "Лорд Гемини",
			Role:          roles[0],
			Position:      1,
			CharacterType: "gemini",
			Model:         "google/gemini-3.1-pro-preview",
		},
		&dto.PlayerV2{
			Name:          "Сэр ЧатГпт",
			Role:          roles[1],
			Position:      2,
			CharacterType: "chatGPT",
			Model:         "openai/gpt-5.2-chat",
		},
		&dto.PlayerV2{
			Name:          "Мастер Грок",
			Role:          roles[2],
			Position:      3,
			CharacterType: "grok",
			Model:         "x-ai/grok-4",
		},
		&dto.PlayerV2{
			Name:          "Заи-сама",
			Role:          roles[3],
			Position:      4,
			CharacterType: "zai",
			Model:         "z-ai/glm-5",
		},
		&dto.PlayerV2{
			Name:          "Господин Клод",
			Role:          roles[4],
			Position:      5,
			CharacterType: "claude",
			Model:         "anthropic/claude-opus-4.5",
		},
	}
}
