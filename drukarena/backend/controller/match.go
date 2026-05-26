package controller

import (
	"encoding/json"
	"net/http"
	"strconv"

	"drukarena/backend/model"
	"drukarena/utils/httpResp"

	"github.com/gorilla/mux"
)

// GetAllMatches returns all matches
func GetAllMatches(w http.ResponseWriter, r *http.Request) {
	matches, err := model.GetAllMatches()
	if err != nil {
		httpResp.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpResp.RespondWithJSON(w, http.StatusOK, matches)
}

// CreateMatch creates a new match (admin only)
func CreateMatch(w http.ResponseWriter, r *http.Request) {
	_, role, err := GetSessionUser(r)
	if err != nil || role != "admin" {
		httpResp.RespondWithError(w, http.StatusUnauthorized, "admin access required")
		return
	}

	var m model.Match
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		httpResp.RespondWithError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	defer r.Body.Close()

	if m.Status == "" {
		m.Status = "scheduled"
	}

	if err := m.Create(); err != nil {
		httpResp.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	httpResp.RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"status":   "match created",
		"match_id": m.MatchID,
	})
}

// UpdateMatch updates score/result (admin only)
func UpdateMatch(w http.ResponseWriter, r *http.Request) {
	_, role, err := GetSessionUser(r)
	if err != nil || role != "admin" {
		httpResp.RespondWithError(w, http.StatusUnauthorized, "admin access required")
		return
	}

	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		httpResp.RespondWithError(w, http.StatusBadRequest, "invalid match id")
		return
	}

	var m model.Match
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		httpResp.RespondWithError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	defer r.Body.Close()
	m.MatchID = id

	if err := m.Update(); err != nil {
		httpResp.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if m.Status == "completed" {
		saved := model.Match{MatchID: id}
		if err := saved.Read(); err == nil {
			_ = model.RecordMatchStats(saved.MatchID, saved.Team1ID, saved.Team2ID, m.WinnerTeamID)
		}
	}
	httpResp.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "match updated"})
}

// DeleteMatch removes a match (admin only)
func DeleteMatch(w http.ResponseWriter, r *http.Request) {
	_, role, err := GetSessionUser(r)
	if err != nil || role != "admin" {
		httpResp.RespondWithError(w, http.StatusUnauthorized, "admin access required")
		return
	}
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	m := model.Match{MatchID: id}
	if err := m.Delete(); err != nil {
		httpResp.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpResp.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
