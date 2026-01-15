package dto

type GameState struct {
	Missions     []int
	Players      []*Character
	MissionIndex int
	LeaderIndex  int
	SkipsCount   int
	Wins         int
	Fails        int
	Logs         []Action
}
