package dto

type Event struct {
	ID      int
	GameID  int
	Source  string
	Type    string
	Content string
	Hidden  bool
}
