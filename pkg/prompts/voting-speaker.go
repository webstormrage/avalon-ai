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
Лидер ({{.Leader}}) предложил состав на {{.Mission.Name}},
(допустимое количество провалов {{.Mission.MaxFails}})
Предлагаемый состав - {{.Team}}

Вы можете проголосовать либо против, либо за.

Все игроки голосуют одновременно.

Другие игроки не видят вашу речь, но после голосования будут знать за что вы проголосовали. 

Единственное предложение в вашей речи должно быть либо Голосовать: ПРОТИВ, либо Голосовать: ЗА
`

func ExtractVote(text string) string {
	re := regexp.MustCompile(`(?i)Голосовать:\s*(ЗА|ПРОТИВ)\s*$`)
	m := re.FindStringSubmatch(text)

	if len(m) < 2 {
		return ""
	}

	// нормализуем результат
	switch strings.ToUpper(m[1]) {
	case "ЗА":
		return "ЗА"
	case "ПРОТИВ":
		return "ПРОТИВ"
	default:
		return ""
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
