package presets

import (
	"avalon/pkg/constants"
	"math/rand"
	"time"
)

type Role struct {
	Name        string
	Description string
	Count       int
}

var Roles5 = []Role{
	Role{
		constants.ROLE_MERLIN,
		"Член команды 'Синие'. Знает кто из игроков относится к команде 'Красные'",
		1,
	},
	Role{
		constants.ROLE_ARTHURS_LOYAL,
		"Член команды 'Синие'",
		2,
	},
	Role{
		constants.ROLE_ASSASSIN,
		"Член команды 'Красные'. Знает кто из игроков относится к команде 'Красные'.",
		1,
	},
	Role{
		constants.ROLE_MORDRED_MINION,
		"Член команды 'Красные'. Знает кто из игроков относится к команде 'Красные'",
		1,
	},
}

func GenRolesOrder(roles []Role) []string {
	var deck []string

	// собираем "колоду"
	for _, role := range roles {
		for i := 0; i < role.Count; i++ {
			deck = append(deck, role.Name)
		}
	}

	// инициализация генератора случайных чисел
	rand.Seed(time.Now().UnixNano())

	// перемешивание (алгоритм Фишера–Йейтса)
	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})

	return deck
}

var Missions5 = []int{2, 3, 2, 3, 3}
