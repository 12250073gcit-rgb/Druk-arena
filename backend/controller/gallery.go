package controller

import (
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"drukarena/backend/model"
	"drukarena/utils/httpResp"

	"github.com/gorilla/mux"
)

func GetGallery(w http.ResponseWriter, r *http.Request) {
	items, err := model.GetApprovedGalleryItems()
	if err != nil {
		httpResp.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpResp.RespondWithJSON(w, http.StatusOK, items)
}

func GetAdminGallery(w http.ResponseWriter, r *http.Request) {
	_, role, err := GetSessionUser(r)
	if err != nil || role != "admin" {
		httpResp.RespondWithError(w, http.StatusUnauthorized, "admin access required")
		return
	}
	items, err := model.GetAllGalleryItems()
	if err != nil {
		httpResp.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpResp.RespondWithJSON(w, http.StatusOK, items)
}

func UploadGallery(w http.ResponseWriter, r *http.Request) {
	var userID int64
	var uploaderName string
	var uploaderEmail string
	if id, _, err := GetSessionUser(r); err == nil {
		userID = id
		u := model.User{UserID: id}
		if err := u.GetByID(); err == nil {
			uploaderName = u.Username
			uploaderEmail = u.Email
		}
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		httpResp.RespondWithError(w, http.StatusBadRequest, "image upload must be 10MB or less")
		return
	}

	if r.FormValue("agree_terms") != "true" {
		httpResp.RespondWithError(w, http.StatusBadRequest, "you must agree to the gallery upload rules")
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	if title == "" {
		title = "Match Photo"
	}
	if len(title) > 80 {
		httpResp.RespondWithError(w, http.StatusBadRequest, "photo title must be 80 characters or less")
		return
	}
	if userID == 0 {
		uploaderName = randomGuestName()
	}
	if uploaderEmail != "" {
		blocked, err := model.IsEmailBlocked(uploaderEmail)
		if err != nil {
			httpResp.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if blocked {
			httpResp.RespondWithError(w, http.StatusForbidden, "this account is blocked from uploading photos")
			return
		}
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		httpResp.RespondWithError(w, http.StatusBadRequest, "image file is required")
		return
	}
	defer file.Close()
	if header.Size > 10<<20 {
		httpResp.RespondWithError(w, http.StatusBadRequest, "image upload must be 10MB or less")
		return
	}
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		httpResp.RespondWithError(w, http.StatusBadRequest, "only JPG, PNG, and WEBP images are allowed")
		return
	}

	filename := fmt.Sprintf("match-%d%s", time.Now().UnixNano(), ext)
	dest, err := createUploadFile(filename)
	if err != nil {
		httpResp.RespondWithError(w, http.StatusInternalServerError, "could not save image: "+err.Error())
		return
	}
	defer dest.Close()

	if _, err := io.Copy(dest, file); err != nil {
		httpResp.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	item := model.GalleryItem{
		Title:         title,
		ImageURL:      "/uploads/" + filename,
		UploadedBy:    userID,
		UploaderName:  uploaderName,
		UploaderEmail: uploaderEmail,
	}
	if err := model.CreateGalleryItem(&item); err != nil {
		httpResp.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	httpResp.RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"status":     "uploaded",
		"gallery_id": item.GalleryID,
	})
}

func ApproveGallery(w http.ResponseWriter, r *http.Request) {
	setGalleryApproval(w, r, "approved")
}

func RejectGallery(w http.ResponseWriter, r *http.Request) {
	setGalleryApproval(w, r, "rejected")
}

func setGalleryApproval(w http.ResponseWriter, r *http.Request, approval string) {
	_, role, err := GetSessionUser(r)
	if err != nil || role != "admin" {
		httpResp.RespondWithError(w, http.StatusUnauthorized, "admin access required")
		return
	}
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		httpResp.RespondWithError(w, http.StatusBadRequest, "invalid gallery id")
		return
	}
	if err := model.SetGalleryApproval(id, approval); err != nil {
		httpResp.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpResp.RespondWithJSON(w, http.StatusOK, map[string]string{"status": approval})
}

func DeleteGallery(w http.ResponseWriter, r *http.Request) {
	_, role, err := GetSessionUser(r)
	if err != nil || role != "admin" {
		httpResp.RespondWithError(w, http.StatusUnauthorized, "admin access required")
		return
	}
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		httpResp.RespondWithError(w, http.StatusBadRequest, "invalid gallery id")
		return
	}
	if err := model.DeleteGalleryItem(id); err != nil {
		httpResp.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpResp.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func BlockGalleryUploader(w http.ResponseWriter, r *http.Request) {
	_, role, err := GetSessionUser(r)
	if err != nil || role != "admin" {
		httpResp.RespondWithError(w, http.StatusUnauthorized, "admin access required")
		return
	}
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		httpResp.RespondWithError(w, http.StatusBadRequest, "invalid gallery id")
		return
	}
	item, err := model.GetGalleryItem(id)
	if err != nil {
		httpResp.RespondWithError(w, http.StatusNotFound, "gallery upload not found")
		return
	}
	if item.UploaderEmail == "" {
		httpResp.RespondWithError(w, http.StatusBadRequest, "this upload does not have an email to block")
		return
	}
	if err := model.BlockEmail(item.UploaderEmail, "Blocked after inappropriate gallery upload"); err != nil {
		httpResp.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpResp.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "email blocked"})
}

func randomGuestName() string {
	var b [2]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("Guest-%d", time.Now().Unix()%10000)
	}
	return fmt.Sprintf("Guest-%04X", int(b[0])<<8|int(b[1]))
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func createUploadFile(filename string) (*os.File, error) {
	var lastErr error
	for _, dir := range uploadDirs() {
		if err := os.MkdirAll(dir, 0755); err != nil {
			lastErr = err
			continue
		}
		file, err := os.Create(filepath.Join(dir, filename))
		if err == nil {
			return file, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

func uploadDirs() []string {
	primary := getEnv("UPLOAD_DIR", filepath.Join("view", "uploads"))
	fallback := filepath.Join(os.TempDir(), "drukarena", "uploads")
	if primary == fallback {
		return []string{primary}
	}
	return []string{primary, fallback}
}
