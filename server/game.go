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

// Skill level system: BG1 < BG2 < S < N < P- < P < P+
var skillLevelOrder = []string{"BG1", "BG2", "S", "N", "P-", "P", "P+"}

var skillLevelValue = map[string]int{
	"BG1": 1,
	"BG2": 2,
	"S":   3,
	"N":   4,
	"P-":  5,
	"P":   6,
	"P+":  7,
}

var skillLevelLabels = map[string]string{
	"BG1": "มือใหม่เริ่มหัด",
	"BG2": "มือหน้าบ้าน",
	"S":   "มือกลาง S",
	"N":   "มือกลาง N",
	"P-":  "ใกล้เคียงโค้ช",
	"P":   "มือโค้ชทั่วไป",
	"P+":  "ฟอร์มนักกีฬา",
}

func getSkillValue(skillLevel string, skillStars int) float64 {
	base := skillLevelValue[skillLevel]
	if base == 0 {
		base = 1
	}
	// Stars (1-5) add fractional value within a skill level
	return float64(base) + float64(skillStars-1)*0.2
}

// calcSkillScaling returns an EXP multiplier and RP penalty multiplier
// based on the skill gap between winner and loser teams.
// If winner is stronger: reduced EXP (down to 0.3x), normal RP
// If winner is weaker: bonus EXP (up to 2.0x), normal RP
// If loser is stronger: 2x RP penalty
func calcSkillScaling(winnerAvgSkill, loserAvgSkill float64) (expMult float64, rpLosePenaltyMult float64) {
	gap := winnerAvgSkill - loserAvgSkill

	if gap > 0 {
		// Winner is stronger → less EXP for winning
		// Gap of 1 skill level → 0.7x, gap of 2 → 0.5x, gap of 3+ → 0.3x
		expMult = math.Max(0.3, 1.0-gap*0.3)
	} else if gap < 0 {
		// Winner is weaker → bonus EXP for upset
		// Gap of -1 → 1.3x, -2 → 1.6x, -3+ → 2.0x
		expMult = math.Min(2.0, 1.0+math.Abs(gap)*0.3)
	} else {
		expMult = 1.0
	}

	// RP penalty: if loser is stronger (had higher skill), double the RP loss
	if loserAvgSkill > winnerAvgSkill {
		rpLosePenaltyMult = 2.0
	} else {
		rpLosePenaltyMult = 1.0
	}

	return
}

var expRewards = map[string]map[string]int{
	"casual":     {"win": 30, "lose": 10, "draw": 15},
	"ranked":     {"win": 50, "lose": 15, "draw": 25},
	"skill_test": {"win": 40, "lose": 10, "draw": 20},
}

var rankPointRewards = map[string]int{
	"win":  25,
	"lose": -10,
	"draw": 5,
}
