package prompts

import (
	"regexp"
	"strings"
)

const AssassinPrompt = `
Команда 'Синие' близка к победе. Они выполнили успешно 3 миссии. Но у 'Красных есть шанс победить
Назовите имя игрока за столом, который по вашему мнению имеет роль 'Мерлин'
Если вы угадаете, то победа будет за 'Красными'
Последнее предложение вашей речи должно быть Выбрать: имя игрока
`

func ExtractPlayerName(text string) (string, bool) {
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
