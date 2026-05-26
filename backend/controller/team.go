package controller

import (
	"encoding/json"
	"net/http"
	"strconv"

	"drukarena/backend/model"
	"drukarena/utils/httpResp"

	"github.com/gorilla/mux"
)

// CreateTeam creates a new team for a tournament
func CreateTeam(w http.ResponseWriter, r *http.Request) {
	userID, _, err := GetSessionUser(r)
	if err != nil {
		httpResp.RespondWithError(w, http.StatusUnauthorized, "login required")
		return
	}

	var t model.Team
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&t); err != nil {
		httpResp.RespondWithError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	defer r.Body.Close()
	t.CaptainID = userID

	if err := t.Create(); err != nil {
		httpResp.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Auto-add captain as a member
	m := model.TeamMember{TeamID: t.TeamID, UserID: userID, Role: "captain"}
	_ = m.AddMember()

	httpResp.RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"status":  "team created",
		"team_id": t.TeamID,
	})
}

// GetTeam returns a single team with its members
func GetTeam(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		httpResp.RespondWithError(w, http.StatusBadRequest, "invalid team id")
		return
	}

	t := model.Team{TeamID: id}
	if err := t.Read(); err != nil {
		httpResp.RespondWithError(w, http.StatusNotFound, "team not found")
		return
	}

	members, _ := model.GetTeamMembers(id)
	httpResp.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"team":    t,
		"members": members,
	})
}

// GetTeamsByTournament returns all teams for a tournament
func GetTeamsByTournament(w http.ResponseWriter, r *http.Request) {
	tid, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		httpResp.RespondWithError(w, http.StatusBadRequest, "invalid tournament id")
		return
	}
	teams, err := model.GetTeamsByTournament(tid)
	if err != nil {
		httpResp.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpResp.RespondWithJSON(w, http.StatusOK, teams)
}

// AddTeamMember adds a user to a team
func AddTeamMember(w http.ResponseWriter, r *http.Request) {
	userID, _, err := GetSessionUser(r)
	if err != nil {
		httpResp.RespondWithError(w, http.StatusUnauthorized, "login required")
		return
	}

	var m model.TeamMember
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		httpResp.RespondWithError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	defer r.Body.Close()
	_ = userID
	m.Role = "player"

	if err := m.AddMember(); err != nil {
		httpResp.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	httpResp.RespondWithJSON(w, http.StatusCreated, map[string]string{"status": "member added"})
}

// RemoveTeamMember removes a member from a team (captain or admin only)
func RemoveTeamMember(w http.ResponseWriter, r *http.Request) {
	_, _, err := GetSessionUser(r)
	if err != nil {
		httpResp.RespondWithError(w, http.StatusUnauthorized, "login required")
		return
	}

	memberID, _ := strconv.ParseInt(mux.Vars(r)["memberid"], 10, 64)
	teamID, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)

	if err := model.RemoveMember(memberID, teamID); err != nil {
		httpResp.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpResp.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "member removed"})
}

// DeleteTeam deletes a team (admin only)
func DeleteTeam(w http.ResponseWriter, r *http.Request) {
	_, role, err := GetSessionUser(r)
	if err != nil || role != "admin" {
		httpResp.RespondWithError(w, http.StatusUnauthorized, "admin access required")
		return
	}
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	t := model.Team{TeamID: id}
	if err := t.Delete(); err != nil {
		httpResp.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpResp.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
