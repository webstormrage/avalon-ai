package statemachine

import (
	"avalon/pkg/server"
	"avalon/pkg/store"
)

func GetState(h *Handler, tx store.QueryRower, gameID int) (*server.GameState, error) {
	return getState(h, tx, gameID)
}

func HandleNextState(h *Handler, gameID int) (*server.GameState, error) {
	return handleNextState(h, gameID)
}

func ApplyLeaderDiscussionPrompt(h *Handler, tx store.QueryRower, gameID int) error {
	return applyLeaderDiscussionPrompt(h, tx, gameID)
}

func ApplySpeakerDiscussionPrompt(h *Handler, tx store.QueryRower, gameID int) error {
	return applySpeakerDiscussionPrompt(h, tx, gameID)
}

func ApplyLeaderVotingPrompt(h *Handler, tx store.QueryRower, gameID int) error {
	return applyLeaderVotingPrompt(h, tx, gameID)
}

func ApplySpeakerVotingPrompt(h *Handler, tx store.QueryRower, gameID int) error {
	return applySpeakerVotingPrompt(h, tx, gameID)
}

func ApplyMissionPrompt(h *Handler, tx store.QueryRower, gameID int) error {
	return applyMissionPrompt(h, tx, gameID)
}

func ApplyAssassinationPrompt(h *Handler, tx store.QueryRower, gameID int) error {
	return applyAssassinationPrompt(h, tx, gameID)
}
