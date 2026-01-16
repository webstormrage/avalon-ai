package presets

import (
	"avalon/pkg/dto"
)

type CharacterConfig struct {
	Name  string
	Mood  string
	Voice string
}

var Characters5 = []CharacterConfig{
	{
		"Сэр Эдрик",
		"Манера речи осторожная, обтекаемая, с недосказанностью. Не любит рисковать, предпочитает играть наверняка.",
		"Puck",
	},
	{
		"Сэр Лиорен",
		"Манера речи - уверенная, развёрнутая, любит рассуждать вслух. Готов рисковать, если может это красиво объяснить",
		"Iapetus",
	},
	{
		"Сэр Кайлен",
		"Манера речи - короткая, прямая, без украшений. Легко идёт на риск, не боясь последствий",
		"Enceladus",
	},
	{
		"Сэр Бранн",
		"Манера речи - редкая, тяжёлая, по делу.  Рискует только если уверен почти полностью",
		"Charon",
	},
	{
		"Леди Ивеллин",
		"Манера речи -  быстрая, колкая, с акцентом на детали. Любит острвые ходы и провокации.",
		"Zephyr",
	},
}

func GetPlayersV2() []*dto.PlayerV2 {
	roles := GenRolesOrder(Roles5)
	return []*dto.PlayerV2{
		&dto.PlayerV2{
			Name:     "Сэр Эдрик",
			Role:     roles[0],
			Position: 1,
			Voice:    "Puck",
			Mood:     "Манера речи осторожная, обтекаемая, с недосказанностью. Не любит рисковать, предпочитает играть наверняка.",
			Model:    "models/gemini-2.5-flash",
		},
		&dto.PlayerV2{
			Name:     "Сэр Лиорен",
			Role:     roles[1],
			Position: 2,
			Mood:     "Манера речи - уверенная, развёрнутая, любит рассуждать вслух. Готов рисковать, если может это красиво объяснить",
			Voice:    "Iapetus",
			Model:    "models/gemini-2.5-flash",
		},
		&dto.PlayerV2{
			Name:     "Сэр Кайлен",
			Role:     roles[2],
			Position: 3,
			Mood:     "Манера речи - короткая, прямая, без украшений. Легко идёт на риск, не боясь последствий",
			Voice:    "Enceladus",
			Model:    "models/gemini-2.5-flash",
		},
		&dto.PlayerV2{
			Name:     "Сэр Бранн",
			Role:     roles[3],
			Position: 4,
			Mood:     "Манера речи - редкая, тяжёлая, по делу.  Рискует только если уверен почти полностью",
			Voice:    "Charon",
			Model:    "models/gemini-2.5-flash",
		},
		&dto.PlayerV2{
			Name:     "Леди Ивеллин",
			Role:     roles[4],
			Position: 5,
			Mood:     "Манера речи -  быстрая, колкая, с акцентом на детали. Любит острвые ходы и провокации.",
			Voice:    "Zephyr",
			Model:    "models/gemini-2.5-flash",
		},
	}
}
