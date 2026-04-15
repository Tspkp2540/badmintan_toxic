package main

import (
	"database/sql"
	"math"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func handleGetRankings(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var totalPlayers int
	sqlDB.QueryRow("SELECT COUNT(*) FROM users WHERE total_matches > 0").Scan(&totalPlayers)

	rows, err := sqlDB.Query(`
		SELECT id, username, full_name, avatar_url, level, rank, skill_level, skill_stars, wins, losses, win_rate, total_matches, points, rank_points
		FROM users WHERE total_matches > 0
		ORDER BY rank_points DESC, wins DESC
		LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		writeError(w, 500, "เกิดข้อผิดพลาด")
		return
	}
	defer rows.Close()

	type rankEntry struct {
		Position     int     `json:"position"`
		UserID       string  `json:"userId"`
		Username     string  `json:"username"`
		FullName     string  `json:"fullName"`
		AvatarURL    *string `json:"avatarUrl"`
		Level        int     `json:"level"`
		Rank         string  `json:"rank"`
		SkillLevel   string  `json:"skillLevel"`
		SkillStars   int     `json:"skillStars"`
		Wins         int     `json:"wins"`
		Losses       int     `json:"losses"`
		WinRate      float64 `json:"winRate"`
		TotalMatches int     `json:"totalMatches"`
		Points       int     `json:"points"`
	}

	rankings := []rankEntry{}
	i := 0
	for rows.Next() {
		var e rankEntry
		var avatar sql.NullString
		var rankPoints int
		rows.Scan(&e.UserID, &e.Username, &e.FullName, &avatar, &e.Level, &e.Rank,
			&e.SkillLevel, &e.SkillStars, &e.Wins, &e.Losses, &e.WinRate, &e.TotalMatches, &e.Points, &rankPoints)
		e.Position = offset + i + 1
		if avatar.Valid {
			e.AvatarURL = &avatar.String
		}
		rankings = append(rankings, e)
		i++
	}

	writeJSON(w, 200, map[string]interface{}{
		"rankings":     rankings,
		"totalPlayers": totalPlayers,
		"currentPage":  page,
		"totalPages":   int(math.Ceil(float64(totalPlayers) / float64(limit))),
	})
}

func handleGetPlayerRank(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "userId")

	var (
		id, username, fullName, rank string
		avatar                       sql.NullString
		level, wins, losses, totalMatches, points, rankPoints int
		winRate                                                float64
	)
	err := sqlDB.QueryRow(`
		SELECT id, username, full_name, avatar_url, level, rank, wins, losses, win_rate, total_matches, points, rank_points
		FROM users WHERE id = ?`, userId).Scan(
		&id, &username, &fullName, &avatar, &level, &rank,
		&wins, &losses, &winRate, &totalMatches, &points, &rankPoints)
	if err != nil {
		writeError(w, 404, "ไม่พบผู้ใช้")
		return
	}

	var position int
	sqlDB.QueryRow(`
		SELECT COUNT(*) + 1 FROM users
		WHERE total_matches > 0 AND (rank_points > ? OR (rank_points = ? AND wins > ?))`,
		rankPoints, rankPoints, wins).Scan(&position)

	var avatarPtr *string
	if avatar.Valid {
		avatarPtr = &avatar.String
	}

	writeJSON(w, 200, map[string]interface{}{
		"position":     position,
		"userId":       id,
		"username":     username,
		"fullName":     fullName,
		"avatarUrl":    avatarPtr,
		"level":        level,
		"rank":         rank,
		"wins":         wins,
		"losses":       losses,
		"winRate":      winRate,
		"totalMatches": totalMatches,
		"points":       points,
	})
}
