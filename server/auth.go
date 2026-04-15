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
	Role           string
	AvatarURL      sql.NullString
	SkillLevel     string
	SkillStars     int
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
	PromoWins      int
	PromoLosses    int
	CreatedAt      string
	UpdatedAt      string
}

type userResponse struct {
	ID             string  `json:"id"`
	Username       string  `json:"username"`
	Email          string  `json:"email"`
	FullName       string  `json:"fullName"`
	Role           string  `json:"role"`
	AvatarURL      *string `json:"avatarUrl"`
	SkillLevel     string  `json:"skillLevel"`
	SkillStars     int     `json:"skillStars"`
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
	PromoWins      int     `json:"promoWins"`
	PromoLosses    int     `json:"promoLosses"`
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
		FullName: u.FullName, Role: u.Role, AvatarURL: avatar,
		SkillLevel: u.SkillLevel, SkillStars: u.SkillStars,
		Level: u.Level, Exp: u.Exp, ExpToNextLevel: u.ExpToNextLevel,
		Rank: u.Rank, Wins: u.Wins, Losses: u.Losses,
		TotalMatches: u.TotalMatches, WinRate: u.WinRate,
		Points: u.Points, RankPoints: u.RankPoints,
		PromoWins: u.PromoWins, PromoLosses: u.PromoLosses,
		CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt,
	}
}

const userSelectCols = `id, username, email, password_hash, full_name, role, avatar_url,
	skill_level, skill_stars, level, exp, exp_to_next_level, rank, wins, losses, draws, total_matches,
	win_rate, points, rank_points, promo_wins, promo_losses, created_at, updated_at`

func scanUserRow(row *sql.Row) (*userRow, error) {
	var u userRow
	err := row.Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.FullName, &u.Role, &u.AvatarURL,
		&u.SkillLevel, &u.SkillStars, &u.Level, &u.Exp, &u.ExpToNextLevel, &u.Rank,
		&u.Wins, &u.Losses, &u.Draws, &u.TotalMatches,
		&u.WinRate, &u.Points, &u.RankPoints, &u.PromoWins, &u.PromoLosses, &u.CreatedAt, &u.UpdatedAt,
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
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, 400, "ข้อมูลไม่ถูกต้อง")
		return
	}

	if req.Username == "" || req.Password == "" {
		writeError(w, 400, "กรุณากรอกชื่อผู้ใช้และรหัสผ่าน")
		return
	}

	user, err := scanUserRow(sqlDB.QueryRow(fmt.Sprintf("SELECT %s FROM users WHERE username = ?", userSelectCols), req.Username))
	if err != nil {
		writeError(w, 401, "ชื่อผู้ใช้หรือรหัสผ่านไม่ถูกต้อง")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		writeError(w, 401, "ชื่อผู้ใช้หรือรหัสผ่านไม่ถูกต้อง")
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

// handleGetUsers returns all users (admin/leader/vice_leader only)
func handleGetUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := sqlDB.Query(fmt.Sprintf("SELECT %s FROM users ORDER BY created_at DESC", userSelectCols))
	if err != nil {
		writeError(w, 500, "เกิดข้อผิดพลาด")
		return
	}
	defer rows.Close()

	var users []userResponse
	for rows.Next() {
		var u userRow
		err := rows.Scan(
			&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.FullName, &u.Role, &u.AvatarURL,
			&u.SkillLevel, &u.SkillStars, &u.Level, &u.Exp, &u.ExpToNextLevel, &u.Rank,
			&u.Wins, &u.Losses, &u.Draws, &u.TotalMatches,
			&u.WinRate, &u.Points, &u.RankPoints, &u.CreatedAt, &u.UpdatedAt,
		)
		if err != nil {
			continue
		}
		users = append(users, formatUserResp(&u))
	}
	writeJSON(w, 200, users)
}

// handleUpdateUserRole allows admin/leader to change a user's role
func handleUpdateUserRole(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"userId"`
		Role   string `json:"role"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, 400, "ข้อมูลไม่ถูกต้อง")
		return
	}

	validRoles := map[string]bool{"admin": true, "leader": true, "vice_leader": true, "player": true}
	if !validRoles[req.Role] {
		writeError(w, 400, "บทบาทไม่ถูกต้อง")
		return
	}

	if req.UserID == "" {
		writeError(w, 400, "กรุณาระบุผู้ใช้")
		return
	}

	callerRole := getUserRole(r)
	callerID := getUserID(r)

	// Prevent changing own role
	if req.UserID == callerID {
		writeError(w, 400, "ไม่สามารถเปลี่ยนบทบาทตัวเองได้")
		return
	}

	// Leader can only assign vice_leader/player, not admin/leader
	if callerRole == "leader" && (req.Role == "admin" || req.Role == "leader") {
		writeError(w, 403, "หัวหน้าไม่สามารถแต่งตั้ง admin หรือ leader ได้")
		return
	}

	// Vice leader cannot change roles
	if callerRole == "vice_leader" {
		writeError(w, 403, "รองหัวหน้าไม่สามารถเปลี่ยนบทบาทได้")
		return
	}

	_, err := sqlDB.Exec("UPDATE users SET role = ?, updated_at = datetime('now') WHERE id = ?", req.Role, req.UserID)
	if err != nil {
		writeError(w, 500, "เกิดข้อผิดพลาด")
		return
	}

	user, err := scanUserRow(sqlDB.QueryRow(fmt.Sprintf("SELECT %s FROM users WHERE id = ?", userSelectCols), req.UserID))
	if err != nil {
		writeError(w, 404, "ไม่พบผู้ใช้")
		return
	}
	writeJSON(w, 200, formatUserResp(user))
}
