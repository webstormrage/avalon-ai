package prompts

import (
	"bytes"
	"strings"
	"text/template"
)

const proposeSquadPromptTpl = `
Вы лидер.
{{template "resumePrompt" .Resume}}

Вы должны предложить состав на {{.Mission.Name}},
состав из {{.Mission.SquadSize}} любых игроков,
допустимое количество провалов {{.Mission.MaxFails}}

Вашу речь услышат другие игроки, учитывайте что ваши цели могут не совпадать.

Последнее предложение речи должно быть в следующем формате:
Выставить: имена игроков через запятую
`

func RenderProposeSquadPrompt(view StatementProps) string {
	tpl, err := template.New("proposeSquadPrompt").Parse(`
{{define "resumePrompt"}}` + resumePromptTpl + `{{end}}
{{define "proposeSquadPrompt"}}` + proposeSquadPromptTpl + `{{end}}
`)
	tpl = template.Must(tpl, err)

	var buf bytes.Buffer
	if err := tpl.ExecuteTemplate(&buf, "proposeSquadPrompt", view); err != nil {
		panic(err)
	}

	return strings.TrimSpace(buf.String()) + "\n"
}
