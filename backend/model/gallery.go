package model

import (
	"strings"

	"drukarena/backend/dataStore/postgres"
)

type GalleryItem struct {
	GalleryID      int64  `json:"gallery_id"`
	Title          string `json:"title"`
	ImageURL       string `json:"image_url"`
	UploadedBy     int64  `json:"uploaded_by"`
	ApprovalStatus string `json:"approval_status"`
	CreatedAt      string `json:"created_at"`
	Username       string `json:"username,omitempty"`
	UploaderName   string `json:"uploader_name,omitempty"`
	UploaderEmail  string `json:"uploader_email,omitempty"`
}

func CreateGalleryItem(g *GalleryItem) error {
	if err := EnsureGalleryModerationSchema(); err != nil {
		return err
	}
	if g.UploadedBy == 0 {
		return postgres.Db.QueryRow(
			`INSERT INTO gallery_uploads(title, image_url, uploader_name, uploader_email, approval_status)
			 VALUES($1,$2,$3,$4,'approved') RETURNING gallery_id`,
			g.Title, g.ImageURL, g.UploaderName, strings.ToLower(strings.TrimSpace(g.UploaderEmail)),
		).Scan(&g.GalleryID)
	}
	return postgres.Db.QueryRow(
		`INSERT INTO gallery_uploads(title, image_url, uploaded_by, uploader_name, uploader_email, approval_status)
		 VALUES($1,$2,$3,$4,$5,'approved') RETURNING gallery_id`,
		g.Title, g.ImageURL, g.UploadedBy, g.UploaderName, strings.ToLower(strings.TrimSpace(g.UploaderEmail)),
	).Scan(&g.GalleryID)
}

func GetApprovedGalleryItems() ([]GalleryItem, error) {
	if err := EnsureGalleryModerationSchema(); err != nil {
		return nil, err
	}
	return scanGalleryItems(
		`SELECT g.gallery_id, g.title, g.image_url, COALESCE(g.uploaded_by,0),
		        g.approval_status, g.created_at::text,
		        COALESCE(NULLIF(g.uploader_name,''), u.username, 'Guest')
		 FROM gallery_uploads g
		 LEFT JOIN users u ON u.user_id=g.uploaded_by
		 WHERE g.approval_status='approved'
		 ORDER BY g.created_at DESC`,
	)
}

func GetAllGalleryItems() ([]GalleryItem, error) {
	if err := EnsureGalleryModerationSchema(); err != nil {
		return nil, err
	}
	return scanGalleryItems(
		`SELECT g.gallery_id, g.title, g.image_url, COALESCE(g.uploaded_by,0),
		        g.approval_status, g.created_at::text,
		        COALESCE(NULLIF(g.uploader_name,''), u.username, 'Guest'),
		        COALESCE(NULLIF(g.uploader_email,''), u.email, '')
		 FROM gallery_uploads g
		 LEFT JOIN users u ON u.user_id=g.uploaded_by
		 ORDER BY g.created_at DESC`,
	)
}

func scanGalleryItems(query string) ([]GalleryItem, error) {
	rows, err := postgres.Db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []GalleryItem{}
	for rows.Next() {
		var g GalleryItem
		dest := []interface{}{&g.GalleryID, &g.Title, &g.ImageURL, &g.UploadedBy, &g.ApprovalStatus, &g.CreatedAt, &g.Username}
		if strings.Contains(query, "uploader_email") {
			dest = append(dest, &g.UploaderEmail)
		}
		if err := rows.Scan(dest...); err != nil {
			return nil, err
		}
		g.UploaderName = g.Username
		items = append(items, g)
	}
	return items, nil
}

func SetGalleryApproval(id int64, approval string) error {
	_, err := postgres.Db.Exec(`UPDATE gallery_uploads SET approval_status=$1 WHERE gallery_id=$2`, approval, id)
	return err
}

func DeleteGalleryItem(id int64) error {
	_, err := postgres.Db.Exec(`DELETE FROM gallery_uploads WHERE gallery_id=$1`, id)
	return err
}

func BlockEmail(email, reason string) error {
	if err := EnsureGalleryModerationSchema(); err != nil {
		return err
	}
	email = strings.ToLower(strings.TrimSpace(email))
	_, err := postgres.Db.Exec(
		`INSERT INTO blocked_emails(email, reason) VALUES($1,$2)
		 ON CONFLICT(email) DO UPDATE SET reason=EXCLUDED.reason, created_at=CURRENT_TIMESTAMP`,
		email, reason,
	)
	return err
}

func IsEmailBlocked(email string) (bool, error) {
	if err := EnsureGalleryModerationSchema(); err != nil {
		return false, err
	}
	var blocked bool
	err := postgres.Db.QueryRow(`SELECT EXISTS(SELECT 1 FROM blocked_emails WHERE email=$1)`, strings.ToLower(strings.TrimSpace(email))).Scan(&blocked)
	return blocked, err
}

func GetGalleryItem(id int64) (GalleryItem, error) {
	if err := EnsureGalleryModerationSchema(); err != nil {
		return GalleryItem{}, err
	}
	var g GalleryItem
	err := postgres.Db.QueryRow(
		`SELECT g.gallery_id, g.title, g.image_url, COALESCE(g.uploaded_by,0),
		        g.approval_status, g.created_at::text,
		        COALESCE(NULLIF(g.uploader_name,''), u.username, 'Guest'),
		        COALESCE(NULLIF(g.uploader_email,''), u.email, '')
		 FROM gallery_uploads g
		 LEFT JOIN users u ON u.user_id=g.uploaded_by
		 WHERE g.gallery_id=$1`,
		id,
	).Scan(&g.GalleryID, &g.Title, &g.ImageURL, &g.UploadedBy, &g.ApprovalStatus, &g.CreatedAt, &g.Username, &g.UploaderEmail)
	g.UploaderName = g.Username
	return g, err
}

func EnsureGalleryModerationSchema() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS gallery_uploads (
			gallery_id BIGSERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			image_url TEXT NOT NULL,
			uploaded_by BIGINT REFERENCES users(user_id) ON DELETE SET NULL,
			uploader_name TEXT,
			uploader_email TEXT,
			approval_status TEXT NOT NULL DEFAULT 'approved',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`ALTER TABLE gallery_uploads ADD COLUMN IF NOT EXISTS uploader_name TEXT`,
		`ALTER TABLE gallery_uploads ADD COLUMN IF NOT EXISTS uploader_email TEXT`,
		`CREATE TABLE IF NOT EXISTS blocked_emails (
			blocked_id BIGSERIAL PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			reason TEXT,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_gallery_uploads_approval_status ON gallery_uploads(approval_status)`,
		`CREATE INDEX IF NOT EXISTS idx_gallery_uploads_uploaded_by ON gallery_uploads(uploaded_by)`,
		`UPDATE gallery_uploads SET approval_status='approved' WHERE approval_status IS NULL OR approval_status='pending'`,
	}
	for _, stmt := range statements {
		if _, err := postgres.Db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}
