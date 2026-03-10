package server

import (
	"avalon/pkg/constants"
	"avalon/pkg/dto"
	"avalon/pkg/presets"
	"avalon/pkg/prompts"
	"avalon/pkg/store"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type renderSystemPromptRequest struct {
	GameState GameState `json:"gameState"`
	PlayerID  int       `json:"playerId"`
}

type renderSystemPromptResponse struct {
	SystemPrompt string               `json:"systemPrompt"`
	ActionPrompt string               `json:"actionPrompt,omitempty"`
	History      []promptHistoryEvent `json:"history"`
}

type promptHistoryEvent struct {
	PlayerName string `json:"playerName"`
	Type       string `json:"type"`
	Content    string `json:"content"`
}

func (h *GameHandler) RenderSystemPrompt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req renderSystemPromptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.PlayerID == 0 {
		http.Error(w, "missing playerId", http.StatusBadRequest)
		return
	}
	if req.GameState.Game.ID == 0 {
		http.Error(w, "missing gameState.game.id", http.StatusBadRequest)
		return
	}

	player, err := store.GetPlayerByID(h.Ctx, h.DB, req.PlayerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if player == nil {
		http.Error(w, "player not found", http.StatusNotFound)
		return
	}
	if player.GameID != req.GameState.Game.ID {
		http.Error(w, "player does not belong to game", http.StatusForbidden)
		return
	}

	requiredAction := req.GameState.RequiredAction
	FilterStateByRequester(&req.GameState, *player)

	missions, err := store.GetMissionsByGameID(h.Ctx, h.DB, req.GameState.Game.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	events, err := store.GetEventsByGameID(h.Ctx, h.DB, req.GameState.Game.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	systemPrompt := prompts.GetSystemPrompt(prompts.SystemPromptProps{
		Name:     player.Name,
		Players:  req.GameState.Players,
		Roles:    presets.Roles5,
		Role:     player.Role,
		Missions: missions,
	})
	actionPrompt, err := h.renderActionPrompt(req.GameState.Game.ID, requiredAction)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(renderSystemPromptResponse{
		SystemPrompt: systemPrompt,
		ActionPrompt: actionPrompt,
		History:      mapPromptHistoryEvents(events, req.GameState.Players),
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func mapPromptHistoryEvents(events []*dto.Event, players []dto.PlayerV2) []promptHistoryEvent {
	playerNamesByID := make(map[int]string, len(players))
	for _, p := range players {
		playerNamesByID[p.ID] = p.Name
	}

	out := make([]promptHistoryEvent, 0, len(events))
	for _, event := range events {
		if event == nil {
			continue
		}
		playerName := event.PlayerName
		if playerName == "" {
			playerName = playerNamesByID[event.PlayerID]
		}
		if playerName == "" && event.PlayerID != 0 {
			playerName = fmt.Sprintf("player#%d", event.PlayerID)
		}

		out = append(out, promptHistoryEvent{
			PlayerName: playerName,
			Type:       event.Type,
			Content:    event.Content,
		})
	}

	return out
}

func (h *GameHandler) renderActionPrompt(gameID int, requiredAction *RequiredAction) (string, error) {
	if requiredAction == nil || requiredAction.Name == "" {
		return "", nil
	}

	game, err := store.GetGame(h.Ctx, h.DB, gameID)
	if err != nil {
		return "", err
	}
	mission, err := store.GetMissionByPriority(h.Ctx, h.DB, gameID, game.MissionPriority)
	if err != nil {
		return "", err
	}
	leader, err := store.GetPlayerByPosition(h.Ctx, h.DB, gameID, game.LeaderPosition)
	if err != nil {
		return "", err
	}

	switch requiredAction.Name {
	case "propose_squad":
		return prompts.RenderProposeSquadPrompt(prompts.StatementProps{
			Resume: prompts.ResumeProps{
				Wins:       game.Wins,
				Fails:      game.Fails,
				SkipsCount: game.SkipsCount,
			},
			Mission: *mission,
		}), nil
	case "rate_squad":
		team := joinSquadPositions(mission.Squad)
		return prompts.RenderRateSquadPrompt(prompts.VoteProps{
			Mission: *mission,
			Leader:  leader.Name,
			Team:    team,
		}), nil
	case "announce_squad":
		return prompts.RenderAnnounceSquadPrompt(prompts.StatementProps{
			Resume: prompts.ResumeProps{
				Wins:       game.Wins,
				Fails:      game.Fails,
				SkipsCount: game.SkipsCount,
			},
			Mission: *mission,
		}), nil
	case "vote_squad":
		team := joinSquadPositions(mission.Squad)
		return prompts.RenderVoteSquadPrompt(prompts.VoteProps{
			Mission: *mission,
			Leader:  leader.Name,
			Team:    team,
		}), nil
	case "mission_action":
		leaderName := leader.Name
		team := joinSquadPositions(mission.Squad)
		return prompts.RenderMissionActionPrompt(prompts.VoteProps{
			Mission: *mission,
			Leader:  leaderName,
			Team:    team,
		}), nil
	case "propose_assassination":
		return "Предложите публично цель для убийства. Последнее предложение должно быть в формате:\nВыставить: имя игрока\n", nil
	case "rate_assassination":
		speaker, err := store.GetPlayerByPosition(h.Ctx, h.DB, gameID, game.SpeakerPosition)
		if err != nil {
			return "", err
		}
		target := ""
		lastSpeech, err := store.GetLastEventByGameIDAndType(h.Ctx, h.DB, gameID, constants.EVENT_PLAYER_SPEECH)
		if err != nil {
			return "", err
		}
		if lastSpeech != nil {
			if extracted, ok := prompts.ExtractAssassinationTarget(lastSpeech.Content); ok {
				target = extracted
			}
		}
		if target == "" {
			target = "не указана"
		}
		return prompts.RenderRateAssassinationPrompt(prompts.RateAssassinationProps{
			Speaker: speaker.Name,
			Target:  target,
		}), nil
	case "announce_assassination":
		return prompts.AssassinationPrompt, nil
	default:
		return "", nil
	}
}

func joinSquadPositions(squad []int) string {
	if len(squad) == 0 {
		return ""
	}

	parts := make([]string, len(squad))
	for i, pos := range squad {
		parts[i] = fmt.Sprintf("%d", pos)
	}
	return strings.Join(parts, ", ")
}
