package dto

type Character struct {
	agent        Agent
	Persona      Persona
	systemPrompt string
}

func NewCharacter(
	agent Agent,
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
	logs []Action,
) (string, error) {
	return c.agent.Send(c.Persona, c.systemPrompt, instruction, logs)
}

type PlayerV2 struct {
	ID       int
	Name     string
	Model    string
	Role     string
	Voice    string
	Mood     string
	Position int
	GameID   int
}

type MissionV2 struct {
	ID        int
	Name      string
	MaxFails  int
	SquadSize int
	Priority  int
	GameID    int
}

type GameV2 struct {
	ID              int
	MissionPriority int
	LeaderPosition  int
	SkipsCount      int
	Wins            int
	Fails           int
}
