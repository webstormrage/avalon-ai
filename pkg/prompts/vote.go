package prompts

import (
	"bytes"
	"strings"
	"text/template"
)

type VoteProps struct {
	Mission MissionProps
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
