package model

import (
	"drukarena/backend/dataStore/postgres"
)

// News represents a news article
type News struct {
	NewsID        int64  `json:"news_id"`
	Title         string `json:"title"`
	Content       string `json:"content"`
	ImageURL      string `json:"image_url"`
	PublishedDate string `json:"published_date"`
	AuthorID      int64  `json:"author_id"`
}

const (
	queryInsertNews  = `INSERT INTO news(title, content, image_url, author_id) VALUES($1,$2,$3,$4) RETURNING news_id;`
	queryGetAllNews  = `SELECT news_id, title, content, COALESCE(image_url,''), published_date::text, COALESCE(author_id,0) FROM news ORDER BY published_date DESC;`
	queryDeleteNews  = `DELETE FROM news WHERE news_id=$1 RETURNING news_id;`
	queryGetNewsByID = `SELECT news_id, title, content, COALESCE(image_url,''), published_date::text, COALESCE(author_id,0) FROM news WHERE news_id=$1;`
)

// Create inserts a news article
func (n *News) Create() error {
	return postgres.Db.QueryRow(queryInsertNews,
		n.Title, n.Content, n.ImageURL, n.AuthorID,
	).Scan(&n.NewsID)
}

// Delete removes a news article
func (n *News) Delete() error {
	_, err := postgres.Db.Exec(queryDeleteNews, n.NewsID)
	return err
}

// GetAllNews returns all news articles
func GetAllNews() ([]News, error) {
	rows, err := postgres.Db.Query(queryGetAllNews)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	newsList := []News{}
	for rows.Next() {
		var n News
		if err := rows.Scan(&n.NewsID, &n.Title, &n.Content,
			&n.ImageURL, &n.PublishedDate, &n.AuthorID); err != nil {
			return nil, err
		}
		newsList = append(newsList, n)
	}
	return newsList, nil
}
