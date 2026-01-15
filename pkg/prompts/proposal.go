package prompts

import (
	"bytes"
	"strings"
	"text/template"
)

type MissionProps struct {
	Index int // текущий индекс миссии (0-based)
	Size  int // размер команды
}

const proposalPromptTpl = `
Вы лидер.
{{template "resumePrompt" .Resume}}

Вы должны предложить состав на Миссию — {{add .Mission.Index 1}},
состав из {{.Mission.Size}} любых игроков.

Последнее предложение речи должно быть в следующем формате:
Выставить: имена игроков через запятую
`

func RenderProposalPrompt(view StatementProps) string {
	tpl := template.Must(
		template.New("proposalPrompt").
			Funcs(template.FuncMap{
				"add": func(a, b int) int { return a + b },
			}).
			Parse(`
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
