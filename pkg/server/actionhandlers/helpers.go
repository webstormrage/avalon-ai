package actionhandlers

import (
	"avalon/pkg/constants"
	"avalon/pkg/dto"
	"avalon/pkg/store"
	"context"
	"fmt"
	"strconv"
)

func getActionMessage(params GameActionParams) string {
	if params.Message == nil {
		return ""
	}
	return *params.Message
}

func squadNumbersToStrings(squadNumbers []int) ([]string, error) {
	if len(squadNumbers) == 0 {
		return nil, fmt.Errorf("params.squadNumbers must be non-empty")
	}
	roster := make([]string, 0, len(squadNumbers))
	for _, number := range squadNumbers {
		if number <= 0 {
			return nil, fmt.Errorf("invalid squad number: %d", number)
		}
		roster = append(roster, strconv.Itoa(number))
	}
	return roster, nil
}

func actionApproveToVote(params GameActionParams) (string, error) {
	if params.Approve == nil {
		return "", fmt.Errorf("params.approve must be boolean")
	}
	if *params.Approve {
		return "ЗА", nil
	}
	return "ПРОТИВ", nil
}

func actionSuccessToMissionResult(params GameActionParams) (string, error) {
	if params.Success == nil {
		return "", fmt.Errorf("params.success must be boolean")
	}
	if *params.Success {
		return "УСПЕХ", nil
	}
	return "ПРОВАЛ", nil
}

func findCurrentMission(ctx context.Context, tx store.QueryRower, gameID int) (*dto.MissionV2, error) {
	missions, err := store.GetMissionsByGameID(ctx, tx, gameID)
	if err != nil {
		return nil, err
	}
	for i := range missions {
		if missions[i].Status == "COMPLETED" || missions[i].Status == "FAILED" {
			continue
		}
		mission := missions[i]
		return &mission, nil
	}
	return nil, nil
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
		if p.Role == constants.ROLE_ASSASSIN || p.Role == constants.ROLE_MORDRED_MINION {
			if p.Position > 0 {
				order = append(order, p.Position)
			}
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
