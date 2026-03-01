package prompts

import (
	"bytes"
	"strings"
	"text/template"
)

type ResumeProps struct {
	Fails      int // количество проваленных миссий
	Wins       int // количество успешных миссий
	SkipsCount int // сколько лидеров подряд не смогли собрать состав
}

const resumePromptTpl = `
На данный момент провалено {{.Fails}} миссий, выполнено — {{.Wins}} миссий.
{{.SkipsCount}} лидеров подряд не смогли собрать состав на миссию.
`

func RenderResumePrompt(state ResumeProps) string {
	tpl := template.Must(
		template.New("resumePrompt").
			Parse(resumePromptTpl),
	)

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, state); err != nil {
		panic(err)
	}

	return strings.TrimSpace(buf.String()) + "\n"
}
