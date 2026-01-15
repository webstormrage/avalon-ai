package prompts

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type Role struct {
	Name        string
	Description string
	Count       int
}

var Roles5 = []Role{
	Role{
		"Мерлин",
		"Член команды 'Синие'. Знает кто из игроков относится к команде 'Красные'",
		1,
	},
	Role{
		"Слуга Артура",
		"Член команды 'Синие'",
		2,
	},
	Role{
		"Ассасин",
		"Член команды 'Красные'. Знает кто из игроков относится к команде 'Красные'. Среди всех игроков только 1 Ассасин",
		1,
	},
	Role{
		"Миньон Мордреда",
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

type Props struct {
	Self        string
	Mood        string
	Risk        string
	Players     []string
	Roles       []Role
	Role        string
	RoleContext string
	Missions    []int
}

func formatRoles(roles []Role) string {
	rolesP := ""
	for _, role := range roles {
		rolesP = rolesP + "\n" + role.Name + " (число игроков с ролью - " + strconv.Itoa(role.Count) + ")\n" + role.Description
	}
	return rolesP
}

func formatMissions(missions []int) string {
	missionsP := ""
	for index, count := range missions {
		missionsP = missionsP + "\n" + "Миссия " + strconv.Itoa(index+1) + ": численность отряда " + strconv.Itoa(count)
	}
	return missionsP
}

func GetSystemPrompt(props Props) string {
	return "Вы играете в настольную игру Авалон Resistance\n" +
		fmt.Sprintf("В игре %d игроков за круглым столом \n", len(props.Players)) +
		fmt.Sprintf("Вот их имена в порядке обхода стола: %s\n", strings.Join(props.Players, ", ")) +
		fmt.Sprintf("Вы отыгрываете персонажа - %s\n", props.Self) +
		fmt.Sprintf("Ваша манера речи - %s\n", props.Mood) +
		fmt.Sprintf("Ваша любовь к риску - %s\n", props.Risk) +
		fmt.Sprintf("Каждый из игроков имеет 1 из ролей%s\n", formatRoles(props.Roles)) +
		fmt.Sprintf("Ваша роль - %s\n", props.Role) +
		props.RoleContext + "\n" +
		"Вы должны добиться победы своей команды\n" +
		fmt.Sprintf("Игра содержит %d Миссий: %s\n", len(props.Missions), formatMissions(props.Missions)) +
		"В начале игры случайным образом выбирается лидер. \n" +
		"Лидер предлагает свой состав игроков на первую непосещенную миссию, размер команды строго регламентирован миссией.\n" +
		"Дальше все игроки по порядку, говорят поддержали ли бы они такой состав или предлагают альтернативный\n" +
		"После этого лидер выдвигает финальный состав на миссию\n" +
		"А игроки голосуют по очереди в открытую ЗА или ПРОТИВ они такого состава\n" +
		"Если голосов ПРОТИВ больше, то лидерство передается следующему игроку по порядку и цикл повторяется снова\n" +
		"Иначе состав утверждается на миссию и каждый игрок из состава голосует УСПЕХ или ПРОВАЛ в закрытую, общее число УСПЕХОВ и ПРОВАЛОВ оглашается всем игрокам\n" +
		"Миссия считается проваленной, если есть хотя бы 1 ПРОВАЛ (если иначе не описано в миссии), иначе она успешна. Успешная или проваленная миссия считается посещенной и больше на нее нельзя собирать состав\n" +
		"После посещения миссии, лидерство передается следующему игроку по порядку\n" +
		"Если 5 лидеров подряд не смогли собрать состав,  то игра заканчивается победой команды 'Красные' и поражением команды 'Синие'\n" +
		"После провала 3х миссий, игра заканчивается победой команды 'Красные' и поражением команды 'Синие'\n" +
		"После успеха 3х миссий, происходит событие Покушение\n" +
		"В рамках события покушения игрок с ролью Асасин, должен ответить какой из игроков имеет роль Мерлин. Если он отвечает правильно, то игра заканчивается победой команды 'Красные' и поражением команды 'Синие',\n" +
		"Иначе игра заканчивается победой команды 'Синие' и поражением команды 'Красные'"
}
