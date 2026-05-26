package model

import (
	"drukarena/backend/dataStore/postgres"
)

// Match represents a scheduled or completed match
type Match struct {
	MatchID      int64  `json:"match_id"`
	TournamentID int64  `json:"tournament_id"`
	Team1ID      int64  `json:"team1_id"`
	Team2ID      int64  `json:"team2_id"`
	MatchDate    string `json:"match_date"`
	MatchTime    string `json:"match_time"`
	WinnerTeamID int64  `json:"winner_team_id"`
	ScoreTeam1   string `json:"score_team1"`
	ScoreTeam2   string `json:"score_team2"`
	Status       string `json:"status"`
	// Joined fields for display
	Team1Name string `json:"team1_name,omitempty"`
	Team2Name string `json:"team2_name,omitempty"`
}

const (
	queryInsertMatch   = `INSERT INTO matches(tournament_id, team1_id, team2_id, match_date, match_time, status) VALUES($1,$2,$3,$4,$5,$6) RETURNING match_id;`
	queryGetMatch      = `SELECT match_id, tournament_id, team1_id, team2_id, COALESCE(match_date::text,''), COALESCE(match_time::text,''), COALESCE(winner_team_id,0), COALESCE(score_team1,''), COALESCE(score_team2,''), status FROM matches WHERE match_id=$1;`
	queryGetAllMatches = `SELECT m.match_id, m.tournament_id, m.team1_id, m.team2_id, COALESCE(m.match_date::text,''), COALESCE(m.match_time::text,''), COALESCE(m.winner_team_id,0), COALESCE(m.score_team1,''), COALESCE(m.score_team2,''), m.status, COALESCE(t1.team_name,''), COALESCE(t2.team_name,'') FROM matches m LEFT JOIN teams t1 ON t1.team_id=m.team1_id LEFT JOIN teams t2 ON t2.team_id=m.team2_id ORDER BY m.match_date DESC;`
	queryUpdateMatch   = `UPDATE matches SET winner_team_id=$1, score_team1=$2, score_team2=$3, status=$4 WHERE match_id=$5 RETURNING match_id;`
	queryDeleteMatch   = `DELETE FROM matches WHERE match_id=$1 RETURNING match_id;`
)

// Create inserts a new match
func (m *Match) Create() error {
	return postgres.Db.QueryRow(queryInsertMatch,
		m.TournamentID, m.Team1ID, m.Team2ID, m.MatchDate, m.MatchTime, m.Status,
	).Scan(&m.MatchID)
}

// Read fetches a match by ID
func (m *Match) Read() error {
	return postgres.Db.QueryRow(queryGetMatch, m.MatchID).
		Scan(&m.MatchID, &m.TournamentID, &m.Team1ID, &m.Team2ID,
			&m.MatchDate, &m.MatchTime, &m.WinnerTeamID,
			&m.ScoreTeam1, &m.ScoreTeam2, &m.Status)
}

// Update saves match result / score
func (m *Match) Update() error {
	return postgres.Db.QueryRow(queryUpdateMatch,
		m.WinnerTeamID, m.ScoreTeam1, m.ScoreTeam2, m.Status, m.MatchID,
	).Scan(&m.MatchID)
}

// Delete removes a match
func (m *Match) Delete() error {
	_, err := postgres.Db.Exec(queryDeleteMatch, m.MatchID)
	return err
}

// GetAllMatches returns all matches with team names
func GetAllMatches() ([]Match, error) {
	rows, err := postgres.Db.Query(queryGetAllMatches)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	matches := []Match{}
	for rows.Next() {
		var m Match
		if err := rows.Scan(&m.MatchID, &m.TournamentID, &m.Team1ID, &m.Team2ID,
			&m.MatchDate, &m.MatchTime, &m.WinnerTeamID,
			&m.ScoreTeam1, &m.ScoreTeam2, &m.Status,
			&m.Team1Name, &m.Team2Name); err != nil {
			return nil, err
		}
		matches = append(matches, m)
	}
	return matches, nil
}
