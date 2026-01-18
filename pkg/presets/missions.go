package presets

import "avalon/pkg/dto"

func GetMissionsV2() []*dto.MissionV2 {
	return []*dto.MissionV2{
		&dto.MissionV2{
			Name:      "Миссия 1",
			MaxFails:  0,
			Priority:  1,
			SquadSize: 2,
		},
		&dto.MissionV2{
			Name:      "Миссия 2",
			MaxFails:  0,
			Priority:  2,
			SquadSize: 3,
		},
		&dto.MissionV2{
			Name:      "Миссия 3",
			MaxFails:  0,
			Priority:  3,
			SquadSize: 2,
		},
		&dto.MissionV2{
			Name:      "Миссия 4",
			MaxFails:  0,
			Priority:  4,
			SquadSize: 3,
		},
		&dto.MissionV2{
			Name:      "Миссия 5",
			MaxFails:  0,
			Priority:  5,
			SquadSize: 3,
		},
	}
}
