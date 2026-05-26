package model

import "drukarena/backend/dataStore/postgres"

type PlayerProfile struct {
	ProfileID     int64  `json:"profile_id"`
	UserID        int64  `json:"user_id"`
	Username      string `json:"username,omitempty"`
	DisplayName   string `json:"display_name"`
	Bio           string `json:"bio"`
	FavoriteGame  string `json:"favorite_game"`
	MatchesPlayed int    `json:"matches_played"`
	Wins          int    `json:"wins"`
}

func EnsurePlayerProfile(userID int64, username string) error {
	_, err := postgres.Db.Exec(
		`INSERT INTO player_profiles(user_id, display_name)
		 VALUES($1,$2)
		 ON CONFLICT(user_id) DO NOTHING`,
		userID, username,
	)
	return err
}

func GetPlayerProfile(userID int64) (PlayerProfile, error) {
	var p PlayerProfile
	err := postgres.Db.QueryRow(
		`SELECT pp.profile_id, pp.user_id, u.username, COALESCE(pp.display_name,u.username),
		        COALESCE(pp.bio,''), COALESCE(pp.favorite_game,''), pp.matches_played, pp.wins
		 FROM player_profiles pp
		 JOIN users u ON u.user_id=pp.user_id
		 WHERE pp.user_id=$1`,
		userID,
	).Scan(&p.ProfileID, &p.UserID, &p.Username, &p.DisplayName, &p.Bio, &p.FavoriteGame, &p.MatchesPlayed, &p.Wins)
	return p, err
}

func UpdatePlayerProfile(userID int64, displayName, bio, favoriteGame string) error {
	_, err := postgres.Db.Exec(
		`UPDATE player_profiles
		 SET display_name=$1, bio=$2, favorite_game=$3, updated_at=CURRENT_TIMESTAMP
		 WHERE user_id=$4`,
		displayName, bio, favoriteGame, userID,
	)
	return err
}

func RecordMatchStats(matchID, team1ID, team2ID, winnerTeamID int64) error {
	tx, err := postgres.Db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var alreadyRecorded bool
	if err = tx.QueryRow(`SELECT COALESCE(stats_recorded,false) FROM matches WHERE match_id=$1 FOR UPDATE`, matchID).Scan(&alreadyRecorded); err != nil {
		return err
	}
	if alreadyRecorded {
		return tx.Commit()
	}

	if _, err = tx.Exec(
		`UPDATE player_profiles pp
		 SET matches_played = matches_played + 1, updated_at=CURRENT_TIMESTAMP
		 WHERE pp.user_id IN (
		   SELECT user_id FROM team_members WHERE team_id IN ($1,$2)
		   UNION
		   SELECT captain_id FROM teams WHERE team_id IN ($1,$2) AND captain_id IS NOT NULL
		 )`,
		team1ID, team2ID,
	); err != nil {
		return err
	}

	if winnerTeamID > 0 {
		if _, err = tx.Exec(
			`UPDATE player_profiles pp
			 SET wins = wins + 1, updated_at=CURRENT_TIMESTAMP
			 WHERE pp.user_id IN (
			   SELECT user_id FROM team_members WHERE team_id=$1
			   UNION
			   SELECT captain_id FROM teams WHERE team_id=$1 AND captain_id IS NOT NULL
			 )`,
			winnerTeamID,
		); err != nil {
			return err
		}
	}

	if _, err = tx.Exec(`UPDATE matches SET stats_recorded=true WHERE match_id=$1`, matchID); err != nil {
		return err
	}
	return tx.Commit()
}
