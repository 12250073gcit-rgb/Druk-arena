package controller

import (
	"net/http"
	"strconv"

	"drukarena/backend/model"
	"drukarena/utils/httpResp"

	"github.com/gorilla/mux"
)

// GetAdminStats returns overview counts for the admin dashboard
func GetAdminStats(w http.ResponseWriter, r *http.Request) {
	_, role, err := GetSessionUser(r)
	if err != nil || role != "admin" {
		httpResp.RespondWithError(w, http.StatusUnauthorized, "admin access required")
		return
	}

	totalTournaments, _ := model.CountTournaments()
	totalTeams, _ := model.CountTeams()
	totalUsers, _ := model.CountUsers()

	httpResp.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"total_tournaments": totalTournaments,
		"total_teams":       totalTeams,
		"total_users":       totalUsers,
	})
}

// GetAllUsersAdmin returns all users (admin only)
func GetAllUsersAdmin(w http.ResponseWriter, r *http.Request) {
	_, role, err := GetSessionUser(r)
	if err != nil || role != "admin" {
		httpResp.RespondWithError(w, http.StatusUnauthorized, "admin access required")
		return
	}
	users, err := model.GetAllUsers()
	if err != nil {
		httpResp.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpResp.RespondWithJSON(w, http.StatusOK, users)
}

// DeleteUserAdmin removes a non-admin user and their direct profile/registration links.
func DeleteUserAdmin(w http.ResponseWriter, r *http.Request) {
	adminID, role, err := GetSessionUser(r)
	if err != nil || role != "admin" {
		httpResp.RespondWithError(w, http.StatusUnauthorized, "admin access required")
		return
	}
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		httpResp.RespondWithError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	if id == adminID {
		httpResp.RespondWithError(w, http.StatusBadRequest, "admin cannot remove their own account")
		return
	}
	u := model.User{UserID: id}
	if err := u.GetByID(); err != nil {
		httpResp.RespondWithError(w, http.StatusNotFound, "user not found")
		return
	}
	if u.Role == "admin" {
		httpResp.RespondWithError(w, http.StatusBadRequest, "admin users cannot be removed here")
		return
	}
	if err := model.DeleteUser(id); err != nil {
		httpResp.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpResp.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "user removed"})
}
