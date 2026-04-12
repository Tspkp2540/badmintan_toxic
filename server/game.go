package main

import "math"

func getExpForLevel(level int) int {
	return int(math.Floor(100 * math.Pow(1.2, float64(level-1))))
}

func getRankByLevel(level int) string {
	switch {
	case level >= 75:
		return "Grandmaster"
	case level >= 50:
		return "Master"
	case level >= 35:
		return "Diamond"
	case level >= 20:
		return "Platinum"
	case level >= 10:
		return "Gold"
	case level >= 5:
		return "Silver"
	default:
		return "Bronze"
	}
}

var expRewards = map[string]map[string]int{
	"casual": {"win": 30, "lose": 10, "draw": 15},
	"ranked": {"win": 50, "lose": 15, "draw": 25},
}

var rankPointRewards = map[string]int{
	"win":  25,
	"lose": -10,
	"draw": 5,
}
