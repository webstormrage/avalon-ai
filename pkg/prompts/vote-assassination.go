package prompts

import (
	"bytes"
	"regexp"
	"strings"
	"text/template"
)

type VoteAssassinationProps struct {
	Target string
}

const voteAssassinationPromptTpl = `
Обсуждение завершено.
Текущая предлагаемая цель: {{.Target}}.

Назовите игрока, которого ассасин должен выбрать.

Последнее предложение вашей речи должно быть в формате:
Выбрать: имя игрока
`

func ExtractAssassinationVote(text string) (string, bool) {
	re := regexp.MustCompile(`(?i)Выбрать:\s*(.+)\s*$`)
	m := re.FindStringSubmatch(text)
	if len(m) < 2 {
		return "", false
	}
	name := strings.TrimSpace(m[1])
	if name == "" {
		return "", false
	}
	return name, true
}

func RenderVoteAssassinationPrompt(view VoteAssassinationProps) string {
	tpl := template.Must(template.New("voteAssassinationPrompt").Parse(voteAssassinationPromptTpl))

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, view); err != nil {
		panic(err)
	}

	return strings.TrimSpace(buf.String()) + "\n"
}
