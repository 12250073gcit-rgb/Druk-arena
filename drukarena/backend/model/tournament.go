package model

import (
	"database/sql"

	"drukarena/backend/dataStore/postgres"
)

// Tournament represents a tournament record
type Tournament struct {
	TournamentID     int64  `json:"tournament_id"`
	Title            string `json:"title"`
	GameTitle        string `json:"game_title"`
	Description      string `json:"description"`
	StartDate        string `json:"start_date"`
	EndDate          string `json:"end_date"`
	Venue            string `json:"venue"`
	MaxTeams         int    `json:"max_teams"`
	PrizePool        string `json:"prize_pool"`
	Status           string `json:"status"`
	ApprovalStatus   string `json:"approval_status"`
	ImageURL         string `json:"image_url"`
	CreatedBy        int64  `json:"created_by"`
	CreatedAt        string `json:"created_at,omitempty"`
	ParticipantCount int    `json:"participant_count"`
}

type TournamentParticipant struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	JoinedAt string `json:"joined_at,omitempty"`
}

const (
	queryInsertTournament = `INSERT INTO tournaments(title, game_title, description, start_date, end_date, venue, max_teams, prize_pool, status, approval_status, image_url, created_by)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) RETURNING tournament_id;`
	queryGetTournament          = `SELECT tournament_id, title, game_title, COALESCE(description,''), start_date::text, end_date::text, COALESCE(venue,''), max_teams, COALESCE(prize_pool,''), status, COALESCE(approval_status,'approved'), COALESCE(image_url,''), COALESCE(created_by,0), (SELECT COUNT(*) FROM tournament_players tp WHERE tp.tournament_id=tournaments.tournament_id) FROM tournaments WHERE tournament_id=$1;`
	queryGetAllTournaments      = `SELECT tournament_id, title, game_title, COALESCE(description,''), start_date::text, end_date::text, COALESCE(venue,''), max_teams, COALESCE(prize_pool,''), status, COALESCE(approval_status,'approved'), COALESCE(image_url,''), COALESCE(created_by,0), (SELECT COUNT(*) FROM tournament_players tp WHERE tp.tournament_id=tournaments.tournament_id) FROM tournaments WHERE COALESCE(approval_status,'approved')='approved' ORDER BY start_date DESC;`
	queryGetAllTournamentsAdmin = `SELECT tournament_id, title, game_title, COALESCE(description,''), start_date::text, end_date::text, COALESCE(venue,''), max_teams, COALESCE(prize_pool,''), status, COALESCE(approval_status,'approved'), COALESCE(image_url,''), COALESCE(created_by,0), (SELECT COUNT(*) FROM tournament_players tp WHERE tp.tournament_id=tournaments.tournament_id) FROM tournaments ORDER BY created_at DESC;`
	queryUpdateTournament       = `UPDATE tournaments SET title=$1, game_title=$2, description=$3, start_date=$4, end_date=$5, venue=$6, max_teams=$7, prize_pool=$8, status=$9, image_url=$10 WHERE tournament_id=$11 RETURNING tournament_id;`
	queryDeleteTournament       = `DELETE FROM tournaments WHERE tournament_id=$1 RETURNING tournament_id;`
	queryCountTournaments       = `SELECT COUNT(*) FROM tournaments;`
	queryJoinTournament         = `INSERT INTO tournament_players(tournament_id, user_id) VALUES($1,$2) ON CONFLICT(tournament_id, user_id) DO NOTHING RETURNING registration_id;`
)

// Create inserts a new tournament
func (t *Tournament) Create() error {
	return postgres.Db.QueryRow(queryInsertTournament,
		t.Title, t.GameTitle, t.Description, t.StartDate, t.EndDate,
		t.Venue, t.MaxTeams, t.PrizePool, t.Status, t.ApprovalStatus, t.ImageURL, t.CreatedBy,
	).Scan(&t.TournamentID)
}

// Read fetches a single tournament by ID
func (t *Tournament) Read() error {
	return postgres.Db.QueryRow(queryGetTournament, t.TournamentID).
		Scan(&t.TournamentID, &t.Title, &t.GameTitle, &t.Description,
			&t.StartDate, &t.EndDate, &t.Venue, &t.MaxTeams,
			&t.PrizePool, &t.Status, &t.ApprovalStatus, &t.ImageURL, &t.CreatedBy, &t.ParticipantCount)
}

// Update modifies a tournament record
func (t *Tournament) Update(oldID int64) error {
	return postgres.Db.QueryRow(queryUpdateTournament,
		t.Title, t.GameTitle, t.Description, t.StartDate, t.EndDate,
		t.Venue, t.MaxTeams, t.PrizePool, t.Status, t.ImageURL, oldID,
	).Scan(&t.TournamentID)
}

// Delete removes a tournament
func (t *Tournament) Delete() error {
	_, err := postgres.Db.Exec(queryDeleteTournament, t.TournamentID)
	return err
}

func GetAllTournaments() ([]Tournament, error) {
	return scanTournaments(queryGetAllTournaments)
}

func GetAllTournamentsAdmin() ([]Tournament, error) {
	return scanTournaments(queryGetAllTournamentsAdmin)
}

func scanTournaments(query string) ([]Tournament, error) {
	rows, err := postgres.Db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tournaments := []Tournament{}
	for rows.Next() {
		var t Tournament
		if err := rows.Scan(&t.TournamentID, &t.Title, &t.GameTitle, &t.Description,
			&t.StartDate, &t.EndDate, &t.Venue, &t.MaxTeams,
			&t.PrizePool, &t.Status, &t.ApprovalStatus, &t.ImageURL, &t.CreatedBy, &t.ParticipantCount); err != nil {
			return nil, err
		}
		tournaments = append(tournaments, t)
	}
	return tournaments, nil
}

func SetTournamentApproval(id int64, approval string) error {
	_, err := postgres.Db.Exec(`UPDATE tournaments SET approval_status=$1 WHERE tournament_id=$2`, approval, id)
	return err
}

func JoinTournament(tournamentID, userID int64) error {
	var registrationID int64
	err := postgres.Db.QueryRow(queryJoinTournament, tournamentID, userID).Scan(&registrationID)
	if err == sql.ErrNoRows {
		return nil
	}
	return err
}

func CountTournamentParticipants(tournamentID int64) (int, error) {
	var count int
	err := postgres.Db.QueryRow(`SELECT COUNT(*) FROM tournament_players WHERE tournament_id=$1`, tournamentID).Scan(&count)
	return count, err
}

func GetTournamentParticipants(tournamentID int64) ([]TournamentParticipant, error) {
	rows, err := postgres.Db.Query(
		`SELECT u.user_id, u.username, ''
		 FROM tournament_players tp
		 JOIN users u ON u.user_id=tp.user_id
		 WHERE tp.tournament_id=$1
		 ORDER BY tp.joined_at ASC`,
		tournamentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	participants := []TournamentParticipant{}
	for rows.Next() {
		var p TournamentParticipant
		if err := rows.Scan(&p.UserID, &p.Username, &p.JoinedAt); err != nil {
			return nil, err
		}
		participants = append(participants, p)
	}
	return participants, nil
}

// CountTournaments returns total tournament count
func CountTournaments() (int, error) {
	var count int
	err := postgres.Db.QueryRow(queryCountTournaments).Scan(&count)
	return count, err
}
