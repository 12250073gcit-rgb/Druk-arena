package model

import (
	"drukarena/backend/dataStore/postgres"
)

// Team represents a tournament team
type Team struct {
	TeamID       int64  `json:"team_id"`
	TeamName     string `json:"team_name"`
	CaptainID    int64  `json:"captain_id"`
	TournamentID int64  `json:"tournament_id"`
	CreatedAt    string `json:"created_at,omitempty"`
}

// TeamMember represents a team member entry
type TeamMember struct {
	MemberID int64  `json:"member_id"`
	TeamID   int64  `json:"team_id"`
	UserID   int64  `json:"user_id"`
	Username string `json:"username,omitempty"`
	Role     string `json:"role"`
	JoinedAt string `json:"joined_at,omitempty"`
}

const (
	queryInsertTeam           = `INSERT INTO teams(team_name, captain_id, tournament_id) VALUES($1,$2,$3) RETURNING team_id;`
	queryGetTeam              = `SELECT team_id, team_name, captain_id, tournament_id FROM teams WHERE team_id=$1;`
	queryGetTeamsByTournament = `SELECT team_id, team_name, captain_id, tournament_id FROM teams WHERE tournament_id=$1;`
	queryDeleteTeam           = `DELETE FROM teams WHERE team_id=$1 RETURNING team_id;`
	queryCountTeams           = `SELECT COUNT(*) FROM teams;`
	queryInsertMember         = `INSERT INTO team_members(team_id, user_id, role) VALUES($1,$2,$3) RETURNING member_id;`
	queryGetTeamMembers       = `SELECT tm.member_id, tm.team_id, tm.user_id, u.username, tm.role FROM team_members tm JOIN users u ON u.user_id=tm.user_id WHERE tm.team_id=$1;`
	queryDeleteMember         = `DELETE FROM team_members WHERE member_id=$1 AND team_id=$2 RETURNING member_id;`
)

// Create inserts a new team
func (t *Team) Create() error {
	return postgres.Db.QueryRow(queryInsertTeam, t.TeamName, t.CaptainID, t.TournamentID).
		Scan(&t.TeamID)
}

// Read fetches a team by ID
func (t *Team) Read() error {
	return postgres.Db.QueryRow(queryGetTeam, t.TeamID).
		Scan(&t.TeamID, &t.TeamName, &t.CaptainID, &t.TournamentID)
}

// Delete removes a team
func (t *Team) Delete() error {
	_, err := postgres.Db.Exec(queryDeleteTeam, t.TeamID)
	return err
}

// GetTeamsByTournament returns all teams for a tournament
func GetTeamsByTournament(tournamentID int64) ([]Team, error) {
	rows, err := postgres.Db.Query(queryGetTeamsByTournament, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teams := []Team{}
	for rows.Next() {
		var t Team
		if err := rows.Scan(&t.TeamID, &t.TeamName, &t.CaptainID, &t.TournamentID); err != nil {
			return nil, err
		}
		teams = append(teams, t)
	}
	return teams, nil
}

// CountTeams returns total team count
func CountTeams() (int, error) {
	var count int
	err := postgres.Db.QueryRow(queryCountTeams).Scan(&count)
	return count, err
}

// AddMember adds a user to a team
func (m *TeamMember) AddMember() error {
	return postgres.Db.QueryRow(queryInsertMember, m.TeamID, m.UserID, m.Role).
		Scan(&m.MemberID)
}

// GetTeamMembers returns all members of a team
func GetTeamMembers(teamID int64) ([]TeamMember, error) {
	rows, err := postgres.Db.Query(queryGetTeamMembers, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := []TeamMember{}
	for rows.Next() {
		var m TeamMember
		if err := rows.Scan(&m.MemberID, &m.TeamID, &m.UserID, &m.Username, &m.Role); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, nil
}

// RemoveMember removes a member from a team
func RemoveMember(memberID, teamID int64) error {
	_, err := postgres.Db.Exec(queryDeleteMember, memberID, teamID)
	return err
}
