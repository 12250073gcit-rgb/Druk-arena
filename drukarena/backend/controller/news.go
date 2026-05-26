package controller

import (
	"encoding/json"
	"net/http"
	"strconv"

	"drukarena/backend/model"
	"drukarena/utils/httpResp"

	"github.com/gorilla/mux"
)

// GetAllNews returns all news articles
func GetAllNews(w http.ResponseWriter, r *http.Request) {
	newsList, err := model.GetAllNews()
	if err != nil {
		httpResp.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpResp.RespondWithJSON(w, http.StatusOK, newsList)
}

// CreateNews adds a news article (admin only)
func CreateNews(w http.ResponseWriter, r *http.Request) {
	userID, role, err := GetSessionUser(r)
	if err != nil || role != "admin" {
		httpResp.RespondWithError(w, http.StatusUnauthorized, "admin access required")
		return
	}

	var n model.News
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&n); err != nil {
		httpResp.RespondWithError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	defer r.Body.Close()
	n.AuthorID = userID

	if err := n.Create(); err != nil {
		httpResp.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	httpResp.RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"status":  "news created",
		"news_id": n.NewsID,
	})
}

// DeleteNews removes a news article (admin only)
func DeleteNews(w http.ResponseWriter, r *http.Request) {
	_, role, err := GetSessionUser(r)
	if err != nil || role != "admin" {
		httpResp.RespondWithError(w, http.StatusUnauthorized, "admin access required")
		return
	}
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		httpResp.RespondWithError(w, http.StatusBadRequest, "invalid news id")
		return
	}
	n := model.News{NewsID: id}
	if err := n.Delete(); err != nil {
		httpResp.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpResp.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
