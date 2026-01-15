package domain

import (
	"avalon/pkg/domain/stages"
	"avalon/pkg/dto"
)

func RunCircle(state *dto.GameState) {
	leaderProposal := stages.RunProposal(state)

	for i := 1; i < len(state.Players); i++ {
		stages.RunComment(state, i, leaderProposal)
	}

	leaderTeam := stages.RunStatement(state)

	votesForLeader, votesAgainstLeader := stages.RunVoting(state, leaderTeam)

	if votesForLeader < votesAgainstLeader {
		stages.RunSkipLeader(state)
		return
	}

	stages.RunMission(state, leaderTeam)
}

func RunGame(state *dto.GameState) {

	for state.Wins < 3 && state.Fails < 3 && state.SkipsCount < 5 {
		RunCircle(state)
	}

	if state.SkipsCount == 5 {
		stages.TooManySkipsFinal(state)
	} else if state.Fails == 3 {
		stages.TooManyFailsFinal(state)
	} else if state.Wins == 3 {
		stages.AssassinStage(state)
	}
}
