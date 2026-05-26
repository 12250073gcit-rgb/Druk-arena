package routes

import (
	"log"
	"net/http"
	"os"

	"drukarena/backend/controller"

	"github.com/gorilla/mux"
)

// InitializeRoutes sets up all API routes and starts the HTTP server
func InitializeRoutes() {
	port := getEnv("PORT", "8080")
	router := mux.NewRouter()

	// Serve static frontend files
	router.PathPrefix("/view/").Handler(http.StripPrefix("/view/", http.FileServer(http.Dir("./view/"))))
	router.PathPrefix("/css/").Handler(http.StripPrefix("/css/", http.FileServer(http.Dir("./view/css/"))))
	router.PathPrefix("/js/").Handler(http.StripPrefix("/js/", http.FileServer(http.Dir("./view/js/"))))
	router.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir(getEnv("UPLOAD_DIR", "./view/uploads")))))
	router.Handle("/logo.png", http.FileServer(http.Dir("./view/")))
	router.Handle("/hero-bg.png", http.FileServer(http.Dir("./view/")))

	//Serve HTML pages
	router.HandleFunc("/", serveFile("./view/index.html")).Methods("GET")
	router.HandleFunc("/login", serveFile("./view/login.html")).Methods("GET")
	router.HandleFunc("/signup", serveFile("./view/signup.html")).Methods("GET")
	router.HandleFunc("/tournaments", serveFile("./view/tournaments.html")).Methods("GET")
	router.HandleFunc("/tournament/{id}", serveFile("./view/tournament-detail.html")).Methods("GET")
	router.HandleFunc("/create-tournament", serveFile("./view/create-tournament.html")).Methods("GET")
	router.HandleFunc("/teams", serveFile("./view/teams.html")).Methods("GET")
	router.HandleFunc("/matches", serveFile("./view/matches.html")).Methods("GET")
	router.HandleFunc("/news", serveFile("./view/news.html")).Methods("GET")
	router.HandleFunc("/profile", serveFile("./view/profile.html")).Methods("GET")
	router.HandleFunc("/gallery", serveFile("./view/gallery.html")).Methods("GET")
	router.HandleFunc("/admin", serveFile("./view/admin-dashboard.html")).Methods("GET")

	//Auth API
	router.HandleFunc("/signup", controller.Signup).Methods("POST")
	router.HandleFunc("/login", controller.Login).Methods("POST")
	router.HandleFunc("/logout", controller.Logout).Methods("POST")
	router.HandleFunc("/verify", controller.VerifySession).Methods("GET")

	//Tournaments API
	router.HandleFunc("/api/tournaments", controller.GetAllTournaments).Methods("GET")
	router.HandleFunc("/api/tournament/{id}", controller.GetTournament).Methods("GET")
	router.HandleFunc("/api/tournament", controller.CreateTournament).Methods("POST")
	router.HandleFunc("/api/tournament/{id}", controller.UpdateTournament).Methods("PUT")
	router.HandleFunc("/api/tournament/{id}", controller.DeleteTournament).Methods("DELETE")
	router.HandleFunc("/api/tournament/{id}/approve", controller.ApproveTournament).Methods("POST")
	router.HandleFunc("/api/tournament/{id}/reject", controller.RejectTournament).Methods("POST")
	router.HandleFunc("/api/tournament/{id}/join", controller.JoinTournament).Methods("POST")
	router.HandleFunc("/api/tournament/{id}/participants", controller.GetTournamentParticipants).Methods("GET")

	//Teams API
	router.HandleFunc("/api/team", controller.CreateTeam).Methods("POST")
	router.HandleFunc("/api/team/{id}", controller.GetTeam).Methods("GET")
	router.HandleFunc("/api/team/{id}", controller.DeleteTeam).Methods("DELETE")
	router.HandleFunc("/api/tournament/{id}/teams", controller.GetTeamsByTournament).Methods("GET")
	router.HandleFunc("/api/team/{id}/member", controller.AddTeamMember).Methods("POST")
	router.HandleFunc("/api/team/{id}/member/{memberid}", controller.RemoveTeamMember).Methods("DELETE")

	//Matches API
	router.HandleFunc("/api/matches", controller.GetAllMatches).Methods("GET")
	router.HandleFunc("/api/match", controller.CreateMatch).Methods("POST")
	router.HandleFunc("/api/match/{id}", controller.UpdateMatch).Methods("PUT")
	router.HandleFunc("/api/match/{id}", controller.DeleteMatch).Methods("DELETE")

	// News API
	router.HandleFunc("/api/news", controller.GetAllNews).Methods("GET")
	router.HandleFunc("/api/news", controller.CreateNews).Methods("POST")
	router.HandleFunc("/api/news/{id}", controller.DeleteNews).Methods("DELETE")

	//Player Profile API
	router.HandleFunc("/api/profile", controller.GetProfile).Methods("GET")
	router.HandleFunc("/api/profile", controller.UpdateProfile).Methods("PUT")

	//Gallery API
	router.HandleFunc("/api/gallery", controller.GetGallery).Methods("GET")
	router.HandleFunc("/api/gallery", controller.UploadGallery).Methods("POST")
	router.HandleFunc("/api/admin/gallery", controller.GetAdminGallery).Methods("GET")
	router.HandleFunc("/api/admin/gallery/{id}/approve", controller.ApproveGallery).Methods("POST")
	router.HandleFunc("/api/admin/gallery/{id}/reject", controller.RejectGallery).Methods("POST")
	router.HandleFunc("/api/admin/gallery/{id}", controller.DeleteGallery).Methods("DELETE")
	router.HandleFunc("/api/admin/gallery/{id}/block-uploader", controller.BlockGalleryUploader).Methods("POST")

	//Admin API
	router.HandleFunc("/api/admin/stats", controller.GetAdminStats).Methods("GET")
	router.HandleFunc("/api/admin/users", controller.GetAllUsersAdmin).Methods("GET")
	router.HandleFunc("/api/admin/users/{id}", controller.DeleteUserAdmin).Methods("DELETE")
	router.HandleFunc("/api/admin/tournaments", controller.GetAllTournaments).Methods("GET")

	log.Printf("DrukArena running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

// serveFile returns an http.HandlerFunc that serves a specific HTML file
func serveFile(filename string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filename)
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
