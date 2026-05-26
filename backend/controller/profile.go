package controller

import (
	"encoding/json"
	"net/http"

	"drukarena/backend/model"
	"drukarena/utils/httpResp"
)

func GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, _, err := GetSessionUser(r)
	if err != nil {
		httpResp.RespondWithError(w, http.StatusUnauthorized, "login required")
		return
	}
	u := model.User{UserID: userID}
	if err := u.GetByID(); err != nil {
		httpResp.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := model.EnsurePlayerProfile(userID, u.Username); err != nil {
		httpResp.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	profile, err := model.GetPlayerProfile(userID)
	if err != nil {
		httpResp.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpResp.RespondWithJSON(w, http.StatusOK, profile)
}

func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, _, err := GetSessionUser(r)
	if err != nil {
		httpResp.RespondWithError(w, http.StatusUnauthorized, "login required")
		return
	}
	var payload struct {
		DisplayName  string `json:"display_name"`
		Bio          string `json:"bio"`
		FavoriteGame string `json:"favorite_game"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpResp.RespondWithError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	defer r.Body.Close()
	if err := model.UpdatePlayerProfile(userID, payload.DisplayName, payload.Bio, payload.FavoriteGame); err != nil {
		httpResp.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpResp.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "profile updated"})
}
