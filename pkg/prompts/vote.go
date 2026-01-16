package prompts

import (
	"avalon/pkg/dto"
	"bytes"
	"regexp"
	"strings"
	"text/template"
)

type VoteProps struct {
	Mission dto.MissionV2
	Team    string
	Leader  string
}

const votePromptTpl = `
Лидер ({{.Leader}}) выставил свой финальный состав на Миссию — {{add .Mission.Index 1}} (участников: {{.Mission.Size}})
Предлагаемый состав - {{.Team}}

Вы можете проголосовать либо против, либо за.

Все игроки голосуют одновременно.

Другие игроки не видят вашу речь, но после голосования будут знать за что вы проголосовали. 

Последнее предложение в вашей речи должно быть либо Голосовать: ПРОТИВ, либо Голосовать: ЗА
`

func ExtractVote(text string) bool {
	re := regexp.MustCompile(`(?i)Голосовать:\s*(ЗА|ПРОТИВ)\s*$`)
	m := re.FindStringSubmatch(text)

	if len(m) < 2 {
		return false
	}

	// нормализуем результат
	switch strings.ToUpper(m[1]) {
	case "ЗА":
		return true
	case "ПРОТИВ":
		return false
	default:
		return false
	}
}

func RenderVotePrompt(view VoteProps) string {
	tpl := template.Must(
		template.New("votePrompt").
			Funcs(template.FuncMap{
				"add": func(a, b int) int { return a + b },
			}).
			Parse(votePromptTpl),
	)

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, view); err != nil {
		panic(err)
	}

	return strings.TrimSpace(buf.String()) + "\n"
}
