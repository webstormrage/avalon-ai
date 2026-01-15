package logging

import (
	"avalon/pkg/dto"
	"bufio"
	"fmt"
	"os"
)

func Console(player *dto.Character, message string) {
	fmt.Printf("[%s](%s): %s\n", player.Persona.Self, player.Persona.Role, message)
}

func Info(message string) {
	fmt.Printf("<Игровое Событие>: %s\n", message)
}

func Action(state *dto.GameState, player *dto.Character, message string) {
	state.Logs = append(state.Logs, dto.Action{User: player.Persona.Self, Message: message})
}

func Event(state *dto.GameState, message string) {
	state.Logs = append(state.Logs, dto.Action{User: dto.System, Message: message})
}

func WaitForMasterInput() {
	reader := bufio.NewReader(os.Stdin)
	_, _, err := reader.ReadRune()
	if err != nil {
		panic(err)
	}
}
