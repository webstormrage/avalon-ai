package domain

import (
	"avalon/pkg/gemini"
	"fmt"
	"strings"
)

func FmtPlayers(players []*gemini.Character) string {
	names := []string{}
	for _, player := range players {
		names = append(names, player.Persona.Self)
	}
	return strings.Join(names, ", ")
}

func FmtVotes(leader string, team string, votes map[string]string) string {
	items := []string{}
	for k, v := range votes {
		items = append(items, fmt.Sprintf("%s: %s", k, v))
	}
	return fmt.Sprintf("Результаты голосования (Лидер: %s, Команда: %s):\n %s\n", leader, team, strings.Join(items, "\n"))
}
