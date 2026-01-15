package gemini

import (
	"avalon/pkg/action"
)

type Character struct {
	agent        *Agent
	Persona      Persona
	systemPrompt string
}

func NewCharacter(
	agent *Agent,
	persona Persona,
	systemPrompt string,
) *Character {

	return &Character{
		agent:        agent,
		Persona:      persona,
		systemPrompt: systemPrompt,
	}
}

func (c *Character) Send(
	instruction string,
	logs []action.Action,
) (string, error) {
	return c.agent.Send(c.Persona, c.systemPrompt, instruction, logs)
}
