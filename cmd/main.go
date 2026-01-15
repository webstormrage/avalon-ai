package main

import (
	"avalon/pkg/action"
	"avalon/pkg/constants"
	"avalon/pkg/gemini"
	"avalon/pkg/presets"
	"avalon/pkg/prompts"
	"bufio"
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"math/rand"
	"os"
	"regexp"
	"slices"
	"strings"
)

func getActors(roles []presets.Role, missions []int) []*gemini.Character {
	ctx := context.Background()

	_ = godotenv.Load()
	apiKey := os.Getenv("GEMINI_API_KEY")

	agent, err := gemini.NewAgent(ctx, apiKey)
	if err != nil {
		log.Fatal(err)
	}

	characters := presets.Characters5
	rolesOrder := presets.GenRolesOrder(presets.Roles5)
	players := extractPlayers(characters)

	actors := make([]*gemini.Character, 0, len(characters))

	redPlayers := "Вам известно что следующие игроки принадлежат команде 'Красные':"
	for i, cfg := range characters {
		if rolesOrder[i] == constants.ROLE_MORDRED_MINION || rolesOrder[i] == constants.ROLE_ASSASSIN {
			redPlayers += " " + cfg.Name
		}
	}

	for i, cfg := range characters {
		roleContext := ""
		if rolesOrder[i] == constants.ROLE_MERLIN || rolesOrder[i] == constants.ROLE_MORDRED_MINION || rolesOrder[i] == constants.ROLE_ASSASSIN {
			roleContext = redPlayers
		}
		actor := gemini.NewCharacter(
			agent,
			gemini.Persona{
				Self:      cfg.Name,
				ModelName: "models/gemini-2.5-flash",
				Role:      rolesOrder[i],
			},
			prompts.GetSystemPrompt(prompts.Props{
				Self:        cfg.Name,
				Mood:        cfg.Mood,
				Risk:        cfg.Risk,
				Players:     players,
				Roles:       roles,
				Role:        rolesOrder[i],
				RoleContext: roleContext,
				Missions:    missions,
			}),
		)

		actors = append(actors, actor)
	}
	return actors
}

type GameState struct {
	Missions     []int
	Players      []*gemini.Character
	MissionIndex int
	LeaderIndex  int
	SkipsCount   int
	Wins         int
	Fails        int
	Logs         []action.Action
}

func wait() {
	reader := bufio.NewReader(os.Stdin)
	_, _, err := reader.ReadRune()
	if err != nil {
		panic(err)
	}
}

func extractVote(text string) bool {
	re := regexp.MustCompile(`(?i)Голосовать:\s*(ЗА|ПРОТИВ)\s*$`)
	m := re.FindStringSubmatch(text)

	if len(m) < 2 {
		return false
	}

	// нормализуем результат
	switch strings.ToUpper(m[1]) {
	case "ЗА":
		return true
	case "ПРОТИВ":
		return false
	default:
		return false
	}
}

func extractTeam(text string) ([]string, bool) {
	re := regexp.MustCompile(`(?i)Выставить:\s*(.+)\s*$`)
	m := re.FindStringSubmatch(text)

	if len(m) < 2 {
		return nil, false
	}

	raw := m[1]
	parts := strings.Split(raw, ",")

	var players []string
	for _, p := range parts {
		name := strings.TrimSpace(p)
		if name != "" {
			players = append(players, name)
		}
	}

	if len(players) == 0 {
		return nil, false
	}

	return players, true
}

func extractMissionResult(text string) bool {
	re := regexp.MustCompile(`(?i)Совершить:\s*(УСПЕХ|ПРОВАЛ)\s*$`)
	m := re.FindStringSubmatch(text)

	if len(m) < 2 {
		return true
	}

	switch strings.ToUpper(m[1]) {
	case "УСПЕХ":
		return true
	case "ПРОВАЛ":
		return false
	default:
		return true
	}
}

func extractPlayerName(text string) (string, bool) {
	re := regexp.MustCompile(`(?i)Выбрать:\s*(.+)\s*$`)
	m := re.FindStringSubmatch(text)

	if len(m) < 2 {
		return "", false
	}

	name := strings.TrimSpace(m[1])
	if name == "" {
		return "", false
	}

	return name, true
}

func messageToTeam(message string, players []*gemini.Character) []*gemini.Character {
	leaderStatement, _ := extractTeam(message)
	leaderTeam := []*gemini.Character{}
	for _, l := range leaderStatement {
		idx := slices.IndexFunc(players, func(p *gemini.Character) bool {
			return strings.ToLower(p.Persona.Self) == strings.ToLower(l)
		})
		leaderTeam = append(leaderTeam, players[idx])
	}
	return leaderTeam
}

func playersToString(players []*gemini.Character) string {
	names := []string{}
	for _, player := range players {
		names = append(names, player.Persona.Self)
	}
	return strings.Join(names, ", ")
}

func votesToString(leader string, team string, votes map[string]string) string {
	items := []string{}
	for k, v := range votes {
		items = append(items, fmt.Sprintf("%s: %s", k, v))
	}
	return fmt.Sprintf("Результаты голосования (Лидер: %s, Команда: %s):\n %s\n", leader, team, strings.Join(items, "\n"))
}

func consoleLog(player *gemini.Character, message string) {
	fmt.Printf("[%s](%s): %s\n", player.Persona.Self, player.Persona.Role, message)
}

func systemLog(message string) {
	fmt.Printf("<Игровое Событие>: %s\n", message)
}

func registerAction(state *GameState, player *gemini.Character, message string) {
	state.Logs = append(state.Logs, action.Action{User: player.Persona.Self, Message: message})
}

func registerSystemAction(state *GameState, message string) {
	state.Logs = append(state.Logs, action.Action{User: action.System, Message: message})
}

func runGame(state *GameState) {
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
		consoleLog(leader, message)
		wait()
		registerAction(state, leader, message)
		leaderProposal := playersToString(messageToTeam(message, state.Players))

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
			consoleLog(player, message)
			wait()
			registerAction(state, player, message)
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
		consoleLog(leader, message)
		wait()
		registerAction(state, leader, message)
		leaderTeam := messageToTeam(message, state.Players)
		votesForLeader := 0
		votesAgainstLeader := 0
		votes := map[string]string{}

		for i := 1; i < len(state.Players); i++ {
			player := state.Players[(state.LeaderIndex+i)%len(state.Players)]

			message, err = player.Send(prompts.RenderVotePrompt(
				prompts.VoteProps{
					Leader:  leader.Persona.Self,
					Team:    playersToString(leaderTeam),
					Mission: mission,
				},
			), state.Logs)
			if err != nil {
				panic(err)
			}
			consoleLog(player, message)
			wait()
			if extractVote(message) {
				votesForLeader++
				votes[player.Persona.Self] = "ЗА"
			} else {
				votesAgainstLeader++
				votes[player.Persona.Self] = "ПРОТИВ"
			}
		}
		votingMessage := votesToString(leader.Persona.Self, playersToString(leaderTeam), votes)
		systemLog(votingMessage)
		registerSystemAction(state, votingMessage)

		if votesForLeader < votesAgainstLeader {
			state.SkipsCount++
			state.LeaderIndex = (state.LeaderIndex + 1) % len(state.Players)
			missionMessage := fmt.Sprintf(
				"Лидер %s не смог собрать состав на Миссию %d\n",
				leader.Persona.Self,
				state.MissionIndex+1,
			)
			systemLog(missionMessage)
			registerSystemAction(state, missionMessage)
			continue
		}

		missionMessage := fmt.Sprintf(
			"Лидер %s смог собрать состав на Миссию %d - %s\n",
			leader.Persona.Self,
			state.MissionIndex+1,
			playersToString(leaderTeam),
		)
		systemLog(missionMessage)
		registerSystemAction(state, missionMessage)

		votesFail := 0
		votesSuccess := 0

		for _, player := range leaderTeam {
			message, err = player.Send(prompts.RenderCompletionPrompt(
				prompts.VoteProps{
					Leader:  leader.Persona.Self,
					Team:    playersToString(leaderTeam),
					Mission: mission,
				},
			), state.Logs)
			if err != nil {
				panic(err)
			}
			consoleLog(player, message)
			wait()
			if extractMissionResult(message) {
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
		systemLog(missionMessage)
		registerSystemAction(state, missionMessage)
		state.SkipsCount = 0
		state.MissionIndex++
		state.LeaderIndex = (state.LeaderIndex + 1) % len(state.Players)
	}
	if state.SkipsCount == 5 {
		gameMessage := "Игра окончена. 5 лидеров подряд не смогли собрать состав на миссию. Победила команда 'Красные'"
		systemLog(gameMessage)
		registerSystemAction(state, gameMessage)
		return
	}
	if state.Fails == 3 {
		gameMessage := "Игра окончена. 3 миссии были провалены. Победила команда 'Красные'"
		systemLog(gameMessage)
		registerSystemAction(state, gameMessage)
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
		consoleLog(assassin, message)
		wait()
		registerAction(state, assassin, message)
		targetName, _ := extractPlayerName(message)
		idx = slices.IndexFunc(state.Players, func(p *gemini.Character) bool {
			return strings.ToLower(targetName) == strings.ToLower(p.Persona.Self)
		})
		target := state.Players[idx]
		if target.Persona.Role == constants.ROLE_MERLIN {
			gameMessage := "Игра окончена. Ассассин убил Мерлина. Победила команда 'Красные'"
			systemLog(gameMessage)
			registerSystemAction(state, gameMessage)
			return
		}
		gameMessage := "Игра окончена. Ассассин не нашел Мерлина. Победила команда 'Синие'"
		systemLog(gameMessage)
		registerSystemAction(state, gameMessage)
	}
}

func main() {
	missions := presets.Missions5
	players := getActors(presets.Roles5, missions)
	state := &GameState{
		Missions:     missions,
		Players:      players,
		MissionIndex: 0,
		LeaderIndex:  rand.Intn(len(players)),
		SkipsCount:   0,
		Wins:         0,
		Fails:        0,
		Logs:         []action.Action{},
	}
	runGame(state)
}

func extractPlayers(chars []presets.CharacterConfig) []string {
	players := make([]string, len(chars))
	for i, c := range chars {
		players[i] = c.Name
	}
	return players
}
