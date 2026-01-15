package prompts

import (
	"bytes"
	"strings"
	"text/template"
)

const completionPromptTpl = `
Лидер ({{.Leader}}) выставил вас в свой финальный состав на Миссию — {{add .Mission.Index 1}} (участников: {{.Mission.Size}})
Полный состав - {{.Team}}

Вы сейчас находитесь на Миссии. Вы можете либо успещно выполнить свою часть, либо провалить.
Остальные игроки не узнают, что именно вы выбрали и не увидят вашу речь.
Но они будут знать число провалов и успехов в этой миссии.
Для провала миссии достаточно 1 провала, а для успеха требуется, чтобы все участники миссии совершили успех.

Последнее предложение в вашей речи должно быть либо Совершить: УСПЕХ или Совершить: ПРОВАЛ
`

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
