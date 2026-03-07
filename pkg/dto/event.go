package dto

type Event struct {
	ID         int
	GameID     int
	PlayerID   int
	PlayerName string
	Type       string
	Content    string
	Hidden     bool
}
