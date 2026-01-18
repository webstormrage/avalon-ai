package dto

type Agent interface {
	Send(
		persona Persona,
		systemPrompt string,
		instruction string,
		logs []*Event,
	) (string, error)

	Close() error
}

type TtsAgent interface {
	Send(
		model string,
		voiceName string,
		text string,
		systemStyle *string,
	) ([]byte, error)
}
