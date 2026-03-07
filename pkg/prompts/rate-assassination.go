package prompts

import (
	"bytes"
	"strings"
	"text/template"
)

type RateAssassinationProps struct {
	Speaker string
	Target  string
}

const rateAssassinationPromptTpl = `
Игрок {{.Speaker}} предложил цель для убийства: {{.Target}}.

Остальные красные игроки обсуждают предложение.
Ваша речь публичная, ее увидят все игроки.

Вы можете поддержать предложенную цель или предложить другую.

Последнее предложение вашей речи должно быть в формате:
Поддерживаю: имя игрока
`

func RenderRateAssassinationPrompt(view RateAssassinationProps) string {
	tpl := template.Must(template.New("rateAssassinationPrompt").Parse(rateAssassinationPromptTpl))

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, view); err != nil {
		panic(err)
	}

	return strings.TrimSpace(buf.String()) + "\n"
}
