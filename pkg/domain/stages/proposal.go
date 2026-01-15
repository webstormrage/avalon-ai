package stages

import (
	"avalon/pkg/dto"
	"avalon/pkg/logging"
	"avalon/pkg/prompts"
	"avalon/pkg/selectors"
)

func RunProposal(state *dto.GameState) string {
	leader := selectors.GetLeader(state)
	mission := selectors.GetMission(state)
	message, err := leader.Send(prompts.RenderProposalPrompt(prompts.StatementProps{
		Resume: prompts.ResumeProps{
			Wins:       state.Wins,
			Fails:      state.Fails,
			SkipsCount: state.SkipsCount,
		},
		Mission: mission,
	}), state.Logs)
	if err != nil {
		panic(err)
	}
	logging.Console(leader, message)
	logging.WaitForMasterInput()
	logging.Action(state, leader, message)
	return logging.FmtPlayers(prompts.ExtractCharacters(message, state.Players))
}
