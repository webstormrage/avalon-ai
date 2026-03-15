package actionhandlers

type GameActionParams struct {
	Message      *string `json:"message,omitempty"`
	Success      *bool   `json:"success,omitempty"`
	Approve      *bool   `json:"approve,omitempty"`
	SquadNumbers []int   `json:"squadNumbers,omitempty"`
	Target       *int    `json:"target,omitempty"`
}

type GameAction struct {
	PlayerID int              `json:"playerId"`
	Name     string           `json:"name"`
	Params   GameActionParams `json:"params"`
}
