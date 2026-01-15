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
	RoundNumber  int
	SkipsCount   int
	Wins         int
	Fails        int
	WinnerTeam   string
	Logs         []action.Action
}

func wait() {
	reader := bufio.NewReader(os.Stdin)
	_, _, err := reader.ReadRune()
	if err != nil {
		panic(err)
	}
}

func handleGame(state *GameState) {
	for state.Wins < 3 && state.Fails < 3 && state.SkipsCount < 5 {
		leader := state.Players[state.LeaderIndex]
		message, err := leader.Send("Вы лидер. На данный момент провалено %d миссий,  выполненно - %d миссий. %d лидеров поряд не смогли собрать состав на миссию.\n"+
			"Вы должны собрать состав на Миссию - %d, состав из %d любых игроков.\n"+
			"Последнее предложение речи должно быть в следующем формате: Выставить: имена игроков через запятую", state.Logs)
		if err != nil {
			panic(err)
		}
		fmt.Printf("[%s](%s): %s\n", leader.Persona.Self, leader.Persona.Role, message)
		wait()
		state.Logs = append(state.Logs, action.Action{User: leader.Persona.Self, Message: message})

		for i := 1; i < len(state.Players); i++ {
			player := state.Players[(state.LeaderIndex+i)%len(state.Players)]

			message, err = player.Send(" На данный момент провалено %d миссий,  выполненно - %d миссий. %d лидеров поряд не смогли собрать состав на миссию.\n"+
				"Лидер предложил состав на Миссию - %d"+
				"Вы можете поддержать состав лидера, или любой другой озвученный ранее либо предложить альтернативный состав\n"+
				"В любом случае последнее предложение речи должно быть в следующем формате: Поддерживаю: имена игроков через запятую", state.Logs)
			if err != nil {
				panic(err)
			}
			fmt.Printf("[%s](%s): %s\n", player.Persona.Self, player.Persona.Role, message)
			wait()
			state.Logs = append(state.Logs, action.Action{User: player.Persona.Self, Message: message})
		}

		message, err = leader.Send("Вы лидер. На данный момент провалено %d миссий,  выполненно - %d миссий. %d лидеров поряд не смогли собрать состав на миссию.\n"+
			"Вы должны выставить на голосование состав на Миссию - %d, состав из %d любых игроков.\n"+
			"Последнее предложение речи должно быть в следующем формате: Выставить: имена игроков через запятую", state.Logs)
		if err != nil {
			panic(err)
		}
		fmt.Printf("[%s](%s): %s\n", leader.Persona.Self, leader.Persona.Role, message)
		wait()
		state.Logs = append(state.Logs, action.Action{User: leader.Persona.Self, Message: message})

		for i := 1; i < len(state.Players); i++ {
			player := state.Players[(state.LeaderIndex+i)%len(state.Players)]

			message, err = player.Send(" На данный момент провалено %d миссий,  выполненно - %d миссий. %d лидеров поряд не смогли собрать состав на миссию.\n"+
				"Лидер выставил состав на Миссию - %d для голосования"+
				"Вы можете проголосовать либо против, либо за.\n"+
				"Последнее слово в вашей речи должно быть либо ПРОТИВ, либо ЗА", state.Logs)
			if err != nil {
				panic(err)
			}
			fmt.Printf("[%s](%s): %s\n", player.Persona.Self, player.Persona.Role, message)
			wait()
			state.Logs = append(state.Logs, action.Action{User: player.Persona.Self, Message: message})
			// TODO: подсчет голосов
		}
		return // DEGUG return
		//TODO: обработка конца раунда, поход на миссию, обработка скипа, переключение лидера
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
		RoundNumber:  0,
		SkipsCount:   0,
		Wins:         0,
		Fails:        0,
		WinnerTeam:   "",
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
