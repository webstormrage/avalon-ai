package stages

import (
	"avalon/pkg/dto"
	"avalon/pkg/logging"
	"avalon/pkg/prompts"
	"avalon/pkg/selectors"
)

func RunComment(state *dto.GameState, index int, leaderProposal string) {
	leader := selectors.GetLeader(state)
	mission := selectors.GetMission(state)
	player := state.Players[(state.LeaderIndex+index)%len(state.Players)]

	message, err := player.Send(prompts.RenderCommentPrompt(prompts.VoteProps{
		Leader:  leader.Persona.Self,
		Team:    leaderProposal,
		Mission: mission,
	}), state.Logs)
	if err != nil {
		panic(err)
	}
	logging.Console(player, message)
	logging.WaitForMasterInput()
	logging.Action(state, player, message)
}
