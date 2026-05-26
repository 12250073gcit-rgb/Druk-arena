package model

import (
	"database/sql"

	"drukarena/backend/dataStore/postgres"
)

// User represents a registered user (admin or player)
type User struct {
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password,omitempty"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at,omitempty"`
}

const (
	queryInsertUser     = `INSERT INTO users(username, email, password, role) VALUES($1, $2, $3, $4) RETURNING user_id;`
	queryGetUserByEmail = `SELECT user_id, username, email, password, role FROM users WHERE email=$1;`
	queryGetUserByID    = `SELECT user_id, username, email, role FROM users WHERE user_id=$1;`
	queryGetAllUsers    = `SELECT user_id, username, email, role, created_at FROM users ORDER BY created_at DESC;`
	queryCountUsers     = `SELECT COUNT(*) FROM users;`
)

// Create inserts a new user into the database
func (u *User) Create() error {
	err := postgres.Db.QueryRow(queryInsertUser,
		u.Username, u.Email, u.Password, u.Role,
	).Scan(&u.UserID)
	return err
}

// GetByEmail fetches a user by email (used for login)
func (u *User) GetByEmail() error {
	return postgres.Db.QueryRow(queryGetUserByEmail, u.Email).
		Scan(&u.UserID, &u.Username, &u.Email, &u.Password, &u.Role)
}

// GetByID fetches a user by ID
func (u *User) GetByID() error {
	return postgres.Db.QueryRow(queryGetUserByID, u.UserID).
		Scan(&u.UserID, &u.Username, &u.Email, &u.Role)
}

// GetAllUsers returns all users
func GetAllUsers() ([]User, error) {
	rows, err := postgres.Db.Query(queryGetAllUsers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.UserID, &u.Username, &u.Email, &u.Role, &u.CreatedAt); err != nil {
			return nil, err
		}
		u.Password = "" // never expose password
		users = append(users, u)
	}
	return users, nil
}

// CountUsers returns total user count
func CountUsers() (int, error) {
	var count int
	err := postgres.Db.QueryRow(queryCountUsers).Scan(&count)
	return count, err
}

// IsFirstUser checks if no users exist yet (first signup = admin)
func IsFirstUser() (bool, error) {
	count, err := CountUsers()
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return count == 0, nil
}

func EnsureAdminUser(email, passwordHash string) error {
	var id int64
	err := postgres.Db.QueryRow(`SELECT user_id FROM users WHERE email=$1`, email).Scan(&id)
	if err == nil {
		_, err = postgres.Db.Exec(`UPDATE users SET password=$1, role='admin' WHERE email=$2`, passwordHash, email)
		return err
	}
	if err != sql.ErrNoRows {
		return err
	}

	_, err = postgres.Db.Exec(
		`INSERT INTO users(username, email, password, role) VALUES($1,$2,$3,'admin')`,
		"DrukArena Admin", email, passwordHash,
	)
	return err
}

func DeleteUser(userID int64) error {
	tx, err := postgres.Db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	statements := []string{
		`DELETE FROM tournament_players WHERE user_id=$1`,
		`DELETE FROM team_members WHERE user_id=$1`,
		`UPDATE gallery_uploads SET uploaded_by=NULL WHERE uploaded_by=$1`,
		`UPDATE teams SET captain_id=NULL WHERE captain_id=$1`,
		`DELETE FROM player_profiles WHERE user_id=$1`,
		`DELETE FROM users WHERE user_id=$1`,
	}
	for _, stmt := range statements {
		if _, err = tx.Exec(stmt, userID); err != nil {
			return err
		}
	}
	return tx.Commit()
}
