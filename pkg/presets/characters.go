package presets

import (
	"avalon/pkg/dto"
)

func GetPlayersV2() []*dto.PlayerV2 {
	roles := GenRolesOrder(Roles5)
	return []*dto.PlayerV2{
		&dto.PlayerV2{
			Name:          "Петир",
			Role:          roles[0],
			Position:      1,
			CharacterType: "petir",
			Model:         "models/gemini-2.5-flash",
		},
		&dto.PlayerV2{
			Name:          "Варис",
			Role:          roles[1],
			Position:      2,
			CharacterType: "varis",
			Model:         "models/gemini-2.5-flash",
		},
		&dto.PlayerV2{
			Name:          "Серсея",
			Role:          roles[2],
			Position:      3,
			CharacterType: "cercei",
			Model:         "models/gemini-2.5-flash",
		},
		&dto.PlayerV2{
			Name:          "Тирион",
			Role:          roles[3],
			Position:      4,
			CharacterType: "tirion",
			Model:         "models/gemini-2.5-flash",
		},
		&dto.PlayerV2{
			Name:          "Барристан",
			Role:          roles[4],
			Position:      5,
			CharacterType: "baristan",
			Model:         "models/gemini-2.5-flash",
		},
	}
}
