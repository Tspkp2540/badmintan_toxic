package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	initDB()
	migrateDB()

	r := chi.NewRouter()

	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Route("/api", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, req *http.Request) {
			writeJSON(w, 200, map[string]string{
				"status":    "ok",
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			})
		})

		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", handleRegister)
			r.Post("/login", handleLogin)
			r.With(authMiddleware).Get("/profile", handleGetProfile)
			r.With(authMiddleware).Put("/profile", handleUpdateProfile)
		})

		// User management (admin/leader only)
		r.Route("/users", func(r chi.Router) {
			r.Use(authMiddleware)
			r.With(requireRole("admin", "leader", "vice_leader")).Get("/", handleGetUsers)
			r.With(requireRole("admin", "leader")).Put("/role", handleUpdateUserRole)
		})

		// Courts (server browser)
		r.Route("/courts", func(r chi.Router) {
			r.Get("/", handleGetCourts)                   // public: browse courts
			r.Get("/{id}", handleGetCourt)                // public: court detail
			r.Get("/{id}/matches", handleGetCourtMatches) // public: matches in a court

			r.Group(func(r chi.Router) {
				r.Use(authMiddleware)
				// Leader can update bonus EXP on assigned courts
				r.With(requireRole("admin", "leader")).Put("/{id}/bonus", handleUpdateCourtBonus)

				r.With(requireRole("admin")).Post("/", handleCreateCourt)       // admin only: create court
				r.With(requireRole("admin")).Put("/{id}", handleUpdateCourt)    // admin only: update court
				r.With(requireRole("admin")).Delete("/{id}", handleDeleteCourt) // admin only: close court

				// Court leader management (admin only)
				r.With(requireRole("admin")).Post("/{id}/leaders", handleAssignCourtLeader)
				r.With(requireRole("admin")).Delete("/{id}/leaders", handleRemoveCourtLeader)
			})
		})

		r.Route("/rankings", func(r chi.Router) {
			r.Get("/", handleGetRankings)
			r.Get("/{userId}", handleGetPlayerRank)
		})

		r.Route("/matches", func(r chi.Router) {
			r.Use(authMiddleware)
			r.Get("/", handleGetMatches)
			r.Post("/", handleCreateMatch)
			r.Get("/{id}", handleGetMatch)
			r.Post("/{id}/join", handleJoinMatch)
			r.Post("/{id}/referee", handleJoinAsReferee)
			r.Post("/{id}/start", handleStartMatch)
			r.Post("/{id}/end", handleEndMatch)
			r.Post("/{id}/score", handleSubmitScores)
			r.Post("/{id}/leave", handleLeaveMatch)
			r.With(requireRole("admin", "leader")).Put("/skill-level", handleUpdateSkillLevel)
		})

		// My court leaderships
		r.With(authMiddleware).With(requireRole("admin", "leader", "vice_leader")).Get("/my-courts", handleGetMyCourtLeaderships)
	})

	// Serve static frontend in production
	if os.Getenv("NODE_ENV") == "production" {
		distPath := filepath.Join(".", "dist")
		if info, err := os.Stat(distPath); err == nil && info.IsDir() {
			fs := http.FileServer(http.Dir(distPath))
			r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				fpath := filepath.Join(distPath, req.URL.Path)
				if _, err := os.Stat(fpath); os.IsNotExist(err) {
					http.ServeFile(w, req, filepath.Join(distPath, "index.html"))
					return
				}
				fs.ServeHTTP(w, req)
			}))
		}
	}

	mode := "development"
	if os.Getenv("NODE_ENV") == "production" {
		mode = "production"
	}
	fmt.Printf("🏸 Badminton Hub server running on port %s\n", port)
	fmt.Printf("   Mode: %s\n", mode)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, r))
}
