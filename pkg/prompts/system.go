package prompts

import (
	"avalon/pkg/presets"
	"bytes"
	"strconv"
	"strings"
	"text/template"
)

type Props struct {
	Self        string
	Mood        string
	Risk        string
	Players     []string
	Roles       []presets.Role
	Role        string
	RoleContext string
	Missions    []int
}

func formatRoles(roles []presets.Role) string {
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

const systemPromptTpl = `
Вы играете в настольную игру Авалон Resistance
В игре {{len .Players}} игроков за круглым столом
Вот их имена в порядке обхода стола: {{join .Players ", "}}
Вы отыгрываете персонажа - {{.Self}}
Ваша манера речи - {{.Mood}}
Ваша любовь к риску - {{.Risk}}
Каждый из игроков имеет 1 из ролей {{formatRoles .Roles}}
Ваша роль - {{.Role}}
{{.RoleContext}}
Вы должны добиться победы своей команды
Игра содержит {{len .Missions}} Миссий: {{formatMissions .Missions}}

В начале игры случайным образом выбирается лидер.
Лидер предлагает свой состав игроков на первую непосещенную миссию, размер команды строго регламентирован миссией.
Дальше все игроки по порядку, говорят поддержали ли бы они такой состав или предлагают альтернативный.
После этого лидер выдвигает финальный состав на миссию.
А игроки голосуют одновременно в открытую ЗА или ПРОТИВ они такого состава.
Если голосов ПРОТИВ больше, то лидерство передается следующему игроку по порядку и цикл повторяется снова.
Иначе состав утверждается на миссию и каждый игрок из состава голосует УСПЕХ или ПРОВАЛ в закрытую, общее число УСПЕХОВ и ПРОВАЛОВ оглашается всем игрокам.
Миссия считается проваленной, если есть хотя бы 1 ПРОВАЛ (если иначе не описано в миссии), иначе она успешна.
Успешная или проваленная миссия считается посещенной и больше на нее нельзя собирать состав.
После посещения миссии, лидерство передается следующему игроку по порядку.
Если 5 лидеров подряд не смогли собрать состав, то игра заканчивается победой команды 'Красные' и поражением команды 'Синие'.
После провала 3х миссий, игра заканчивается победой команды 'Красные' и поражением команды 'Синие'.
После успеха 3х миссий, происходит событие Покушение.
В рамках события покушения игрок с ролью Асасин должен ответить, какой из игроков имеет роль Мерлин.
Если он отвечает правильно — побеждает команда 'Красные', иначе побеждает команда 'Синие'.
Учитывайте что не все игроки обладают одинаковой осведомленностью о ролях других игроков, а также то что ваши цели не всегда совпадают.
`

func GetSystemPrompt(props Props) string {
	tpl := template.Must(
		template.New("systemPrompt").
			Funcs(template.FuncMap{
				"join":           strings.Join,
				"formatRoles":    formatRoles,
				"formatMissions": formatMissions,
			}).
			Parse(systemPromptTpl),
	)

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, props); err != nil {
		panic(err)
	}

	return strings.TrimSpace(buf.String())
}
