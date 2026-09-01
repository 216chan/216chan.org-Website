package repository

import (
	"database/sql"
	"fmt"

	"216chan/backend/internal/model"
)

type StatsRepo struct {
	db *sql.DB
}

func NewStatsRepo(db *sql.DB) *StatsRepo {
	return &StatsRepo{db: db}
}

func (r *StatsRepo) GetStats() (*model.Stats, error) {
	s := &model.Stats{}

	if err := r.db.QueryRow(`SELECT COUNT(*) FROM posts`).Scan(&s.TotalPosts); err != nil {
		return nil, fmt.Errorf("count posts: %w", err)
	}

	if err := r.db.QueryRow(
		`SELECT COUNT(*) FROM sessions WHERE last_seen_at > NOW() - INTERVAL '15 minutes'`,
	).Scan(&s.CurrentUsers); err != nil {
		return nil, fmt.Errorf("count sessions: %w", err)
	}

	var totalBytes int64
	if err := r.db.QueryRow(
		`SELECT COALESCE(SUM(file_size_bytes), 0) FROM posts
		 WHERE file_size_bytes IS NOT NULL
		   AND created_at > NOW() - INTERVAL '24 hours'`,
	).Scan(&totalBytes); err != nil {
		return nil, fmt.Errorf("sum content: %w", err)
	}
	s.ActiveContent = formatBytes(totalBytes)

	return s, nil
}

func formatBytes(b int64) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)
	switch {
	case b >= gb:
		return fmt.Sprintf("%.0f GB", float64(b)/float64(gb))
	case b >= mb:
		return fmt.Sprintf("%.0f MB", float64(b)/float64(mb))
	case b >= kb:
		return fmt.Sprintf("%.0f KB", float64(b)/float64(kb))
	default:
		return fmt.Sprintf("%d B", b)
	}
}
