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

func RunGame(state *dto.GameState) {
	for state.Wins < 3 && state.Fails < 3 && state.SkipsCount < 5 {
		leader := state.Players[state.LeaderIndex]
		mission := prompts.MissionProps{
			Index: state.MissionIndex,
			Size:  state.Missions[state.MissionIndex],
		}
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
		leaderProposal := FmtPlayers(prompts.ExtractCharacters(message, state.Players))

		for i := 1; i < len(state.Players); i++ {
			player := state.Players[(state.LeaderIndex+i)%len(state.Players)]

			message, err = player.Send(prompts.RenderCommentPrompt(prompts.VoteProps{
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

		message, err = leader.Send(prompts.RenderStatementPrompt(
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
		leaderTeam := prompts.ExtractCharacters(message, state.Players)
		votesForLeader := 0
		votesAgainstLeader := 0
		votes := map[string]string{}

		for i := 1; i < len(state.Players); i++ {
			player := state.Players[(state.LeaderIndex+i)%len(state.Players)]

			message, err = player.Send(prompts.RenderVotePrompt(
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

		if votesForLeader < votesAgainstLeader {
			state.SkipsCount++
			state.LeaderIndex = (state.LeaderIndex + 1) % len(state.Players)
			missionMessage := fmt.Sprintf(
				"Лидер %s не смог собрать состав на Миссию %d\n",
				leader.Persona.Self,
				state.MissionIndex+1,
			)
			logging.Info(missionMessage)
			logging.Event(state, missionMessage)
			continue
		}

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
			message, err = player.Send(prompts.RenderCompletionPrompt(
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
	if state.SkipsCount == 5 {
		gameMessage := "Игра окончена. 5 лидеров подряд не смогли собрать состав на миссию. Победила команда 'Красные'"
		logging.Info(gameMessage)
		logging.Event(state, gameMessage)
		return
	}
	if state.Fails == 3 {
		gameMessage := "Игра окончена. 3 миссии были провалены. Победила команда 'Красные'"
		logging.Info(gameMessage)
		logging.Event(state, gameMessage)
		return
	}
	if state.Wins == 3 {
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
}
