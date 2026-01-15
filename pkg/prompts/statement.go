package prompts

import (
	"bytes"
	"strings"
	"text/template"
)

type StatementProps struct {
	Resume  ResumeProps
	Mission MissionProps
}

const statementPromptTpl = `
Вы лидер.
{{template "resumePrompt" .Resume}}

Вы должны выставить на голосование состав на Миссию — {{add .Mission.Index 1}},
состав из {{.Mission.Size}} любых игроков.

Последнее предложение речи должно быть в следующем формате:
Выставить: имена игроков через запятую
`

func RenderStatementPrompt(view StatementProps) string {
	tpl := template.Must(
		template.New("statementPrompt").
			Funcs(template.FuncMap{
				"add": func(a, b int) int { return a + b },
			}).
			Parse(`
{{define "resumePrompt"}}` + resumePromptTpl + `{{end}}
{{define "statementPrompt"}}` + statementPromptTpl + `{{end}}
`),
	)

	var buf bytes.Buffer
	if err := tpl.ExecuteTemplate(&buf, "leader", view); err != nil {
		panic(err)
	}

	return strings.TrimSpace(buf.String()) + "\n"
}
