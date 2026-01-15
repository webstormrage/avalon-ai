package stages

import (
	"avalon/pkg/constants"
	"avalon/pkg/dto"
	"avalon/pkg/logging"
	"avalon/pkg/prompts"
	"avalon/pkg/selectors"
	"strings"
)

func AssassinStage(state *dto.GameState) {
	assassin := selectors.GetPlayerByRole(state, constants.ROLE_ASSASSIN)
	message, err := assassin.Send(
		prompts.AssassinPrompt,
		state.Logs,
	)
	if err != nil {
		panic(err)
	}
	logging.Console(assassin, message)
	logging.WaitForMasterInput()
	logging.Action(state, assassin, message)
	targetName, _ := prompts.ExtractPlayerName(message)
	merlin := selectors.GetPlayerByRole(state, constants.ROLE_MERLIN)
	if strings.ToLower(merlin.Persona.Self) == strings.ToLower(targetName) {
		gameMessage := "Игра окончена. Ассассин убил Мерлина. Победила команда 'Красные'"
		logging.Info(gameMessage)
		logging.Event(state, gameMessage)
		return
	}
	gameMessage := "Игра окончена. Ассассин не нашел Мерлина. Победила команда 'Синие'"
	logging.Info(gameMessage)
	logging.Event(state, gameMessage)
}
