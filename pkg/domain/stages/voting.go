package stages

import (
	"avalon/pkg/dto"
	"avalon/pkg/logging"
	"avalon/pkg/prompts"
	"avalon/pkg/selectors"
)

func RunVoting(state *dto.GameState, leaderTeam []*dto.Character) (votesForLeader int, votesAgainstLeader int) {
	leader := selectors.GetLeader(state)
	mission := selectors.GetMission(state)
	votes := map[string]string{}
	votesForLeader = 0
	votesAgainstLeader = 0

	for i := 1; i < len(state.Players); i++ {
		player := state.Players[(state.LeaderIndex+i)%len(state.Players)]

		message, err := player.Send(prompts.RenderVotePrompt(
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
		if prompts.ExtractVote(message) {
			votesForLeader++
			votes[player.Persona.Self] = "ЗА"
		} else {
			votesAgainstLeader++
			votes[player.Persona.Self] = "ПРОТИВ"
		}
	}
	votingMessage := logging.FmtVotes(leader.Persona.Self, logging.FmtPlayers(leaderTeam), votes)
	logging.Info(votingMessage)
	logging.Event(state, votingMessage)
	return votesForLeader, votesAgainstLeader
}
