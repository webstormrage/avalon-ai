package dto

type Agent interface {
	Send(
		persona Persona,
		systemPrompt string,
		instruction string,
		logs []Action,
	) (string, error)

	Close() error
}
