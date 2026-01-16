package domain

import (
	"avalon/pkg/constants"
	"avalon/pkg/dto"
	"avalon/pkg/gemini"
	"avalon/pkg/presets"
	"avalon/pkg/prompts"
	"context"
	"log"
)

func extractPlayers(chars []presets.CharacterConfig) []string {
	players := make([]string, len(chars))
	for i, c := range chars {
		players[i] = c.Name
	}
	return players
}

func GenerateActors(ctx context.Context, apiKey string, roles []presets.Role, missions []int) []*dto.Character {
	agent, err := gemini.NewAgent(ctx, apiKey)
	if err != nil {
		log.Fatal(err)
	}

	characters := presets.Characters5
	rolesOrder := presets.GenRolesOrder(presets.Roles5)
	players := extractPlayers(characters)

	actors := make([]*dto.Character, 0, len(characters))

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
		actor := dto.NewCharacter(
			agent,
			dto.Persona{
				Self:      cfg.Name,
				ModelName: "models/gemini-2.5-flash",
				Role:      rolesOrder[i],
			},
			prompts.GetSystemPrompt(prompts.SystemPromptProps{
				Name:        cfg.Name,
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
