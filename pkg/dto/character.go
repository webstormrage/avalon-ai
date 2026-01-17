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
	logs []*Event,
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
	SpeakerPosition int
	SkipsCount      int
	Wins            int
	Fails           int
	GameState       string
}

type Prompt struct {
	ID            int
	GameID        int
	Model         string
	SystemPrompt  string
	MessagePrompt string
	Response      string
	Status        string
}

type Event struct {
	ID      int
	GameID  int
	Source  string
	Type    string
	Content string
}
