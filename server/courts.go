package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type courtResponse struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	Description     string          `json:"description"`
	Location        string          `json:"location"`
	MaxRooms        int             `json:"maxRooms"`
	BonusExpPercent int             `json:"bonusExpPercent"`
	Status          string          `json:"status"`
	CreatedBy       string          `json:"createdBy"`
	CreatorName     string          `json:"creatorName"`
	ActiveRooms     int             `json:"activeRooms"`
	TotalPlayers    int             `json:"totalPlayers"`
	Leaders         []courtLeaderResp `json:"leaders"`
	CreatedAt       string          `json:"createdAt"`
	UpdatedAt       string          `json:"updatedAt"`
}

type courtLeaderResp struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	FullName string `json:"fullName"`
}

func buildCourtResponse(courtID string) (*courtResponse, error) {
	var c courtResponse
	var desc, loc sql.NullString

	err := sqlDB.QueryRow(`SELECT c.id, c.name, c.description, c.location, c.max_rooms, c.bonus_exp_percent, c.status,
		c.created_by, u.full_name, c.created_at, c.updated_at
		FROM courts c JOIN users u ON u.id = c.created_by
		WHERE c.id = ?`, courtID).Scan(
		&c.ID, &c.Name, &desc, &loc, &c.MaxRooms, &c.BonusExpPercent, &c.Status,
		&c.CreatedBy, &c.CreatorName, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if desc.Valid {
		c.Description = desc.String
	}
	if loc.Valid {
		c.Location = loc.String
	}

	// Count active rooms in this court
	sqlDB.QueryRow(`SELECT COUNT(*) FROM matches WHERE court_id = ? AND status IN ('waiting', 'playing', 'scoring')`,
		courtID).Scan(&c.ActiveRooms)

	// Count total players in active rooms
	sqlDB.QueryRow(`SELECT COUNT(*) FROM match_players mp
		JOIN matches m ON m.id = mp.match_id
		WHERE m.court_id = ? AND m.status IN ('waiting', 'playing', 'scoring')`,
		courtID).Scan(&c.TotalPlayers)

	// Load assigned leaders
	c.Leaders = []courtLeaderResp{}
	lRows, _ := sqlDB.Query(`SELECT cl.user_id, u.username, u.full_name
		FROM court_leaders cl JOIN users u ON u.id = cl.user_id
		WHERE cl.court_id = ?`, courtID)
	if lRows != nil {
		defer lRows.Close()
		for lRows.Next() {
			var l courtLeaderResp
			lRows.Scan(&l.UserID, &l.Username, &l.FullName)
			c.Leaders = append(c.Leaders, l)
		}
	}

	return &c, nil
}

// handleGetCourts returns all courts (public, no auth required)
func handleGetCourts(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	var query string
	var args []interface{}

	if status != "" {
		query = "SELECT id FROM courts WHERE status = ? ORDER BY created_at DESC"
		args = append(args, status)
	} else {
		query = "SELECT id FROM courts ORDER BY created_at DESC"
	}

	rows, err := sqlDB.Query(query, args...)
	if err != nil {
		writeError(w, 500, "เกิดข้อผิดพลาด")
		return
	}
	defer rows.Close()

	courts := []courtResponse{}
	for rows.Next() {
		var id string
		rows.Scan(&id)
		c, err := buildCourtResponse(id)
		if err == nil {
			courts = append(courts, *c)
		}
	}
	writeJSON(w, 200, courts)
}

// handleGetCourt returns a single court by ID (public)
func handleGetCourt(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	c, err := buildCourtResponse(id)
	if err != nil {
		writeError(w, 404, "ไม่พบสนาม")
		return
	}
	writeJSON(w, 200, c)
}

// handleCreateCourt creates a new court (admin only)
func handleCreateCourt(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Location    string `json:"location"`
		MaxRooms    int    `json:"maxRooms"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, 400, "ข้อมูลไม่ถูกต้อง")
		return
	}

	if req.Name == "" {
		writeError(w, 400, "กรุณาตั้งชื่อสนาม")
		return
	}
	if req.MaxRooms <= 0 {
		req.MaxRooms = 10
	}

	id := uuid.New().String()
	_, err := sqlDB.Exec(`INSERT INTO courts (id, name, description, location, max_rooms, created_by)
		VALUES (?, ?, ?, ?, ?, ?)`, id, req.Name, req.Description, req.Location, req.MaxRooms, userID)
	if err != nil {
		writeError(w, 500, "เกิดข้อผิดพลาด")
		return
	}

	c, _ := buildCourtResponse(id)
	writeJSON(w, 201, c)
}

// handleUpdateCourt updates a court (admin only)
func handleUpdateCourt(w http.ResponseWriter, r *http.Request) {
	courtID := chi.URLParam(r, "id")
	var req struct {
		Name            *string `json:"name"`
		Description     *string `json:"description"`
		Location        *string `json:"location"`
		MaxRooms        *int    `json:"maxRooms"`
		BonusExpPercent *int    `json:"bonusExpPercent"`
		Status          *string `json:"status"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, 400, "ข้อมูลไม่ถูกต้อง")
		return
	}

	updates := []string{}
	values := []interface{}{}

	if req.Name != nil {
		updates = append(updates, "name = ?")
		values = append(values, *req.Name)
	}
	if req.Description != nil {
		updates = append(updates, "description = ?")
		values = append(values, *req.Description)
	}
	if req.Location != nil {
		updates = append(updates, "location = ?")
		values = append(values, *req.Location)
	}
	if req.MaxRooms != nil {
		updates = append(updates, "max_rooms = ?")
		values = append(values, *req.MaxRooms)
	}
	if req.BonusExpPercent != nil {
		updates = append(updates, "bonus_exp_percent = ?")
		values = append(values, *req.BonusExpPercent)
	}
	if req.Status != nil {
		validStatuses := map[string]bool{"open": true, "closed": true, "maintenance": true}
		if !validStatuses[*req.Status] {
			writeError(w, 400, "สถานะไม่ถูกต้อง")
			return
		}
		updates = append(updates, "status = ?")
		values = append(values, *req.Status)
	}

	if len(updates) == 0 {
		writeError(w, 400, "ไม่มีข้อมูลที่จะอัปเดต")
		return
	}

	updates = append(updates, "updated_at = datetime('now')")
	values = append(values, courtID)

	query := fmt.Sprintf("UPDATE courts SET %s WHERE id = ?", joinStrings(updates, ", "))
	_, err := sqlDB.Exec(query, values...)
	if err != nil {
		writeError(w, 500, "เกิดข้อผิดพลาด")
		return
	}

	c, _ := buildCourtResponse(courtID)
	writeJSON(w, 200, c)
}

// handleDeleteCourt soft-deletes a court by setting status to closed (admin only)
func handleDeleteCourt(w http.ResponseWriter, r *http.Request) {
	courtID := chi.URLParam(r, "id")

	// Check if there are active matches
	var activeCount int
	sqlDB.QueryRow(`SELECT COUNT(*) FROM matches WHERE court_id = ? AND status IN ('waiting', 'playing', 'scoring')`,
		courtID).Scan(&activeCount)
	if activeCount > 0 {
		writeError(w, 400, "ไม่สามารถปิดสนามได้ ยังมีห้องที่ใช้งานอยู่")
		return
	}

	sqlDB.Exec("UPDATE courts SET status = 'closed', updated_at = datetime('now') WHERE id = ?", courtID)
	writeJSON(w, 200, map[string]string{"message": "ปิดสนามเรียบร้อย"})
}

// handleGetCourtMatches returns matches inside a specific court
func handleGetCourtMatches(w http.ResponseWriter, r *http.Request) {
	courtID := chi.URLParam(r, "id")
	status := r.URL.Query().Get("status")

	var query string
	var args []interface{}

	if status != "" {
		query = "SELECT id FROM matches WHERE court_id = ? AND status = ? ORDER BY created_at DESC"
		args = []interface{}{courtID, status}
	} else {
		query = "SELECT id FROM matches WHERE court_id = ? AND status IN ('waiting', 'playing', 'scoring') ORDER BY created_at DESC"
		args = []interface{}{courtID}
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

func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

// handleAssignCourtLeader assigns a leader to a specific court (admin only)
func handleAssignCourtLeader(w http.ResponseWriter, r *http.Request) {
	courtID := chi.URLParam(r, "id")

	var req struct {
		UserID string `json:"userId"`
	}
	if err := readJSON(r, &req); err != nil || req.UserID == "" {
		writeError(w, 400, "กรุณาระบุผู้ใช้")
		return
	}

	// Verify user exists and is a leader
	var userRole string
	err := sqlDB.QueryRow("SELECT role FROM users WHERE id = ?", req.UserID).Scan(&userRole)
	if err != nil {
		writeError(w, 404, "ไม่พบผู้ใช้")
		return
	}
	if userRole != "leader" && userRole != "vice_leader" {
		writeError(w, 400, "ผู้ใช้ต้องมีบทบาท leader หรือ vice_leader")
		return
	}

	// Verify court exists
	var courtExists string
	if sqlDB.QueryRow("SELECT id FROM courts WHERE id = ?", courtID).Scan(&courtExists) != nil {
		writeError(w, 404, "ไม่พบสนาม")
		return
	}

	_, err = sqlDB.Exec("INSERT INTO court_leaders (court_id, user_id) VALUES (?, ?)", courtID, req.UserID)
	if err != nil {
		writeError(w, 409, "ผู้ใช้นี้เป็นหัวก๊วนของสนามนี้อยู่แล้ว")
		return
	}

	c, _ := buildCourtResponse(courtID)
	writeJSON(w, 200, c)
}

// handleRemoveCourtLeader removes a leader from a specific court (admin only)
func handleRemoveCourtLeader(w http.ResponseWriter, r *http.Request) {
	courtID := chi.URLParam(r, "id")

	var req struct {
		UserID string `json:"userId"`
	}
	if err := readJSON(r, &req); err != nil || req.UserID == "" {
		writeError(w, 400, "กรุณาระบุผู้ใช้")
		return
	}

	result, err := sqlDB.Exec("DELETE FROM court_leaders WHERE court_id = ? AND user_id = ?", courtID, req.UserID)
	if err != nil {
		writeError(w, 500, "เกิดข้อผิดพลาด")
		return
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		writeError(w, 404, "ไม่พบหัวก๊วนในสนามนี้")
		return
	}

	c, _ := buildCourtResponse(courtID)
	writeJSON(w, 200, c)
}

// handleUpdateCourtBonus allows leaders assigned to a court to update its bonus EXP
func handleUpdateCourtBonus(w http.ResponseWriter, r *http.Request) {
	courtID := chi.URLParam(r, "id")
	callerID := getUserID(r)
	callerRole := getUserRole(r)

	var req struct {
		BonusExpPercent int `json:"bonusExpPercent"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, 400, "ข้อมูลไม่ถูกต้อง")
		return
	}

	if req.BonusExpPercent < 0 || req.BonusExpPercent > 100 {
		writeError(w, 400, "bonus ต้องอยู่ระหว่าง 0-100%")
		return
	}

	// Admin can always update; leaders must be assigned to this court
	if callerRole != "admin" {
		var isAssigned int
		sqlDB.QueryRow("SELECT COUNT(*) FROM court_leaders WHERE court_id = ? AND user_id = ?",
			courtID, callerID).Scan(&isAssigned)
		if isAssigned == 0 {
			writeError(w, 403, "คุณไม่ได้เป็นหัวก๊วนของสนามนี้")
			return
		}
	}

	sqlDB.Exec("UPDATE courts SET bonus_exp_percent = ?, updated_at = datetime('now') WHERE id = ?",
		req.BonusExpPercent, courtID)

	c, _ := buildCourtResponse(courtID)
	writeJSON(w, 200, c)
}

// handleGetMyCourtLeaderships returns courts where the current user is assigned as leader
func handleGetMyCourtLeaderships(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	rows, err := sqlDB.Query("SELECT court_id FROM court_leaders WHERE user_id = ?", userID)
	if err != nil {
		writeError(w, 500, "เกิดข้อผิดพลาด")
		return
	}
	defer rows.Close()

	courts := []courtResponse{}
	for rows.Next() {
		var courtID string
		rows.Scan(&courtID)
		c, err := buildCourtResponse(courtID)
		if err == nil {
			courts = append(courts, *c)
		}
	}
	writeJSON(w, 200, courts)
}
