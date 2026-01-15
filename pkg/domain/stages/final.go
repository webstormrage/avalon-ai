package stages

import (
	"avalon/pkg/dto"
	"avalon/pkg/logging"
)

func TooManySkipsFinal(state *dto.GameState) {
	gameMessage := "Игра окончена. 5 лидеров подряд не смогли собрать состав на миссию. Победила команда 'Красные'"
	logging.Info(gameMessage)
	logging.Event(state, gameMessage)
}

func TooManyFailsFinal(state *dto.GameState) {
	gameMessage := "Игра окончена. 3 миссии были провалены. Победила команда 'Красные'"
	logging.Info(gameMessage)
	logging.Event(state, gameMessage)
}
