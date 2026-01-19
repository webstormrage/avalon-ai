package dto

type Agent interface {
	Send(
		persona PlayerV2,
		systemPrompt string,
		instruction string,
		logs []*Event,
	) (string, error)

	Close() error
}

type TtsAgent interface {
	Send(
		persona PlayerV2,
		text string,
		systemStyle *string,
	) ([]byte, error)
}
