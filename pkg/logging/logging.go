package logging

import (
	"avalon/pkg/action"
	"avalon/pkg/dto"
	"avalon/pkg/gemini"
	"bufio"
	"fmt"
	"os"
)

func Console(player *gemini.Character, message string) {
	fmt.Printf("[%s](%s): %s\n", player.Persona.Self, player.Persona.Role, message)
}

func Info(message string) {
	fmt.Printf("<Игровое Событие>: %s\n", message)
}

func Action(state *dto.GameState, player *gemini.Character, message string) {
	state.Logs = append(state.Logs, action.Action{User: player.Persona.Self, Message: message})
}

func Event(state *dto.GameState, message string) {
	state.Logs = append(state.Logs, action.Action{User: action.System, Message: message})
}

func WaitForMasterInput() {
	reader := bufio.NewReader(os.Stdin)
	_, _, err := reader.ReadRune()
	if err != nil {
		panic(err)
	}
}
