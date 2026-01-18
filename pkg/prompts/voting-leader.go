package prompts

import (
	"avalon/pkg/dto"
	"bytes"
	"regexp"
	"strings"
	"text/template"
)

type StatementProps struct {
	Resume  ResumeProps
	Mission dto.MissionV2
}

const statementPromptTpl = `
Вы лидер.
{{template "resumePrompt" .Resume}}

Вы должны выставить на голосование состав на {{.Mission.Name}},
состав из {{.Mission.SquadSize}} любых игроков,
допустимое количество провалов {{.Mission.MaxFails}}

Вашу речь услышат другие игроки, учитывайте что ваши цели могут не совпадать.

Последнее предложение речи должно быть в следующем формате:
Выставить: имена игроков через запятую
`

func ExtractTeam(text string) ([]string, bool) {
	re := regexp.MustCompile(`(?i)Выставить:\s*(.+)\s*$`)
	m := re.FindStringSubmatch(text)

	if len(m) < 2 {
		return nil, false
	}

	raw := m[1]
	parts := strings.Split(raw, ",")

	var players []string
	for _, p := range parts {
		name := strings.TrimSpace(p)
		if name != "" {
			players = append(players, name)
		}
	}

	if len(players) == 0 {
		return nil, false
	}

	return players, true
}

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
	if err := tpl.ExecuteTemplate(&buf, "statementPrompt", view); err != nil {
		panic(err)
	}

	return strings.TrimSpace(buf.String()) + "\n"
}
