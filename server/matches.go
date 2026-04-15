package main

import (
	"database/sql"
	"math"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type matchPlayerResp struct {
	UserID           string  `json:"userId"`
	Username         string  `json:"username"`
	FullName         string  `json:"fullName"`
	AvatarURL        *string `json:"avatarUrl"`
	Level            int     `json:"level"`
	Rank             string  `json:"rank"`
	SkillLevel       string  `json:"skillLevel"`
	SkillStars       int     `json:"skillStars"`
	Team             string  `json:"team"`
	ExpGained        int     `json:"expGained"`
	RankPointsGained int     `json:"rankPointsGained"`
}

type matchSetResp struct {
	SetNumber int `json:"setNumber"`
	TeamA     int `json:"teamA"`
	TeamB     int `json:"teamB"`
}

type refereeResp struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	FullName string `json:"fullName"`
}

type matchResponse struct {
	ID         string            `json:"id"`
	CourtID    *string           `json:"courtId"`
	Name       string            `json:"name"`
	MatchType  string            `json:"matchType"`
	MatchMode  string            `json:"matchMode"`
	MaxSets    int               `json:"maxSets"`
	Status     string            `json:"status"`
	WinnerTeam *string           `json:"winnerTeam"`
	Referee    *refereeResp      `json:"referee"`
	CreatedBy  string            `json:"createdBy"`
	StartedAt  *string           `json:"startedAt"`
	EndedAt    *string           `json:"endedAt"`
	CreatedAt  string            `json:"createdAt"`
	Players    []matchPlayerResp `json:"players"`
	Sets       []matchSetResp    `json:"sets"`
}

func buildMatchResponse(matchID string) (*matchResponse, error) {
	var m matchResponse
	var winnerTeam, startedAt, endedAt, refereeID, courtID sql.NullString

	err := sqlDB.QueryRow(`SELECT id, court_id, room_name, match_type, match_mode, max_sets, status,
		winner_team, referee_id, created_by, started_at, ended_at, created_at
		FROM matches WHERE id = ?`, matchID).Scan(
		&m.ID, &courtID, &m.Name, &m.MatchType, &m.MatchMode, &m.MaxSets, &m.Status,
		&winnerTeam, &refereeID, &m.CreatedBy, &startedAt, &endedAt, &m.CreatedAt)
	if err != nil {
		return nil, err
	}

	if courtID.Valid {
		m.CourtID = &courtID.String
	}
	if winnerTeam.Valid {
		m.WinnerTeam = &winnerTeam.String
	}
	if startedAt.Valid {
		m.StartedAt = &startedAt.String
	}
	if endedAt.Valid {
		m.EndedAt = &endedAt.String
	}

	// Referee
	if refereeID.Valid {
		var ref refereeResp
		err := sqlDB.QueryRow("SELECT id, username, full_name FROM users WHERE id = ?", refereeID.String).
			Scan(&ref.ID, &ref.Username, &ref.FullName)
		if err == nil {
			m.Referee = &ref
		}
	}

	// Players
	pRows, _ := sqlDB.Query(`
		SELECT mp.user_id, u.username, u.full_name, u.avatar_url, u.level, u.rank, u.skill_level, u.skill_stars,
		       mp.team, mp.exp_gained, mp.rank_points_gained
		FROM match_players mp JOIN users u ON u.id = mp.user_id
		WHERE mp.match_id = ?`, matchID)
	if pRows != nil {
		defer pRows.Close()
	}
	m.Players = []matchPlayerResp{}
	if pRows != nil {
		for pRows.Next() {
			var p matchPlayerResp
			var avatar sql.NullString
			pRows.Scan(&p.UserID, &p.Username, &p.FullName, &avatar, &p.Level, &p.Rank,
				&p.SkillLevel, &p.SkillStars, &p.Team, &p.ExpGained, &p.RankPointsGained)
			if avatar.Valid {
				p.AvatarURL = &avatar.String
			}
			m.Players = append(m.Players, p)
		}
	}

	// Sets
	sRows, _ := sqlDB.Query(`SELECT set_number, team_a_score, team_b_score
		FROM match_sets WHERE match_id = ? ORDER BY set_number`, matchID)
	if sRows != nil {
		defer sRows.Close()
	}
	m.Sets = []matchSetResp{}
	if sRows != nil {
		for sRows.Next() {
			var s matchSetResp
			sRows.Scan(&s.SetNumber, &s.TeamA, &s.TeamB)
			m.Sets = append(m.Sets, s)
		}
	}

	return &m, nil
}

func handleGetMatches(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	var query string
	var args []interface{}

	if status != "" {
		query = "SELECT id FROM matches WHERE status = ? ORDER BY created_at DESC"
		args = append(args, status)
	} else {
		query = "SELECT id FROM matches WHERE status IN ('waiting', 'playing', 'scoring') ORDER BY created_at DESC"
	}

	rows, err := sqlDB.Query(query, args...)
	if err != nil {
		writeError(w, 500, "เกิดข้อผิดพลาด")
		return
	}
	defer rows.Close()

	matches := []matchResponse{}
	for rows.Next() {
		var id string
		rows.Scan(&id)
		m, err := buildMatchResponse(id)
		if err == nil {
			matches = append(matches, *m)
		}
	}
	writeJSON(w, 200, matches)
}

func handleCreateMatch(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	var req struct {
		CourtID   string `json:"courtId"`
		Name      string `json:"name"`
		MatchType string `json:"matchType"`
		MatchMode string `json:"matchMode"`
		MaxSets   int    `json:"maxSets"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, 400, "ข้อมูลไม่ถูกต้อง")
		return
	}

	if req.Name == "" || req.MatchType == "" || req.MatchMode == "" {
		writeError(w, 400, "กรุณากรอกข้อมูลให้ครบ")
		return
	}
	if req.CourtID == "" {
		writeError(w, 400, "กรุณาเลือกสนาม")
		return
	}
	if req.MatchType != "singles" && req.MatchType != "doubles" {
		writeError(w, 400, "matchType ต้องเป็น singles หรือ doubles")
		return
	}
	if req.MatchMode != "casual" && req.MatchMode != "ranked" && req.MatchMode != "skill_test" {
		writeError(w, 400, "matchMode ต้องเป็น casual, ranked หรือ skill_test")
		return
	}
	if req.MaxSets == 0 {
		req.MaxSets = 3
	}

	// Verify court exists and is open
	var courtStatus string
	var maxRooms int
	err := sqlDB.QueryRow("SELECT status, max_rooms FROM courts WHERE id = ?", req.CourtID).Scan(&courtStatus, &maxRooms)
	if err != nil {
		writeError(w, 404, "ไม่พบสนาม")
		return
	}
	if courtStatus != "open" {
		writeError(w, 400, "สนามนี้ปิดอยู่")
		return
	}

	// skill_test mode: only leaders assigned to this court can create
	if req.MatchMode == "skill_test" {
		userRole := getUserRole(r)
		if userRole != "admin" {
			var isCourtLeader int
			sqlDB.QueryRow("SELECT COUNT(*) FROM court_leaders WHERE court_id = ? AND user_id = ?",
				req.CourtID, userID).Scan(&isCourtLeader)
			if isCourtLeader == 0 {
				writeError(w, 403, "เฉพาะหัวก๊วนที่ได้รับมอบหมายสนามนี้เท่านั้นที่สร้างแมตช์ทดสอบระดับได้")
				return
			}
		}
	}

	// Check room limit
	var activeRooms int
	sqlDB.QueryRow(`SELECT COUNT(*) FROM matches WHERE court_id = ? AND status IN ('waiting', 'playing', 'scoring')`,
		req.CourtID).Scan(&activeRooms)
	if activeRooms >= maxRooms {
		writeError(w, 400, "สนามเต็มแล้ว ไม่สามารถสร้างห้องเพิ่มได้")
		return
	}

	id := uuid.New().String()
	sqlDB.Exec(`INSERT INTO matches (id, court_id, room_name, match_type, match_mode, max_sets, created_by)
		VALUES (?, ?, ?, ?, ?, ?, ?)`, id, req.CourtID, req.Name, req.MatchType, req.MatchMode, req.MaxSets, userID)

	sqlDB.Exec(`INSERT INTO match_players (match_id, user_id, team) VALUES (?, ?, 'A')`, id, userID)

	m, _ := buildMatchResponse(id)
	writeJSON(w, 201, m)
}

func handleGetMatch(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	m, err := buildMatchResponse(id)
	if err != nil {
		writeError(w, 404, "ไม่พบห้อง")
		return
	}
	writeJSON(w, 200, m)
}

func handleJoinMatch(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	matchID := chi.URLParam(r, "id")

	var req struct {
		Team string `json:"team"`
	}
	if err := readJSON(r, &req); err != nil || (req.Team != "A" && req.Team != "B") {
		writeError(w, 400, "กรุณาเลือกทีม A หรือ B")
		return
	}

	var status, matchType string
	var refereeID sql.NullString
	err := sqlDB.QueryRow("SELECT status, match_type, referee_id FROM matches WHERE id = ?", matchID).
		Scan(&status, &matchType, &refereeID)
	if err != nil {
		writeError(w, 404, "ไม่พบห้อง")
		return
	}
	if status != "waiting" {
		writeError(w, 400, "ห้องไม่อยู่ในสถานะรอผู้เล่น")
		return
	}

	var existID int
	if sqlDB.QueryRow("SELECT id FROM match_players WHERE match_id = ? AND user_id = ?", matchID, userID).
		Scan(&existID) == nil {
		writeError(w, 400, "คุณอยู่ในห้องนี้แล้ว")
		return
	}

	if refereeID.Valid && refereeID.String == userID {
		writeError(w, 400, "คุณเป็นกรรมการอยู่แล้ว")
		return
	}

	maxPerTeam := 1
	if matchType == "doubles" {
		maxPerTeam = 2
	}
	var teamCount int
	sqlDB.QueryRow("SELECT COUNT(*) FROM match_players WHERE match_id = ? AND team = ?", matchID, req.Team).
		Scan(&teamCount)
	if teamCount >= maxPerTeam {
		writeError(w, 400, "ทีมนี้เต็มแล้ว")
		return
	}

	sqlDB.Exec("INSERT INTO match_players (match_id, user_id, team) VALUES (?, ?, ?)", matchID, userID, req.Team)

	m, _ := buildMatchResponse(matchID)
	writeJSON(w, 200, m)
}

func handleJoinAsReferee(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	matchID := chi.URLParam(r, "id")

	var status string
	var refereeID sql.NullString
	err := sqlDB.QueryRow("SELECT status, referee_id FROM matches WHERE id = ?", matchID).
		Scan(&status, &refereeID)
	if err != nil {
		writeError(w, 404, "ไม่พบห้อง")
		return
	}
	if status != "waiting" {
		writeError(w, 400, "ห้องไม่อยู่ในสถานะรอผู้เล่น")
		return
	}
	if refereeID.Valid {
		writeError(w, 400, "มีกรรมการแล้ว")
		return
	}

	var existID int
	if sqlDB.QueryRow("SELECT id FROM match_players WHERE match_id = ? AND user_id = ?", matchID, userID).
		Scan(&existID) == nil {
		writeError(w, 400, "คุณเป็นผู้เล่นอยู่แล้ว")
		return
	}

	sqlDB.Exec("UPDATE matches SET referee_id = ? WHERE id = ?", userID, matchID)

	m, _ := buildMatchResponse(matchID)
	writeJSON(w, 200, m)
}

func handleStartMatch(w http.ResponseWriter, r *http.Request) {
	matchID := chi.URLParam(r, "id")

	var status, matchType string
	err := sqlDB.QueryRow("SELECT status, match_type FROM matches WHERE id = ?", matchID).
		Scan(&status, &matchType)
	if err != nil {
		writeError(w, 404, "ไม่พบห้อง")
		return
	}
	if status != "waiting" {
		writeError(w, 400, "ห้องไม่อยู่ในสถานะรอผู้เล่น")
		return
	}

	var teamACount, teamBCount int
	sqlDB.QueryRow("SELECT COUNT(*) FROM match_players WHERE match_id = ? AND team = 'A'", matchID).Scan(&teamACount)
	sqlDB.QueryRow("SELECT COUNT(*) FROM match_players WHERE match_id = ? AND team = 'B'", matchID).Scan(&teamBCount)

	if teamACount == 0 || teamBCount == 0 {
		writeError(w, 400, "ต้องมีผู้เล่นอย่างน้อยฝั่งละ 1 คน")
		return
	}
	if matchType == "doubles" && (teamACount < 2 || teamBCount < 2) {
		writeError(w, 400, "เกมคู่ต้องมีผู้เล่นฝั่งละ 2 คน")
		return
	}

	sqlDB.Exec("UPDATE matches SET status = 'playing', started_at = datetime('now') WHERE id = ?", matchID)

	m, _ := buildMatchResponse(matchID)
	writeJSON(w, 200, m)
}

func handleEndMatch(w http.ResponseWriter, r *http.Request) {
	matchID := chi.URLParam(r, "id")

	var status string
	err := sqlDB.QueryRow("SELECT status FROM matches WHERE id = ?", matchID).Scan(&status)
	if err != nil {
		writeError(w, 404, "ไม่พบห้อง")
		return
	}
	if status != "playing" {
		writeError(w, 400, "เกมไม่ได้กำลังแข่ง")
		return
	}

	sqlDB.Exec("UPDATE matches SET status = 'scoring', ended_at = datetime('now') WHERE id = ?", matchID)

	m, _ := buildMatchResponse(matchID)
	writeJSON(w, 200, m)
}

func handleSubmitScores(w http.ResponseWriter, r *http.Request) {
	matchID := chi.URLParam(r, "id")

	var status, matchMode string
	var courtID sql.NullString
	err := sqlDB.QueryRow("SELECT status, match_mode, court_id FROM matches WHERE id = ?", matchID).
		Scan(&status, &matchMode, &courtID)
	if err != nil {
		writeError(w, 404, "ไม่พบห้อง")
		return
	}
	if status != "scoring" {
		writeError(w, 400, "เกมไม่อยู่ในสถานะกรอกคะแนน")
		return
	}

	var req struct {
		Sets []struct {
			SetNumber int `json:"setNumber"`
			TeamA     int `json:"teamA"`
			TeamB     int `json:"teamB"`
		} `json:"sets"`
	}
	if err := readJSON(r, &req); err != nil || len(req.Sets) == 0 {
		writeError(w, 400, "กรุณากรอกคะแนนอย่างน้อย 1 เซ็ต")
		return
	}

	for _, s := range req.Sets {
		if s.TeamA < 0 || s.TeamB < 0 || s.TeamA > 30 || s.TeamB > 30 {
			writeError(w, 400, "คะแนนต้องอยู่ระหว่าง 0-30")
			return
		}
	}

	tx, _ := sqlDB.Begin()
	defer tx.Rollback()

	for _, s := range req.Sets {
		tx.Exec("INSERT INTO match_sets (match_id, set_number, team_a_score, team_b_score) VALUES (?, ?, ?, ?)",
			matchID, s.SetNumber, s.TeamA, s.TeamB)
	}

	teamAWins, teamBWins := 0, 0
	for _, s := range req.Sets {
		if s.TeamA > s.TeamB {
			teamAWins++
		} else if s.TeamB > s.TeamA {
			teamBWins++
		}
	}
	winner := "draw"
	if teamAWins > teamBWins {
		winner = "A"
	} else if teamBWins > teamAWins {
		winner = "B"
	}

	tx.Exec("UPDATE matches SET status = 'finished', winner_team = ? WHERE id = ?", winner, matchID)

	pRows, _ := tx.Query("SELECT id, user_id, team FROM match_players WHERE match_id = ?", matchID)
	type playerInfo struct {
		ID         int
		UserID     string
		Team       string
		SkillLevel string
		SkillStars int
	}
	var players []playerInfo
	for pRows.Next() {
		var p playerInfo
		pRows.Scan(&p.ID, &p.UserID, &p.Team)
		tx.QueryRow("SELECT skill_level, skill_stars FROM users WHERE id = ?", p.UserID).
			Scan(&p.SkillLevel, &p.SkillStars)
		players = append(players, p)
	}
	pRows.Close()

	// Calculate average skill per team for scaling
	var teamASkill, teamBSkill float64
	var teamACount, teamBCount int
	var teamALevelTotal, teamBLevelTotal int
	for _, p := range players {
		sv := getSkillValue(p.SkillLevel, p.SkillStars)
		lv := skillLevelValue[p.SkillLevel]
		if p.Team == "A" {
			teamASkill += sv
			teamALevelTotal += lv
			teamACount++
		} else {
			teamBSkill += sv
			teamBLevelTotal += lv
			teamBCount++
		}
	}
	if teamACount > 0 {
		teamASkill /= float64(teamACount)
	}
	if teamBCount > 0 {
		teamBSkill /= float64(teamBCount)
	}
	teamAAvgLevel := 1
	if teamACount > 0 {
		teamAAvgLevel = teamALevelTotal / teamACount
	}
	teamBAvgLevel := 1
	if teamBCount > 0 {
		teamBAvgLevel = teamBLevelTotal / teamBCount
	}

	// Determine winner/loser avg skill for scaling
	var winnerAvgSkill, loserAvgSkill float64
	if winner == "A" {
		winnerAvgSkill, loserAvgSkill = teamASkill, teamBSkill
	} else if winner == "B" {
		winnerAvgSkill, loserAvgSkill = teamBSkill, teamASkill
	}
	expMult, rpLosePenalty := calcSkillScaling(winnerAvgSkill, loserAvgSkill)

	// Get court bonus EXP percent if match is in a court
	bonusExpPercent := 0
	if courtID.Valid {
		sqlDB.QueryRow("SELECT bonus_exp_percent FROM courts WHERE id = ?", courtID.String).Scan(&bonusExpPercent)
	}

	rewards := expRewards[matchMode]

	for _, player := range players {
		isWinner := player.Team == winner
		isDraw := winner == "draw"

		// Determine opponent team's avg level for skill gap check
		opponentAvgLevel := teamBAvgLevel
		if player.Team == "B" {
			opponentAvgLevel = teamAAvgLevel
		}
		playerTooHigh := isSkillGapTooLarge(player.SkillLevel, opponentAvgLevel)

		var expGained, rpGained, winInc, lossInc, drawInc int
		var starDelta int // for skill_test only

		if playerTooHigh {
			// Player is 2+ levels above opponent — gets NOTHING on win
			// On lose: penalty RP (ranked) or star loss (skill_test)
			if isDraw {
				drawInc = 1
			} else if isWinner {
				winInc = 1
				// No rewards at all
			} else {
				lossInc = 1
				if matchMode == "ranked" {
					rpGained = int(math.Round(float64(rankPointRewards["lose"]) * rpLosePenalty))
				}
				if matchMode == "skill_test" {
					starDelta = -1
				}
			}
		} else {
			// Normal reward logic
			if isDraw {
				expGained = rewards["draw"]
				drawInc = 1
				if matchMode == "ranked" {
					rpGained = rankPointRewards["draw"]
				}
			} else if isWinner {
				baseExp := rewards["win"]
				expGained = int(math.Round(float64(baseExp) * expMult))
				winInc = 1
				if matchMode == "ranked" {
					rpGained = rankPointRewards["win"]
				}
				if matchMode == "skill_test" {
					starDelta = 1
				}
			} else {
				expGained = rewards["lose"]
				lossInc = 1
				if matchMode == "ranked" {
					baseRP := rankPointRewards["lose"]
					rpGained = int(math.Round(float64(baseRP) * rpLosePenalty))
				}
				if matchMode == "skill_test" {
					starDelta = -1
				}
			}
		}

		// Apply court bonus EXP
		if bonusExpPercent > 0 && expGained > 0 {
			bonus := int(math.Round(float64(expGained) * float64(bonusExpPercent) / 100.0))
			expGained += bonus
		}

		tx.Exec("UPDATE match_players SET exp_gained = ?, rank_points_gained = ? WHERE id = ?",
			expGained, rpGained, player.ID)

		tx.Exec(`UPDATE users SET
			exp = exp + ?, wins = wins + ?, losses = losses + ?, draws = draws + ?,
			total_matches = total_matches + 1,
			rank_points = MAX(0, rank_points + ?),
			points = points + ?,
			updated_at = datetime('now')
			WHERE id = ?`,
			expGained, winInc, lossInc, drawInc, rpGained, expGained, player.UserID)

		// Handle skill star + promotion for skill_test matches
		if matchMode == "skill_test" && starDelta != 0 {
			var curStars int
			var curSkill string
			var promoWins, promoLosses int
			tx.QueryRow("SELECT skill_level, skill_stars, promo_wins, promo_losses FROM users WHERE id = ?",
				player.UserID).Scan(&curSkill, &curStars, &promoWins, &promoLosses)

			maxStars := tierStarsRequired[curSkill]
			inPromo := promoWins > 0 || promoLosses > 0

			if maxStars == 0 {
				// P+ (max tier) — stars still move 1-5 but no promotion
				newStars := curStars + starDelta
				if newStars < 1 {
					newStars = 1
				}
				if newStars > 5 {
					newStars = 5
				}
				tx.Exec("UPDATE users SET skill_stars = ? WHERE id = ?", newStars, player.UserID)
			} else if inPromo {
				// Player is in promotion Bo3
				if starDelta > 0 {
					promoWins++
				} else {
					promoLosses++
				}

				if promoWins >= 2 {
					// Promotion success!
					nextLevel := getNextSkillLevel(curSkill)
					if nextLevel != "" {
						tx.Exec("UPDATE users SET skill_level = ?, skill_stars = 1, promo_wins = 0, promo_losses = 0 WHERE id = ?",
							nextLevel, player.UserID)
					}
				} else if promoLosses >= 2 {
					// Promotion failed — drop 1 star, reset promo
					newStars := maxStars - 1
					if newStars < 1 {
						newStars = 1
					}
					tx.Exec("UPDATE users SET skill_stars = ?, promo_wins = 0, promo_losses = 0 WHERE id = ?",
						newStars, player.UserID)
				} else {
					// Still in promo
					tx.Exec("UPDATE users SET promo_wins = ?, promo_losses = ? WHERE id = ?",
						promoWins, promoLosses, player.UserID)
				}
			} else {
				// Normal star progression
				newStars := curStars + starDelta
				if newStars < 1 {
					newStars = 1
				}

				if newStars > maxStars {
					// Stars full — enter promotion! First promo win counted
					tx.Exec("UPDATE users SET skill_stars = ?, promo_wins = 1, promo_losses = 0 WHERE id = ?",
						maxStars, player.UserID)
				} else {
					tx.Exec("UPDATE users SET skill_stars = ? WHERE id = ?", newStars, player.UserID)
				}
			}
		}

		// Level up from EXP
		var level, exp, expToNext, userWins, totalMatches int
		tx.QueryRow("SELECT level, exp, exp_to_next_level, wins, total_matches FROM users WHERE id = ?",
			player.UserID).Scan(&level, &exp, &expToNext, &userWins, &totalMatches)

		for exp >= expToNext {
			exp -= expToNext
			level++
			expToNext = getExpForLevel(level)
		}
		newRank := getRankByLevel(level)
		winRate := 0.0
		if totalMatches > 0 {
			winRate = math.Round(float64(userWins)/float64(totalMatches)*1000) / 10
		}

		tx.Exec("UPDATE users SET level = ?, exp = ?, exp_to_next_level = ?, rank = ?, win_rate = ? WHERE id = ?",
			level, exp, expToNext, newRank, winRate, player.UserID)
	}

	tx.Commit()

	m, _ := buildMatchResponse(matchID)
	writeJSON(w, 200, m)
}

func handleLeaveMatch(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	matchID := chi.URLParam(r, "id")

	var status string
	var refereeID sql.NullString
	err := sqlDB.QueryRow("SELECT status, referee_id FROM matches WHERE id = ?", matchID).
		Scan(&status, &refereeID)
	if err != nil {
		writeError(w, 404, "ไม่พบห้อง")
		return
	}
	if status == "playing" {
		writeError(w, 400, "ไม่สามารถออกขณะกำลังแข่ง")
		return
	}

	if refereeID.Valid && refereeID.String == userID {
		sqlDB.Exec("UPDATE matches SET referee_id = NULL WHERE id = ?", matchID)
	} else {
		sqlDB.Exec("DELETE FROM match_players WHERE match_id = ? AND user_id = ?", matchID, userID)
	}

	var remaining int
	sqlDB.QueryRow("SELECT COUNT(*) FROM match_players WHERE match_id = ?", matchID).Scan(&remaining)
	if remaining == 0 {
		sqlDB.Exec("UPDATE matches SET status = 'cancelled' WHERE id = ?", matchID)
	}

	m, _ := buildMatchResponse(matchID)
	writeJSON(w, 200, m)
}

// handleUpdateSkillLevel allows leaders (assigned to the court) to update a player's skill level
// This is used after skill_test matches
func handleUpdateSkillLevel(w http.ResponseWriter, r *http.Request) {
	callerID := getUserID(r)
	callerRole := getUserRole(r)

	var req struct {
		UserID     string `json:"userId"`
		SkillLevel string `json:"skillLevel"`
		SkillStars int    `json:"skillStars"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, 400, "ข้อมูลไม่ถูกต้อง")
		return
	}

	if req.UserID == "" || req.SkillLevel == "" {
		writeError(w, 400, "กรุณาระบุผู้เล่นและระดับ")
		return
	}

	// Validate skill level
	if _, ok := skillLevelValue[req.SkillLevel]; !ok {
		writeError(w, 400, "ระดับฝีมือไม่ถูกต้อง (BG1, BG2, S, N, P-, P, P+)")
		return
	}
	if req.SkillStars < 1 || req.SkillStars > 5 {
		writeError(w, 400, "ดาวต้องอยู่ระหว่าง 1-5")
		return
	}

	// Admin can always update; leaders must be assigned to at least one court
	if callerRole != "admin" {
		if callerRole != "leader" {
			writeError(w, 403, "เฉพาะ admin หรือหัวก๊วนเท่านั้น")
			return
		}
		var courtCount int
		sqlDB.QueryRow("SELECT COUNT(*) FROM court_leaders WHERE user_id = ?", callerID).Scan(&courtCount)
		if courtCount == 0 {
			writeError(w, 403, "คุณไม่ได้เป็นหัวก๊วนของสนามใดๆ")
			return
		}
	}

	_, err := sqlDB.Exec("UPDATE users SET skill_level = ?, skill_stars = ?, updated_at = datetime('now') WHERE id = ?",
		req.SkillLevel, req.SkillStars, req.UserID)
	if err != nil {
		writeError(w, 500, "เกิดข้อผิดพลาด")
		return
	}

	writeJSON(w, 200, map[string]interface{}{
		"message":    "อัปเดตระดับฝีมือเรียบร้อย",
		"userId":     req.UserID,
		"skillLevel": req.SkillLevel,
		"skillStars": req.SkillStars,
	})
}
