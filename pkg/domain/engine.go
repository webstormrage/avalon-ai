package domain

import (
	"avalon/pkg/constants"
	"avalon/pkg/dto"
	"avalon/pkg/gemini"
	"avalon/pkg/logging"
	"avalon/pkg/prompts"
	"fmt"
	"slices"
	"strings"
)

func AssassinStage(state *dto.GameState) {
	idx := slices.IndexFunc(state.Players, func(p *gemini.Character) bool {
		return p.Persona.Role == constants.ROLE_ASSASSIN
	})
	assassin := state.Players[idx]
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
	idx = slices.IndexFunc(state.Players, func(p *gemini.Character) bool {
		return strings.ToLower(targetName) == strings.ToLower(p.Persona.Self)
	})
	target := state.Players[idx]
	if target.Persona.Role == constants.ROLE_MERLIN {
		gameMessage := "Игра окончена. Ассассин убил Мерлина. Победила команда 'Красные'"
		logging.Info(gameMessage)
		logging.Event(state, gameMessage)
		return
	}
	gameMessage := "Игра окончена. Ассассин не нашел Мерлина. Победила команда 'Синие'"
	logging.Info(gameMessage)
	logging.Event(state, gameMessage)
}

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

func GetLeader(state *dto.GameState) *gemini.Character {
	return state.Players[state.LeaderIndex]
}

func GetMission(state *dto.GameState) prompts.MissionProps {
	return prompts.MissionProps{
		Index: state.MissionIndex,
		Size:  state.Missions[state.MissionIndex],
	}
}

func RunProposal(state *dto.GameState) string {
	leader := GetLeader(state)
	mission := GetMission(state)
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
	return FmtPlayers(prompts.ExtractCharacters(message, state.Players))
}

func RunComment(state *dto.GameState, index int, leaderProposal string) {
	leader := GetLeader(state)
	mission := GetMission(state)
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

func RunStatement(state *dto.GameState) []*gemini.Character {
	leader := GetLeader(state)
	mission := GetMission(state)
	message, err := leader.Send(prompts.RenderStatementPrompt(
		prompts.StatementProps{
			Resume: prompts.ResumeProps{
				Wins:       state.Wins,
				Fails:      state.Fails,
				SkipsCount: state.SkipsCount,
			},
			Mission: mission,
		},
	), state.Logs)
	if err != nil {
		panic(err)
	}
	logging.Console(leader, message)
	logging.WaitForMasterInput()
	logging.Action(state, leader, message)
	return prompts.ExtractCharacters(message, state.Players)
}

func RunVoting(state *dto.GameState, leaderTeam []*gemini.Character) (votesForLeader int, votesAgainstLeader int) {
	leader := GetLeader(state)
	mission := GetMission(state)
	votes := map[string]string{}
	votesForLeader = 0
	votesAgainstLeader = 0

	for i := 1; i < len(state.Players); i++ {
		player := state.Players[(state.LeaderIndex+i)%len(state.Players)]

		message, err := player.Send(prompts.RenderVotePrompt(
			prompts.VoteProps{
				Leader:  leader.Persona.Self,
				Team:    FmtPlayers(leaderTeam),
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
	votingMessage := FmtVotes(leader.Persona.Self, FmtPlayers(leaderTeam), votes)
	logging.Info(votingMessage)
	logging.Event(state, votingMessage)
	return votesForLeader, votesAgainstLeader
}

func RunSkipLeader(state *dto.GameState) {
	leader := GetLeader(state)
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

func RunMission(state *dto.GameState, leaderTeam []*gemini.Character) {
	leader := GetLeader(state)
	mission := GetMission(state)
	missionMessage := fmt.Sprintf(
		"Лидер %s смог собрать состав на Миссию %d - %s\n",
		leader.Persona.Self,
		state.MissionIndex+1,
		FmtPlayers(leaderTeam),
	)
	logging.Info(missionMessage)
	logging.Event(state, missionMessage)

	votesFail := 0
	votesSuccess := 0

	for _, player := range leaderTeam {
		message, err := player.Send(prompts.RenderCompletionPrompt(
			prompts.VoteProps{
				Leader:  leader.Persona.Self,
				Team:    FmtPlayers(leaderTeam),
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

func RunCircle(state *dto.GameState) {
	leaderProposal := RunProposal(state)

	for i := 1; i < len(state.Players); i++ {
		RunComment(state, i, leaderProposal)
	}

	leaderTeam := RunStatement(state)

	votesForLeader, votesAgainstLeader := RunVoting(state, leaderTeam)

	if votesForLeader < votesAgainstLeader {
		RunSkipLeader(state)
		return
	}

	RunMission(state, leaderTeam)
}

func RunGame(state *dto.GameState) {

	for state.Wins < 3 && state.Fails < 3 && state.SkipsCount < 5 {
		RunCircle(state)
	}

	if state.SkipsCount == 5 {
		TooManySkipsFinal(state)
	} else if state.Fails == 3 {
		TooManyFailsFinal(state)
	} else if state.Wins == 3 {
		AssassinStage(state)
	}
}
