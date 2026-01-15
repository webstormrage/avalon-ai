package stages

import (
	"avalon/pkg/dto"
	"avalon/pkg/logging"
	"avalon/pkg/prompts"
	"avalon/pkg/selectors"
	"fmt"
)

func RunSkipLeader(state *dto.GameState) {
	leader := selectors.GetLeader(state)
	state.SkipsCount++
	state.LeaderIndex = (state.LeaderIndex + 1) % len(state.Players)
	missionMessage := fmt.Sprintf(
		"Лидер %s не смог собрать состав на Миссию %d\n",
		leader.Persona.Self,
		state.MissionIndex+1,
	)
	logging.Info(missionMessage)
	logging.Event(state, missionMessage)
}

func RunMission(state *dto.GameState, leaderTeam []*dto.Character) {
	leader := selectors.GetLeader(state)
	mission := selectors.GetMission(state)
	missionMessage := fmt.Sprintf(
		"Лидер %s смог собрать состав на Миссию %d - %s\n",
		leader.Persona.Self,
		state.MissionIndex+1,
		logging.FmtPlayers(leaderTeam),
	)
	logging.Info(missionMessage)
	logging.Event(state, missionMessage)

	votesFail := 0
	votesSuccess := 0

	for _, player := range leaderTeam {
		message, err := player.Send(prompts.RenderCompletionPrompt(
			prompts.VoteProps{
				Leader:  leader.Persona.Self,
				Team:    logging.FmtPlayers(leaderTeam),
				Mission: mission,
			},
		), state.Logs)
		if err != nil {
			panic(err)
		}
		logging.Console(player, message)
		logging.WaitForMasterInput()
		if prompts.ExtractMissionResult(message) {
			votesSuccess++
		} else {
			votesFail++
		}
	}
	result := "Миссия выполнена"
	if votesFail > 0 {
		result = "Миссия провалена"
		state.Fails++
	} else {
		state.Wins++
	}
	missionMessage = fmt.Sprintf(
		"Результат миссии: %d\nУспехов: %d\nПровалов: %d\n%s\n",
		state.MissionIndex+1,
		votesSuccess,
		votesFail,
		result,
	)
	logging.Info(missionMessage)
	logging.Event(state, missionMessage)
	state.SkipsCount = 0
	state.MissionIndex++
	state.LeaderIndex = (state.LeaderIndex + 1) % len(state.Players)
}
