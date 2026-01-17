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
