package server

import (
	"avalon/pkg/constants"
	"avalon/pkg/dto"
	"avalon/pkg/presets"
	"avalon/pkg/prompts"
	"avalon/pkg/store"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
)

func createDiscussionPrompt(ctx context.Context, db *sql.DB, game *dto.GameV2) error {
	leader, err := store.GetPlayerByPosition(ctx, db, game.ID, game.LeaderPosition)
	if err != nil {
		return err
	}
	players, err := store.GetPlayersByGameID(ctx, db, game.ID)
	if err != nil {
		return err
	}
	mission, err := store.GetMissionByPriority(ctx, db, game.ID, game.MissionPriority)
	if err != nil {
		return err
	}
	missions, err := store.GetMissionsByGameID(ctx, db, game.ID)
	if err != nil {
		return err
	}

	playersNames := []string{}
	redPlayers := "Вам известно что следующие игроки принадлежат команде 'Красные':"
	for _, player := range players {
		playersNames = append(playersNames, player.Name)
		if player.Role == constants.ROLE_ASSASSIN || player.Role == constants.ROLE_MORDRED_MINION {
			redPlayers += " " + player.Name
		}
	}

	if game.SpeakerPosition == game.LeaderPosition {
		roleContext := ""
		if leader.Role == constants.ROLE_ASSASSIN || leader.Role == constants.ROLE_MERLIN || leader.Role == constants.ROLE_MORDRED_MINION {
			roleContext = redPlayers
		}
		return store.CreatePrompt(ctx, db, &dto.Prompt{
			GameID: game.ID,
			Model:  leader.Model,
			SystemPrompt: prompts.GetSystemPrompt(
				prompts.SystemPromptProps{
					Name:        leader.Name,
					Mood:        leader.Mood,
					Players:     playersNames,
					Roles:       presets.Roles5, // TODO: надо брать из базы
					Role:        leader.Role,
					RoleContext: roleContext,
					Missions:    missions,
				},
			),
			MessagePrompt: prompts.RenderProposalPrompt(prompts.StatementProps{
				Resume: prompts.ResumeProps{
					Wins:       game.Wins,
					Fails:      game.Fails,
					SkipsCount: game.SkipsCount,
				},
				Mission: *mission,
			}),
		})
	}
	return err
}

func stateMachine(ctx context.Context, db *sql.DB, game *dto.GameV2, prompts []dto.Prompt) error {
	var err error
	switch game.GameState {
	case constants.STATE_DISCUSSION:
		{
			if len(prompts) == 0 {
				err = createDiscussionPrompt(ctx, db, game)
			} else {
				switch prompts[0].Status {
				case constants.STATUS_NOT_STARTED:
					// sendDiscussionPrompt(game, prompts[0])
				case constants.STATUS_HAS_RESPONSE:
					// approveDiscussionPrompt(game, prompts[0])
				case constants.STATUS_APPROVED:
					// applyDiscussionPrompt(game, prompts[0])
				}
			}
		}
	}
	return err
}

func (h *GameHandler) NextTick(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// получить gameId из query params
	gameIDStr := r.URL.Query().Get("gameId")
	if gameIDStr == "" {
		http.Error(w, "missing gameId query param", http.StatusBadRequest)
		return
	}

	gameID, err := strconv.Atoi(gameIDStr)
	if err != nil {
		http.Error(w, "invalid gameId", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	game, err := store.GetGame(ctx, h.DB, gameID)

	if err != nil {
		http.Error(w, "Failed to find game", http.StatusInternalServerError)
		return
	}

	prompts, err := store.GetPromptsNotCompletedByGameID(ctx, h.DB, gameID)

	if err != nil || len(prompts) > 1 {
		http.Error(w, "too many uncompleted prompts", http.StatusInternalServerError)
		return
	}

	err = stateMachine(ctx, h.DB, game, prompts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]int{
		"gameId": gameID,
	})
}
