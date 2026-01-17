package prompts

import (
	"bytes"
	"regexp"
	"strings"
	"text/template"
)

const completionPromptTpl = `
Лидер ({{.Leader}}) выставил вас в свой финальный состав на {{.Mission.Name}},
(допустимое количество провалов {{.Mission.MaxFails}})
Предлагаемый состав - {{.Team}}

Вы сейчас находитесь на Миссии. Вы можете либо успещно выполнить свою часть, либо провалить.
Остальные игроки не узнают, что именно вы выбрали и не увидят вашу речь.
Но они будут знать число провалов и успехов в этой миссии.
Для провала миссии достаточно {{add .Mission.MaxFails 1}}) провала, иначе миссия будет считаться успешной.

Последнее предложение в вашей речи должно быть либо Совершить: УСПЕХ или Совершить: ПРОВАЛ
`

func ExtractMissionResult(text string) string {
	re := regexp.MustCompile(`(?i)Совершить:\s*(УСПЕХ|ПРОВАЛ)\s*$`)
	m := re.FindStringSubmatch(text)

	if len(m) < 2 {
		return "УСПЕХ"
	}

	switch strings.ToUpper(m[1]) {
	case "УСПЕХ":
		return "УСПЕХ"
	case "ПРОВАЛ":
		return "ПРОВАЛ"
	default:
		return "УСПЕХ"
	}
}

func RenderCompletionPrompt(view VoteProps) string {
	tpl := template.Must(
		template.New("completionPrompt").
			Funcs(template.FuncMap{
				"add": func(a, b int) int { return a + b },
			}).
			Parse(completionPromptTpl),
	)

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, view); err != nil {
		panic(err)
	}

	return strings.TrimSpace(buf.String()) + "\n"
}
