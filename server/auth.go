package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type userRow struct {
	ID             string
	Username       string
	Email          string
	PasswordHash   string
	FullName       string
	AvatarURL      sql.NullString
	Level          int
	Exp            int
	ExpToNextLevel int
	Rank           string
	Wins           int
	Losses         int
	Draws          int
	TotalMatches   int
	WinRate        float64
	Points         int
	RankPoints     int
	CreatedAt      string
	UpdatedAt      string
}

type userResponse struct {
	ID             string  `json:"id"`
	Username       string  `json:"username"`
	Email          string  `json:"email"`
	FullName       string  `json:"fullName"`
	AvatarURL      *string `json:"avatarUrl"`
	Level          int     `json:"level"`
	Exp            int     `json:"exp"`
	ExpToNextLevel int     `json:"expToNextLevel"`
	Rank           string  `json:"rank"`
	Wins           int     `json:"wins"`
	Losses         int     `json:"losses"`
	TotalMatches   int     `json:"totalMatches"`
	WinRate        float64 `json:"winRate"`
	Points         int     `json:"points"`
	RankPoints     int     `json:"rankPoints"`
	CreatedAt      string  `json:"createdAt"`
	UpdatedAt      string  `json:"updatedAt"`
}

func formatUserResp(u *userRow) userResponse {
	var avatar *string
	if u.AvatarURL.Valid {
		avatar = &u.AvatarURL.String
	}
	return userResponse{
		ID: u.ID, Username: u.Username, Email: u.Email,
		FullName: u.FullName, AvatarURL: avatar,
		Level: u.Level, Exp: u.Exp, ExpToNextLevel: u.ExpToNextLevel,
		Rank: u.Rank, Wins: u.Wins, Losses: u.Losses,
		TotalMatches: u.TotalMatches, WinRate: u.WinRate,
		Points: u.Points, RankPoints: u.RankPoints,
		CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt,
	}
}

const userSelectCols = `id, username, email, password_hash, full_name, avatar_url,
	level, exp, exp_to_next_level, rank, wins, losses, draws, total_matches,
	win_rate, points, rank_points, created_at, updated_at`

func scanUserRow(row *sql.Row) (*userRow, error) {
	var u userRow
	err := row.Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.FullName, &u.AvatarURL,
		&u.Level, &u.Exp, &u.ExpToNextLevel, &u.Rank,
		&u.Wins, &u.Losses, &u.Draws, &u.TotalMatches,
		&u.WinRate, &u.Points, &u.RankPoints, &u.CreatedAt, &u.UpdatedAt,
	)
	return &u, err
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
		FullName string `json:"fullName"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, 400, "ข้อมูลไม่ถูกต้อง")
		return
	}

	if req.Username == "" || req.Email == "" || req.Password == "" || req.FullName == "" {
		writeError(w, 400, "กรุณากรอกข้อมูลให้ครบ")
		return
	}
	if len(req.Password) < 6 {
		writeError(w, 400, "รหัสผ่านต้องมีอย่างน้อย 6 ตัวอักษร")
		return
	}

	var exists string
	err := sqlDB.QueryRow("SELECT id FROM users WHERE email = ? OR username = ?", req.Email, req.Username).Scan(&exists)
	if err == nil {
		writeError(w, 409, "อีเมลหรือชื่อผู้ใช้นี้ถูกใช้แล้ว")
		return
	}

	id := uuid.New().String()
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		writeError(w, 500, "เกิดข้อผิดพลาด")
		return
	}
	expToNext := getExpForLevel(1)

	_, err = sqlDB.Exec(`INSERT INTO users (id, username, email, password_hash, full_name, exp_to_next_level)
		VALUES (?, ?, ?, ?, ?, ?)`, id, req.Username, req.Email, string(hash), req.FullName, expToNext)
	if err != nil {
		writeError(w, 500, "เกิดข้อผิดพลาด")
		return
	}

	user, err := scanUserRow(sqlDB.QueryRow(fmt.Sprintf("SELECT %s FROM users WHERE id = ?", userSelectCols), id))
	if err != nil {
		writeError(w, 500, "เกิดข้อผิดพลาด")
		return
	}

	token, _ := signToken(id, req.Email)
	writeJSON(w, 201, map[string]interface{}{"token": token, "user": formatUserResp(user)})
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, 400, "ข้อมูลไม่ถูกต้อง")
		return
	}

	if req.Email == "" || req.Password == "" {
		writeError(w, 400, "กรุณากรอกอีเมลและรหัสผ่าน")
		return
	}

	user, err := scanUserRow(sqlDB.QueryRow(fmt.Sprintf("SELECT %s FROM users WHERE email = ?", userSelectCols), req.Email))
	if err != nil {
		writeError(w, 401, "อีเมลหรือรหัสผ่านไม่ถูกต้อง")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		writeError(w, 401, "อีเมลหรือรหัสผ่านไม่ถูกต้อง")
		return
	}

	token, _ := signToken(user.ID, user.Email)
	writeJSON(w, 200, map[string]interface{}{"token": token, "user": formatUserResp(user)})
}

func handleGetProfile(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	user, err := scanUserRow(sqlDB.QueryRow(fmt.Sprintf("SELECT %s FROM users WHERE id = ?", userSelectCols), userID))
	if err != nil {
		writeError(w, 404, "ไม่พบผู้ใช้")
		return
	}
	writeJSON(w, 200, formatUserResp(user))
}

func handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	var req struct {
		Username  *string `json:"username"`
		FullName  *string `json:"fullName"`
		AvatarURL *string `json:"avatarUrl"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, 400, "ข้อมูลไม่ถูกต้อง")
		return
	}

	var updates []string
	var values []interface{}

	if req.Username != nil {
		var exists string
		err := sqlDB.QueryRow("SELECT id FROM users WHERE username = ? AND id != ?", *req.Username, userID).Scan(&exists)
		if err == nil {
			writeError(w, 409, "ชื่อผู้ใช้นี้ถูกใช้แล้ว")
			return
		}
		updates = append(updates, "username = ?")
		values = append(values, *req.Username)
	}
	if req.FullName != nil {
		updates = append(updates, "full_name = ?")
		values = append(values, *req.FullName)
	}
	if req.AvatarURL != nil {
		updates = append(updates, "avatar_url = ?")
		values = append(values, *req.AvatarURL)
	}

	if len(updates) == 0 {
		writeError(w, 400, "ไม่มีข้อมูลที่จะอัปเดต")
		return
	}

	updates = append(updates, "updated_at = datetime('now')")
	values = append(values, userID)

	query := fmt.Sprintf("UPDATE users SET %s WHERE id = ?", strings.Join(updates, ", "))
	sqlDB.Exec(query, values...)

	user, _ := scanUserRow(sqlDB.QueryRow(fmt.Sprintf("SELECT %s FROM users WHERE id = ?", userSelectCols), userID))
	writeJSON(w, 200, formatUserResp(user))
}
