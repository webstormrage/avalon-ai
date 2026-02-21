package prompts

import (
	"avalon/pkg/constants"
	"avalon/pkg/dto"
	"avalon/pkg/presets"
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"text/template"
)

type SystemPromptProps struct {
	Name        string
	Players     []dto.PlayerV2
	Roles       []presets.Role
	Role        string
	RoleContext string
	Missions    []dto.MissionV2
}

func formatRoles(roles []presets.Role) string {
	rolesP := ""
	for _, role := range roles {
		rolesP = rolesP + "\n" + role.Name + " (число игроков с ролью - " + strconv.Itoa(role.Count) + ")\n" + role.Description
	}
	return rolesP
}

func formatMissions(missions []dto.MissionV2) string {
	missionsP := ""
	for _, mission := range missions {
		missionsP += fmt.Sprintf("%s: численность отряда %d, допустимое число провалов: %d\n", mission.Name, mission.SquadSize, mission.MaxFails)
	}
	return missionsP
}

func formatPlayers(players []dto.PlayerV2) string {
	playersP := ""
	for _, player := range players {
		playersP += player.Name + " "
	}
	return playersP
}

func formatRed(players []dto.PlayerV2, role string) string {
	if role != constants.ROLE_ASSASSIN && role != constants.ROLE_MERLIN && role != constants.ROLE_MORDRED_MINION {
		return ""
	}
	redPlayers := "Вам известно что следующие игроки принадлежат команде 'Красные':"
	for _, player := range players {
		if player.Role == constants.ROLE_ASSASSIN || player.Role == constants.ROLE_MORDRED_MINION {
			redPlayers += " " + player.Name
		}
	}
	return redPlayers
}

const systemPromptTpl = `
Вы играете в настольную игру Авалон Resistance
В игре {{len .Players}} игроков за круглым столом
Вот их имена в порядке обхода стола: {{formatPlayers .Players}}
Вы отыгрываете персонажа - {{.Name}}
Каждый из игроков имеет 1 из ролей {{formatRoles .Roles}}
Ваша роль - {{.Role}}
{{formatRed .Players .Role}}
Вы должны добиться победы своей команды
Игра содержит {{len .Missions}} Миссий: {{formatMissions .Missions}}

В начале игры случайным образом выбирается лидер.
Лидер предлагает свой состав игроков на первую непосещенную миссию, размер команды строго регламентирован миссией.
Дальше все игроки по порядку, говорят поддержали ли бы они такой состав или предлагают альтернативный.
После этого лидер выдвигает финальный состав на миссию.
А игроки голосуют одновременно в открытую ЗА или ПРОТИВ они такого состава.
Если голосов ПРОТИВ больше, то лидерство передается следующему игроку по порядку и цикл повторяется снова.
Иначе состав утверждается на миссию и каждый игрок из состава голосует УСПЕХ или ПРОВАЛ в закрытую, общее число УСПЕХОВ и ПРОВАЛОВ оглашается всем игрокам.
В описании каждой миссии указано допустимое максимальное число провалов, чтобы она считалась успешной.
Успешная или проваленная миссия считается посещенной и больше на нее нельзя собирать состав.
После посещения миссии, лидерство передается следующему игроку по порядку.
Если 5 лидеров подряд не смогли собрать состав, то игра заканчивается победой команды 'Красные' и поражением команды 'Синие'.
После провала 3х миссий, игра заканчивается победой команды 'Красные' и поражением команды 'Синие'.
После успеха 3х миссий, происходит событие Покушение.
В рамках события покушения игрок с ролью Асасин должен ответить, какой из игроков имеет роль Мерлин.
Если он отвечает правильно — побеждает команда 'Красные', иначе побеждает команда 'Синие'.
Учитывайте что не все игроки обладают одинаковой осведомленностью о ролях других игроков, а также то что ваши цели не всегда совпадают.
`

func GetSystemPrompt(props SystemPromptProps) string {
	tpl := template.Must(
		template.New("systemPrompt").
			Funcs(template.FuncMap{
				"formatPlayers":  formatPlayers,
				"formatRoles":    formatRoles,
				"formatMissions": formatMissions,
				"formatRed":      formatRed,
			}).
			Parse(systemPromptTpl),
	)

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, props); err != nil {
		panic(err)
	}

	return strings.TrimSpace(buf.String())
}
