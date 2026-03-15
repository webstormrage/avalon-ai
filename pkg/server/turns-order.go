package server

import (
	"avalon/pkg/constants"
	"avalon/pkg/dto"
	"avalon/pkg/store"
	"fmt"
)

func persistGameWithTurnsOrder(
	h *GameHandler,
	tx store.QueryRower,
	game *dto.GameV2,
) error {
	if err := setTurnsOrderForCurrentPhase(h, tx, game); err != nil {
		return err
	}
	return store.UpdateGame(h.Ctx, tx, game)
}

func setTurnsOrderForCurrentPhase(
	h *GameHandler,
	tx store.QueryRower,
	game *dto.GameV2,
) error {
	players, err := store.GetPlayersByGameID(h.Ctx, tx, game.ID)
	if err != nil {
		return err
	}

	switch game.Phase {
	case constants.STATE_DISCUSSION, constants.STATE_VOTING:
		order, err := orderedAllFromLeader(players, game.LeaderPosition)
		if err != nil {
			return err
		}
		game.TurnsOrder = order
		return nil
	case constants.STATE_MISSION:
		mission, err := store.GetMissionByPriority(h.Ctx, tx, game.ID, game.MissionPriority)
		if err != nil {
			return err
		}
		if mission == nil {
			game.TurnsOrder = []int{}
			return nil
		}
		game.TurnsOrder = append([]int(nil), mission.Squad...)
		return nil
	case constants.STATE_ASSASSINATION_DISCUSSION, constants.STATE_ASSASSIONATION_DISCUSSION:
		order, err := orderedRedsFromAssassin(players)
		if err != nil {
			return err
		}
		game.TurnsOrder = order
		return nil
	case constants.STATE_ASSASSINATION, constants.STATE_ASSASSIONATION:
		assassinPos, err := findAssassinPosition(players)
		if err != nil {
			return err
		}
		game.TurnsOrder = []int{assassinPos}
		return nil
	default:
		game.TurnsOrder = []int{}
		return nil
	}
}

func orderedAllFromLeader(players []dto.PlayerV2, leaderPos int) ([]int, error) {
	if len(players) == 0 {
		return []int{}, nil
	}

	startIdx := -1
	for i, p := range players {
		if p.Position == leaderPos {
			startIdx = i
			break
		}
	}
	if startIdx == -1 {
		return nil, fmt.Errorf("leader position %d not found", leaderPos)
	}

	order := make([]int, 0, len(players))
	i := startIdx
	for range len(players) {
		p := players[i]
		if p.Role != constants.ROLE_GAME_MASTER && p.Position > 0 {
			order = append(order, p.Position)
		}
		i = (i + 1) % len(players)
	}

	return order, nil
}

func orderedRingFromPosition(startPos int, count int) ([]int, error) {
	if count < 0 {
		return nil, fmt.Errorf("players count must be non-negative")
	}
	if count == 0 {
		return []int{}, nil
	}
	if startPos <= 0 || startPos > count {
		return nil, fmt.Errorf("start position %d is out of range 1..%d", startPos, count)
	}

	order := make([]int, 0, count)
	pos := startPos
	for range count {
		order = append(order, pos)
		pos++
		if pos > count {
			pos = 1
		}
	}

	return order, nil
}

func orderedRedsFromAssassin(players []dto.PlayerV2) ([]int, error) {
	if len(players) == 0 {
		return []int{}, nil
	}

	startIdx := -1
	for i, p := range players {
		if p.Role == constants.ROLE_ASSASSIN {
			startIdx = i
			break
		}
	}
	if startIdx == -1 {
		return nil, fmt.Errorf("assassin not found")
	}

	order := make([]int, 0, len(players))
	i := startIdx
	for range len(players) {
		p := players[i]
		if p.Role != constants.ROLE_ASSASSIN && p.Role != constants.ROLE_MORDRED_MINION {
			i = (i + 1) % len(players)
			continue
		}
		if p.Position > 0 {
			order = append(order, p.Position)
		}
		i = (i + 1) % len(players)
	}

	return order, nil
}

func findAssassinPosition(players []dto.PlayerV2) (int, error) {
	for _, p := range players {
		if p.Role == constants.ROLE_ASSASSIN {
			return p.Position, nil
		}
	}
	return 0, fmt.Errorf("assassin not found")
}
