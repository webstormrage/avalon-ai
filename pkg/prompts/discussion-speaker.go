package prompts

import (
	"bytes"
	"strings"
	"text/template"
)

const commentPromptTpl = `
Лидер ({{.Leader}}) предложил состав на {{.Mission.Name}},
(допустимое количество провалов {{.Mission.MaxFails}})
Предлагаемый состав - {{.Team}}

Вы можете поддержать состав лидера, или любой другой озвученный ранее,
либо предложить альтернативный состав.

Вашу речь услышат другие игроки, учитывайте что ваши цели могут не совпадать.

В любом случае последнее предложение речи должно быть в следующем формате:
Поддерживаю: имена игроков через запятую
`

func RenderCommentPrompt(view VoteProps) string {
	tpl := template.Must(
		template.New("commentPrompt").
			Funcs(template.FuncMap{
				"add": func(a, b int) int { return a + b },
			}).
			Parse(commentPromptTpl),
	)

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, view); err != nil {
		panic(err)
	}

	return strings.TrimSpace(buf.String()) + "\n"
}
