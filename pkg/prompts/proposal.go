package prompts

import (
	"bytes"
	"strings"
	"text/template"
)

const proposalPromptTpl = `
Вы лидер.
{{template "resumePrompt" .Resume}}

Вы должны предложить состав на {{.Mission.Name},
состав из {{.Mission.SquadSize}} любых игроков.

Вашу речь услышат другие игроки, учитывайте что ваши цели могут не совпадать.

Последнее предложение речи должно быть в следующем формате:
Выставить: имена игроков через запятую
`

func RenderProposalPrompt(view StatementProps) string {
	tpl := template.Must(
		template.New("proposalPrompt").Parse(`
{{define "resumePrompt"}}` + resumePromptTpl + `{{end}}
{{define "proposalPrompt"}}` + proposalPromptTpl + `{{end}}
`),
	)

	var buf bytes.Buffer
	if err := tpl.ExecuteTemplate(&buf, "proposalPrompt", view); err != nil {
		panic(err)
	}

	return strings.TrimSpace(buf.String()) + "\n"
}
