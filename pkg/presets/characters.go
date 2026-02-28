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
			Name:          "Ван Дипсик",
			Role:          roles[2],
			Position:      3,
			CharacterType: "deepseek",
			Model:         "deepseek/deepseek-v3.2-exp",
		},
		&dto.PlayerV2{
			Name:          "Кими-сама",
			Role:          roles[3],
			Position:      4,
			CharacterType: "kimi",
			Model:         "moonshotai/kimi-k2-thinking",
		},
		&dto.PlayerV2{
			Name:          "Сеньор Мистраль",
			Role:          roles[4],
			Position:      5,
			CharacterType: "mistral",
			Model:         "mistralai/mixtral-8x22b-instruct",
		},
	}
}
