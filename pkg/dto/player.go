package dto

type PlayerV2 struct {
	ID            int    `json:"id,omitempty"`
	Name          string `json:"name"`
	Model         string `json:"model"`
	Role          string `json:"role,omitempty"`
	CharacterType string `json:"characterType"`
	Position      int    `json:"position"`
	Vote          string `json:"vote,omitempty"`
	MissionAction string `json:"missionAction,omitempty"`
	GameID        int    `json:"gameId"`
}
