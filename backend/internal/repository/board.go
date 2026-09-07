package repository

import (
	"database/sql"
	"fmt"

	"216chan/backend/internal/model"
)

type BoardRepo struct {
	db *sql.DB
}

func NewBoardRepo(db *sql.DB) *BoardRepo {
	return &BoardRepo{db: db}
}

func (r *BoardRepo) GetBoard(slug string) (*model.Board, error) {
	b := &model.Board{}
	err := r.db.QueryRow(`SELECT slug, name, description, is_nsfw, created_at FROM boards WHERE slug=$1`, slug).Scan(&b.Slug, &b.Name, &b.Description, &b.IsNSFW, &b.CreatedAt)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (r *BoardRepo) CountThreads(board string) (int, error) {
	var c int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM threads WHERE board_slug=$1`, board).Scan(&c)
	return c, err
}

func (r *BoardRepo) ListThreads(board string, offset, limit int) ([]model.Thread, error) {
	rows, err := r.db.Query(`
		SELECT id, board_slug, is_pinned, is_locked, reply_count, bumped_at, created_at
		FROM threads
		WHERE board_slug=$1
		ORDER BY is_pinned DESC, bumped_at DESC
		LIMIT $2 OFFSET $3`, board, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Thread
	for rows.Next() {
		var t model.Thread
		if err := rows.Scan(&t.ID, &t.BoardSlug, &t.IsPinned, &t.IsLocked, &t.ReplyCount, &t.BumpedAt, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *BoardRepo) GetThread(board string, id int64) (*model.Thread, error) {
	t := &model.Thread{}
	err := r.db.QueryRow(`SELECT id, board_slug, is_pinned, is_locked, reply_count, bumped_at, created_at FROM threads WHERE id=$1 AND board_slug=$2`, id, board).Scan(&t.ID, &t.BoardSlug, &t.IsPinned, &t.IsLocked, &t.ReplyCount, &t.BumpedAt, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *BoardRepo) GetOP(threadID int64) (*model.Post, error) {
	p := &model.Post{}
	err := r.db.QueryRow(`SELECT id, thread_id, board_slug, is_op, body, image_url, image_name, file_size_bytes, country_code, created_at FROM posts WHERE thread_id=$1 AND is_op=true LIMIT 1`, threadID).Scan(&p.ID, &p.ThreadID, &p.BoardSlug, &p.IsOP, &p.Body, &p.ImageURL, &p.ImageName, &p.FileSize, &p.CountryCode, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *BoardRepo) GetRepliesPreview(threadIDs []int64, perThread int) (map[int64][]model.Post, error) {
	if len(threadIDs) == 0 {
		return map[int64][]model.Post{}, nil
	}
	out := make(map[int64][]model.Post)
	for _, tid := range threadIDs {
		rows, err := r.db.Query(`SELECT id, thread_id, board_slug, is_op, body, image_url, image_name, file_size_bytes, country_code, created_at FROM posts WHERE thread_id=$1 AND is_op=false ORDER BY created_at ASC LIMIT $2`, tid, perThread)
		if err != nil {
			return nil, err
		}
		var reps []model.Post
		for rows.Next() {
			var p model.Post
			if err := rows.Scan(&p.ID, &p.ThreadID, &p.BoardSlug, &p.IsOP, &p.Body, &p.ImageURL, &p.ImageName, &p.FileSize, &p.CountryCode, &p.CreatedAt); err != nil {
				rows.Close()
				return nil, err
			}
			reps = append(reps, p)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return nil, err
		}
		out[tid] = reps
	}
	return out, nil
}

func (r *BoardRepo) CountReplies(threadID int64) (int, error) {
	var c int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM posts WHERE thread_id=$1 AND is_op=false`, threadID).Scan(&c)
	return c, err
}

func (r *BoardRepo) ListReplies(threadID int64, offset, limit int) ([]model.Post, error) {
	rows, err := r.db.Query(`SELECT id, thread_id, board_slug, is_op, body, image_url, image_name, file_size_bytes, country_code, created_at FROM posts WHERE thread_id=$1 AND is_op=false ORDER BY created_at ASC LIMIT $2 OFFSET $3`, threadID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Post
	for rows.Next() {
		var p model.Post
		if err := rows.Scan(&p.ID, &p.ThreadID, &p.BoardSlug, &p.IsOP, &p.Body, &p.ImageURL, &p.ImageName, &p.FileSize, &p.CountryCode, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *BoardRepo) CreateThread(board, body string, imageURL, imageName *string, fileSize *int, country, ipHash string) (int64, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	var tid int64
	err = tx.QueryRow(`INSERT INTO threads (board_slug, bumped_at) VALUES ($1, NOW()) RETURNING id`, board).Scan(&tid)
	if err != nil {
		return 0, fmt.Errorf("insert thread: %w", err)
	}
	_, err = tx.Exec(`INSERT INTO posts (thread_id, board_slug, is_op, body, image_url, image_name, file_size_bytes, country_code, ip_hash) VALUES ($1,$2,true,$3,$4,$5,$6,$7,$8)`, tid, board, body, imageURL, imageName, fileSize, country, ipHash)
	if err != nil {
		return 0, fmt.Errorf("insert op: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return tid, nil
}

func (r *BoardRepo) CreateReply(threadID int64, board, body string, imageURL, imageName *string, fileSize *int, country, ipHash string) (int64, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	var isLocked bool
	err = tx.QueryRow(`SELECT is_locked FROM threads WHERE id=$1 AND board_slug=$2 FOR UPDATE`, threadID, board).Scan(&isLocked)
	if err != nil {
		return 0, err
	}
	if isLocked {
		return 0, fmt.Errorf("thread locked")
	}
	var pid int64
	err = tx.QueryRow(`INSERT INTO posts (thread_id, board_slug, is_op, body, image_url, image_name, file_size_bytes, country_code, ip_hash) VALUES ($1,$2,false,$3,$4,$5,$6,$7,$8) RETURNING id`, threadID, board, body, imageURL, imageName, fileSize, country, ipHash).Scan(&pid)
	if err != nil {
		return 0, err
	}
	_, err = tx.Exec(`UPDATE threads SET reply_count = reply_count + 1, bumped_at = NOW() WHERE id=$1`, threadID)
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return pid, nil
}