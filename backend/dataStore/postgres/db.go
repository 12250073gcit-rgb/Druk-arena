package postgres

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/lib/pq"
)

// Db is exported so all model packages can use it
var Db *sql.DB

// init() runs automatically before main() — establishes DB connection
func init() {
	// Read from environment variables (fly.io injects these)
	// Fall back to devcontainer defaults
	host := getEnv("POSTGRES_HOST", "localhost")
	port := getEnv("POSTGRES_PORT", "5432")
	user := getEnv("POSTGRES_USER", "postgres")
	password := getEnv("POSTGRES_PASSWORD", "postgres")
	dbname := getEnv("POSTGRES_DBNAME", "drukarena_db")

	// DATABASE_URL overrides individual params on hosted providers like Render.
	databaseURL := os.Getenv("DATABASE_URL")

	var err error
	if databaseURL != "" {
		Db, err = sql.Open("postgres", normalizeDatabaseURL(databaseURL))
	} else {
		if err = ensureLocalDatabase(host, port, user, password, dbname); err != nil {
			log.Fatal("Cannot create or verify database: ", err)
		}

		// devcontainer / local: build connection string from parts
		db_info := fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			host, port, user, password, dbname,
		)
		Db, err = sql.Open("postgres", db_info)
	}

	if err != nil {
		panic(err)
	}

	if err = pingWithRetry(Db, 10, 2*time.Second); err != nil {
		log.Fatal("Cannot connect to database: ", err)
	}

	if err = ensureSchema(); err != nil {
		log.Fatal("Cannot initialize database schema: ", err)
	}

	log.Println("Database successfully connected")
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func normalizeDatabaseURL(databaseURL string) string {
	parsed, err := url.Parse(databaseURL)
	if err != nil || parsed.Hostname() == "" {
		return databaseURL
	}

	query := parsed.Query()
	if query.Get("sslmode") != "" {
		return databaseURL
	}

	host := strings.ToLower(parsed.Hostname())
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		query.Set("sslmode", "disable")
	} else {
		query.Set("sslmode", "require")
	}
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func pingWithRetry(db *sql.DB, attempts int, delay time.Duration) error {
	var err error
	for i := 1; i <= attempts; i++ {
		err = db.Ping()
		if err == nil {
			return nil
		}
		if i < attempts {
			log.Printf("Database ping failed, retrying (%d/%d): %v", i, attempts, err)
			time.Sleep(delay)
		}
	}
	return err
}

func ensureLocalDatabase(host, port, user, password, dbname string) error {
	maintenanceInfo := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		host, port, user, password,
	)

	maintenanceDb, err := sql.Open("postgres", maintenanceInfo)
	if err != nil {
		return err
	}
	defer maintenanceDb.Close()

	if err = maintenanceDb.Ping(); err != nil {
		return err
	}

	var exists bool
	err = maintenanceDb.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname=$1)", dbname).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	_, err = maintenanceDb.Exec("CREATE DATABASE " + pq.QuoteIdentifier(dbname))
	return err
}

func ensureSchema() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS users (
			user_id BIGSERIAL PRIMARY KEY,
			username TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'user',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS player_profiles (
			profile_id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL UNIQUE REFERENCES users(user_id) ON DELETE CASCADE,
			display_name TEXT,
			bio TEXT,
			favorite_game TEXT,
			matches_played INTEGER NOT NULL DEFAULT 0,
			wins INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS tournaments (
			tournament_id BIGSERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			game_title TEXT NOT NULL,
			description TEXT,
			start_date DATE NOT NULL,
			end_date DATE NOT NULL,
			venue TEXT,
			max_teams INTEGER NOT NULL DEFAULT 0,
			prize_pool TEXT,
			status TEXT NOT NULL DEFAULT 'upcoming',
			approval_status TEXT NOT NULL DEFAULT 'approved',
			image_url TEXT,
			created_by BIGINT REFERENCES users(user_id) ON DELETE SET NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS tournament_players (
			registration_id BIGSERIAL PRIMARY KEY,
			tournament_id BIGINT NOT NULL REFERENCES tournaments(tournament_id) ON DELETE CASCADE,
			user_id BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(tournament_id, user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS teams (
			team_id BIGSERIAL PRIMARY KEY,
			team_name TEXT NOT NULL,
			captain_id BIGINT REFERENCES users(user_id) ON DELETE SET NULL,
			tournament_id BIGINT NOT NULL REFERENCES tournaments(tournament_id) ON DELETE CASCADE,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS team_members (
			member_id BIGSERIAL PRIMARY KEY,
			team_id BIGINT NOT NULL REFERENCES teams(team_id) ON DELETE CASCADE,
			user_id BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			role TEXT NOT NULL DEFAULT 'member',
			joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(team_id, user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS matches (
			match_id BIGSERIAL PRIMARY KEY,
			tournament_id BIGINT NOT NULL REFERENCES tournaments(tournament_id) ON DELETE CASCADE,
			team1_id BIGINT REFERENCES teams(team_id) ON DELETE SET NULL,
			team2_id BIGINT REFERENCES teams(team_id) ON DELETE SET NULL,
			match_date DATE,
			match_time TIME,
			winner_team_id BIGINT REFERENCES teams(team_id) ON DELETE SET NULL,
			score_team1 TEXT,
			score_team2 TEXT,
			status TEXT NOT NULL DEFAULT 'scheduled',
			stats_recorded BOOLEAN NOT NULL DEFAULT false,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS news (
			news_id BIGSERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			image_url TEXT,
			published_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			author_id BIGINT REFERENCES users(user_id) ON DELETE SET NULL
		)`,
		`CREATE TABLE IF NOT EXISTS gallery_uploads (
			gallery_id BIGSERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			image_url TEXT NOT NULL,
			uploaded_by BIGINT REFERENCES users(user_id) ON DELETE SET NULL,
			uploader_name TEXT,
			uploader_email TEXT,
			approval_status TEXT NOT NULL DEFAULT 'approved',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS blocked_emails (
			blocked_id BIGSERIAL PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			reason TEXT,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
		`CREATE INDEX IF NOT EXISTS idx_player_profiles_user_id ON player_profiles(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_tournaments_approval_status ON tournaments(approval_status)`,
		`CREATE INDEX IF NOT EXISTS idx_tournament_players_tournament_id ON tournament_players(tournament_id)`,
		`CREATE INDEX IF NOT EXISTS idx_tournament_players_user_id ON tournament_players(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_teams_tournament_id ON teams(tournament_id)`,
		`CREATE INDEX IF NOT EXISTS idx_team_members_team_id ON team_members(team_id)`,
		`CREATE INDEX IF NOT EXISTS idx_team_members_user_id ON team_members(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_matches_tournament_id ON matches(tournament_id)`,
		`CREATE INDEX IF NOT EXISTS idx_news_published_date ON news(published_date)`,
		`CREATE INDEX IF NOT EXISTS idx_gallery_uploads_approval_status ON gallery_uploads(approval_status)`,
		`CREATE INDEX IF NOT EXISTS idx_gallery_uploads_uploaded_by ON gallery_uploads(uploaded_by)`,
	}

	for _, stmt := range statements {
		if _, err := Db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}
