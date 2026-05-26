package controller

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"drukarena/backend/model"
	"drukarena/utils/httpResp"

	"github.com/gorilla/mux"
)

// GetAllTournaments returns all tournaments
func GetAllTournaments(w http.ResponseWriter, r *http.Request) {
	_, role, authErr := GetSessionUser(r) //authentication check
	var tournaments []model.Tournament
	var err error
	if authErr == nil && role == "admin" && (r.URL.Query().Get("scope") == "admin" || r.URL.Path == "/api/admin/tournaments") {
		tournaments, err = model.GetAllTournamentsAdmin()
	} else {
		tournaments, err = model.GetAllTournaments()
	}
	if err != nil {
		httpResp.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	for i := range tournaments {
		tournaments[i].Status = computedTournamentStatus(tournaments[i])
	}
	httpResp.RespondWithJSON(w, http.StatusOK, tournaments)
}

// GetTournament returns a single tournament by ID
func GetTournament(w http.ResponseWriter, r *http.Request) {
	id, err := getTournamentID(mux.Vars(r)["id"])
	if err != nil {
		httpResp.RespondWithError(w, http.StatusBadRequest, "invalid tournament id")
		return
	}
	t := model.Tournament{TournamentID: id}
	if err := t.Read(); err != nil {
		httpResp.RespondWithError(w, http.StatusNotFound, "tournament not found")
		return
	}
	t.Status = computedTournamentStatus(t)
	httpResp.RespondWithJSON(w, http.StatusOK, t)
}

// CreateTournament lets players host tournaments. Admin approval is required
// before player-hosted tournaments appear publicly.
func CreateTournament(w http.ResponseWriter, r *http.Request) {
	userID, role, err := GetSessionUser(r)
	if err != nil {
		httpResp.RespondWithError(w, http.StatusUnauthorized, "login required")
		return
	}

	var t model.Tournament
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&t); err != nil {
		httpResp.RespondWithError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	defer r.Body.Close()
	if err := validateTournament(t); err != nil {
		httpResp.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	t.CreatedBy = userID
	t.Status = computedTournamentStatus(t)
	//business rule: tournaments created by admins are auto-approved, others require manual approval
	if role == "admin" {
		t.ApprovalStatus = "approved"
	} else {
		t.ApprovalStatus = "pending"
	}

	if err := t.Create(); err != nil {
		httpResp.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	httpResp.RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"status":        "tournament created",
		"tournament_id": t.TournamentID,
		"approval":      t.ApprovalStatus,
	})
}

// UpdateTournament modifies a tournament. Admins can edit any tournament; creators
// can edit their own tournament dates/details after logging in.
func UpdateTournament(w http.ResponseWriter, r *http.Request) {
	userID, role, err := GetSessionUser(r)
	if err != nil {
		httpResp.RespondWithError(w, http.StatusUnauthorized, "login required")
		return
	}

	id, err := getTournamentID(mux.Vars(r)["id"])
	if err != nil {
		httpResp.RespondWithError(w, http.StatusBadRequest, "invalid tournament id")
		return
	}
	existing := model.Tournament{TournamentID: id}
	if err := existing.Read(); err != nil {
		httpResp.RespondWithError(w, http.StatusNotFound, "tournament not found")
		return
	}
	if role != "admin" && existing.CreatedBy != userID {
		httpResp.RespondWithError(w, http.StatusUnauthorized, "only the creator or admin can update this tournament")
		return
	}

	var t model.Tournament
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&t); err != nil {
		httpResp.RespondWithError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	defer r.Body.Close()
	if err := validateTournament(t); err != nil {
		httpResp.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	t.Status = computedTournamentStatus(t)

	if err := t.Update(id); err != nil {
		httpResp.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpResp.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "tournament updated"})
}

func ApproveTournament(w http.ResponseWriter, r *http.Request) {
	setTournamentApproval(w, r, "approved")
}

func RejectTournament(w http.ResponseWriter, r *http.Request) {
	setTournamentApproval(w, r, "rejected")
}

func setTournamentApproval(w http.ResponseWriter, r *http.Request, approval string) {
	_, role, err := GetSessionUser(r)
	if err != nil || role != "admin" {
		httpResp.RespondWithError(w, http.StatusUnauthorized, "admin access required")
		return
	}
	id, err := getTournamentID(mux.Vars(r)["id"])
	if err != nil {
		httpResp.RespondWithError(w, http.StatusBadRequest, "invalid tournament id")
		return
	}
	if err := model.SetTournamentApproval(id, approval); err != nil {
		httpResp.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpResp.RespondWithJSON(w, http.StatusOK, map[string]string{"status": approval})
}

func JoinTournament(w http.ResponseWriter, r *http.Request) {
	userID, _, err := GetSessionUser(r)
	if err != nil {
		httpResp.RespondWithError(w, http.StatusUnauthorized, "login required")
		return
	}
	id, err := getTournamentID(mux.Vars(r)["id"])
	if err != nil {
		httpResp.RespondWithError(w, http.StatusBadRequest, "invalid tournament id")
		return
	}
	t := model.Tournament{TournamentID: id}
	if err := t.Read(); err != nil || t.ApprovalStatus != "approved" {
		httpResp.RespondWithError(w, http.StatusBadRequest, "tournament is not available to join")
		return
	}
	t.Status = computedTournamentStatus(t)
	if t.Status != "upcoming" {
		httpResp.RespondWithError(w, http.StatusBadRequest, "joining is closed for ongoing or completed tournaments")
		return
	}
	participants, err := model.CountTournamentParticipants(id)
	if err != nil {
		httpResp.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if t.MaxTeams > 0 && participants >= t.MaxTeams {
		httpResp.RespondWithError(w, http.StatusBadRequest, "this tournament is full")
		return
	}
	if err := model.JoinTournament(id, userID); err != nil {
		httpResp.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpResp.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "joined"})
}

func GetTournamentParticipants(w http.ResponseWriter, r *http.Request) {
	id, err := getTournamentID(mux.Vars(r)["id"])
	if err != nil {
		httpResp.RespondWithError(w, http.StatusBadRequest, "invalid tournament id")
		return
	}
	participants, err := model.GetTournamentParticipants(id)
	if err != nil {
		httpResp.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpResp.RespondWithJSON(w, http.StatusOK, participants)
}

// DeleteTournament removes a tournament (admin only)
func DeleteTournament(w http.ResponseWriter, r *http.Request) {
	_, role, err := GetSessionUser(r)
	if err != nil || role != "admin" {
		httpResp.RespondWithError(w, http.StatusUnauthorized, "admin access required")
		return
	}

	id, err := getTournamentID(mux.Vars(r)["id"])
	if err != nil {
		httpResp.RespondWithError(w, http.StatusBadRequest, "invalid tournament id")
		return
	}

	t := model.Tournament{TournamentID: id}
	if err := t.Delete(); err != nil {
		httpResp.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpResp.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func getTournamentID(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func validateTournament(t model.Tournament) error {
	if strings.TrimSpace(t.Title) == "" || strings.TrimSpace(t.GameTitle) == "" {
		return errString("title and game are required")
	}
	if !isAllowedGame(t.GameTitle) {
		return errString("please choose one of the supported games or sports")
	}
	start, startErr := parseTournamentDate(t.StartDate)
	end, endErr := parseTournamentDate(t.EndDate)
	if startErr != nil || endErr != nil {
		return errString("valid start and end date/time are required")
	}
	if !end.After(start) {
		return errString("end date/time must be after start date/time")
	}
	if t.MaxTeams < 1 {
		return errString("maximum players must be at least 1")
	}
	return nil
}

func isAllowedGame(game string) bool {
	allowed := map[string]bool{
		"Mobile Legends": true,
		"PUBG":           true,
		"Valorant":       true,
		"Football":       true,
		"Basketball":     true,
		"Volleyball":     true,
		"Badminton":      true,
	}
	return allowed[strings.TrimSpace(game)]
}

func computedTournamentStatus(t model.Tournament) string {
	start, startErr := parseTournamentDate(t.StartDate)
	end, endErr := parseTournamentDate(t.EndDate)
	if startErr != nil || endErr != nil {
		if t.Status == "" {
			return "upcoming"
		}
		return t.Status
	}
	now := time.Now()
	if now.Before(start) {
		return "upcoming"
	}
	if now.After(end) {
		return "completed"
	}
	return "ongoing"
}

func parseTournamentDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006-01-02 15:04:05-07",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	var lastErr error
	for _, layout := range layouts {
		parsed, err := time.ParseInLocation(layout, value, time.Local)
		if err == nil {
			if layout == "2006-01-02" {
				return parsed, nil
			}
			return parsed, nil
		}
		lastErr = err
	}
	return time.Time{}, lastErr
}

type errString string

func (e errString) Error() string { return string(e) }
