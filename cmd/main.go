package main

import (
	"avalon/pkg/action"
	"avalon/pkg/gemini"
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
	"strconv"
	"strings"
)

type CharacterConfig struct {
	Name string
	Mood string
	Risk string
}

var characters = []CharacterConfig{
	{"Сэр Эдрик", "осторожная, обтекаемая, с паузами и недосказанностью", "низкий — предпочитает тянуть время и играть наверняка"},
	{"Сэр Лиорен", "уверенная, развёрнутая, любит рассуждать вслух", "средний — готов рисковать, если может это красиво объяснить"},
	{"Сэр Кайлен", "короткая, прямая, без украшений", "высокий — легко идёт на риск, не боясь последствий"},
	{"Сэр Бранн", "редкая, тяжёлая, по делу", "очень низкий — рискует только если уверен почти полностью"},
	{"Леди Ивеллин", "быстрая, колкая, с акцентом на детали", "средне-высокий — любит острые ходы и провокации"},
}

func getActors(roles []prompts.Role, missions []int) []*gemini.Character {
	ctx := context.Background()

	_ = godotenv.Load()
	apiKey := os.Getenv("GEMINI_API_KEY")

	agent, err := gemini.NewAgent(ctx, apiKey)
	if err != nil {
		log.Fatal(err)
	}
	rolesOrder := prompts.GenRolesOrder(prompts.Roles5)
	players := extractPlayers(characters)

	actors := make([]*gemini.Character, 0, len(characters))

	redPlayers := "Вам известно что следующие игроки принадлежат команде 'Красные':"
	for i, cfg := range characters {
		if rolesOrder[i] == "Миньон Мордреда" || rolesOrder[i] == "Ассасин" {
			redPlayers += " " + cfg.Name
		}
	}

	for i, cfg := range characters {
		roleContext := ""
		if rolesOrder[i] == "Мерлин" || rolesOrder[i] == "Миньон Мордреда" || rolesOrder[i] == "Ассасин" {
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

func getResume(state *GameState) string {
	return fmt.Sprintf("На данный момент провалено %d миссий,  выполненно - %d миссий. %d лидеров поряд не смогли собрать состав на миссию.\n",
		state.Fails, state.Wins, state.SkipsCount)
}

func handleGame(state *GameState) {
	for state.Wins < 3 && state.Fails < 3 && state.SkipsCount < 5 {
		leader := state.Players[state.LeaderIndex]
		message, err := leader.Send("Вы лидер."+getResume(state)+
			fmt.Sprintf("Вы должны предложить состав на Миссию - %d, состав из %d любых игроков.\n", state.MissionIndex+1, state.Missions[state.MissionIndex])+
			"Последнее предложение речи должно быть в следующем формате: Выставить: имена игроков через запятую", state.Logs)
		if err != nil {
			panic(err)
		}
		fmt.Printf("[%s](%s): %s\n", leader.Persona.Self, leader.Persona.Role, message)
		wait()
		state.Logs = append(state.Logs, action.Action{User: leader.Persona.Self, Message: message})

		for i := 1; i < len(state.Players); i++ {
			player := state.Players[(state.LeaderIndex+i)%len(state.Players)]

			message, err = player.Send(getResume(state)+
				fmt.Sprintf("Лидер предложил состав на Миссию - %d (участников: %d)\n", state.MissionIndex+1, state.Missions[state.MissionIndex])+
				"Вы можете поддержать состав лидера, или любой другой озвученный ранее либо предложить альтернативный состав\n"+
				"В любом случае последнее предложение речи должно быть в следующем формате: Поддерживаю: имена игроков через запятую", state.Logs)
			if err != nil {
				panic(err)
			}
			fmt.Printf("[%s](%s): %s\n", player.Persona.Self, player.Persona.Role, message)
			wait()
			state.Logs = append(state.Logs, action.Action{User: player.Persona.Self, Message: message})
		}

		message, err = leader.Send("Вы лидер. "+getResume(state)+
			fmt.Sprintf("Вы должны выставить состав на голосование на Миссию - %d, состав из %d любых игроков.\n", state.MissionIndex+1, state.Missions[state.MissionIndex])+
			"Последнее предложение речи должно быть в следующем формате: Выставить: имена игроков через запятую", state.Logs)
		if err != nil {
			panic(err)
		}
		fmt.Printf("[%s](%s): %s\n", leader.Persona.Self, leader.Persona.Role, message)
		wait()
		state.Logs = append(state.Logs, action.Action{User: leader.Persona.Self, Message: message})
		leaderStatement, _ := extractTeam(message)
		leaderTeam := []*gemini.Character{}
		for _, l := range leaderStatement {
			idx := slices.IndexFunc(state.Players, func(p *gemini.Character) bool {
				return strings.ToLower(p.Persona.Self) == strings.ToLower(l)
			})
			leaderTeam = append(leaderTeam, state.Players[idx])
		}
		votesForLeader := 0
		votesAgainstLeader := 0

		for i := 1; i < len(state.Players); i++ {
			player := state.Players[(state.LeaderIndex+i)%len(state.Players)]

			message, err = player.Send(getResume(state)+
				fmt.Sprintf("Лидер выставил состав на Миссию - %d (число участников: %d)\n", state.MissionIndex+1, state.Missions[state.MissionIndex])+
				"Вы можете проголосовать либо против, либо за.\n"+
				"Последнее предложение в вашей речи должно быть либо Голосовать: ПРОТИВ, либо Голосовать: ЗА", state.Logs)
			if err != nil {
				panic(err)
			}
			fmt.Printf("[%s](%s): %s\n", player.Persona.Self, player.Persona.Role, message)
			wait()
			state.Logs = append(state.Logs, action.Action{User: player.Persona.Self, Message: message})
			if extractVote(message) {
				votesForLeader++
			} else {
				votesAgainstLeader++
			}
		}
		if votesForLeader < votesAgainstLeader {
			state.SkipsCount++
			state.LeaderIndex = (state.LeaderIndex + 1) % len(state.Players)
			missionMessage := "Лидер " + leader.Persona.Self + "не смог собрать состав на Миссию " + strconv.Itoa(state.MissionIndex+1) + "\n"
			fmt.Println("[Система]:" + missionMessage)
			state.Logs = append(state.Logs, action.Action{User: action.System, Message: missionMessage})
			continue
		}

		missionMessage := "Лидер " + leader.Persona.Self + "смог собрал состав на Миссию " + strconv.Itoa(state.MissionIndex+1) + "\n"
		fmt.Println("[Система]:" + missionMessage)
		state.Logs = append(state.Logs, action.Action{User: action.System, Message: missionMessage})

		votesFail := 0
		votesSuccess := 0

		for _, player := range leaderTeam {
			message, err = player.Send(getResume(state)+
				fmt.Sprintf("Лидер выставил вас в состав на Миссию - %d (число участников: %d)\n", state.MissionIndex+1, state.Missions[state.MissionIndex])+
				"Вы сейчас находитесь на Миссии. Вы можете либо успещно выполнить свою часть, либо провалить. Остальные игроки не узнают, что именно вы выбрали и не увидят вашу речь.\n"+
				"Но они будут знать число провалов и успехов в этой миссии\n"+
				"Для провала миссии достаточно 1 провала, а для успеха требуется, чтобы все участники миссии совершили успех\n"+
				"Последнее предложение в вашей речи должно быть либо Совершить: УСПЕХ или Совершить: ПРОВАЛ", state.Logs)
			if err != nil {
				panic(err)
			}
			fmt.Printf("[%s](%s): %s\n", player.Persona.Self, player.Persona.Role, message)
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
		missionMessage = "Результат миссии: " + strconv.Itoa(state.MissionIndex+1) +
			"\nУспехов: " + strconv.Itoa(votesSuccess) +
			"\nПровалов: " + strconv.Itoa(votesFail) + "\n" + result + "\n"
		state.Logs = append(state.Logs, action.Action{
			User:    action.System,
			Message: missionMessage,
		})
		fmt.Println("[Система]:" + missionMessage)
		state.SkipsCount = 0
		state.LeaderIndex = (state.LeaderIndex + 1) % len(state.Players)
	}
	if state.SkipsCount == 5 {
		gameMessage := "Игра окончена. 5 лидеров подряд не смогли собрать состав на миссию. Победила команда 'Красные'"
		state.Logs = append(state.Logs, action.Action{
			User:    action.System,
			Message: gameMessage,
		})
		fmt.Println("[Система]:" + gameMessage)
		return
	}
	if state.Fails == 3 {
		gameMessage := "Игра окончена. 3 миссии были провалены. Победила команда 'Красные'"
		state.Logs = append(state.Logs, action.Action{
			User:    action.System,
			Message: gameMessage,
		})
		fmt.Println("[Система]:" + gameMessage)
		return
	}
	if state.Wins == 3 {
		idx := slices.IndexFunc(state.Players, func(p *gemini.Character) bool {
			return p.Persona.Role == "Ассасин"
		})
		assassin := state.Players[idx]
		message, err := assassin.Send(
			"Команда 'Синие' близка к победе. Они выполнили успешно 3 миссии. Но у 'Красных есть шанс победить\n"+
				"Назовите имя игрока за столом, который по вашему мнению имеет роль 'Мерлин'\n"+
				"Если вы угадаете, то победа будет за 'Красными'\n"+
				"Последнее предложение вашей речи должно быть Выбрать: имя игрока",
			state.Logs,
		)
		if err != nil {
			panic(err)
		}
		fmt.Printf("[%s](%s): %s\n", assassin.Persona.Self, assassin.Persona.Role, message)
		wait()
		state.Logs = append(state.Logs, action.Action{User: assassin.Persona.Self, Message: message})
		targetName, _ := extractPlayerName(message)
		idx = slices.IndexFunc(state.Players, func(p *gemini.Character) bool {
			return strings.ToLower(targetName) == strings.ToLower(p.Persona.Self)
		})
		target := state.Players[idx]
		if target.Persona.Role == "Мерлин" {
			gameMessage := "Игра окончена. Ассассин убил Мерлина. Победила команда 'Красные'"
			state.Logs = append(state.Logs, action.Action{
				User:    action.System,
				Message: gameMessage,
			})
			fmt.Println("[Система]:" + gameMessage)
			return
		}
		gameMessage := "Игра окончена. Ассассин не нашел Мерлина. Победила команда 'Синие'"
		state.Logs = append(state.Logs, action.Action{
			User:    action.System,
			Message: gameMessage,
		})
		fmt.Println("[Система]:" + gameMessage)
	}
}

func main() {
	missions := prompts.Missions5
	players := getActors(prompts.Roles5, missions)
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
	handleGame(state)
}

func extractPlayers(chars []CharacterConfig) []string {
	players := make([]string, len(chars))
	for i, c := range chars {
		players[i] = c.Name
	}
	return players
}
