package publication

import (
	"context"
	"database/sql"
	"fmt"

	"216chan/backend/internal/model"
	"github.com/lib/pq"
)

func (s *Store) snapshot(ctx context.Context) (*sql.Tx, error) {
	return s.reader.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
}
func readBoard(ctx context.Context, tx *sql.Tx, slug string) (model.Board, error) {
	var b model.Board
	err := tx.QueryRowContext(ctx, `SELECT slug,name,description,is_nsfw,created_at FROM pub.public_boards WHERE slug=$1`, slug).Scan(&b.Slug, &b.Name, &b.Description, &b.IsNSFW, &b.CreatedAt)
	return b, err
}

const threadPageSize = 15

func pages(page, total, limit int) (int, int) {
	last := max(1, (total+limit-1)/limit)
	return min(max(1, page), last), last
}

const threadFields = `id,board,pinned,locked,reply_count,activity,created_at`
const postFields = `id,conversation_id,board,is_op,body,created_at,media_size,media_width,media_height,country_code`

type scanner interface{ Scan(...any) error }

// collect owns rows: callers must not reuse the transaction while rows are open.
func collect[T any](rows *sql.Rows, scan func(scanner) (T, error)) ([]T, error) {
	defer rows.Close()
	out := []T{}
	for rows.Next() {
		v, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func scanThread(row scanner) (model.Thread, error) {
	var t model.Thread
	err := row.Scan(&t.ID, &t.BoardSlug, &t.IsPinned, &t.IsLocked, &t.ReplyCount, &t.BumpedAt, &t.CreatedAt)
	return t, err
}
func scanPost(row scanner) (model.Post, error) {
	var p model.Post
	var size sql.NullInt64
	var width, height sql.NullInt64
	var country sql.NullString
	err := row.Scan(&p.ID, &p.ThreadID, &p.BoardSlug, &p.IsOP, &p.Body, &p.CreatedAt, &size, &width, &height, &country)
	p.CountryCode = "XX"
	if country.Valid {
		p.CountryCode = country.String
	}
	p.Permalink = fmt.Sprintf("/posts/%d", p.ID)
	if size.Valid {
		p.Media = &model.Media{URL: fmt.Sprintf("/media/%d", p.ID), Name: fmt.Sprintf("post-%d.webp", p.ID), Size: int(size.Int64), Width: int(width.Int64), Height: int(height.Int64)}
	}
	return p, err
}

func (s *Store) Board(ctx context.Context, slug string, page int) (*model.ThreadsPage, error) {
	tx, err := s.snapshot(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	b, err := readBoard(ctx, tx, slug)
	if err != nil {
		return nil, err
	}
	var total int
	if err = tx.QueryRowContext(ctx, `SELECT count(*) FROM pub.visible_conversations WHERE board=$1`, slug).Scan(&total); err != nil {
		return nil, err
	}
	page, last := pages(page, total, 10)
	rows, err := tx.QueryContext(ctx, `SELECT `+threadFields+` FROM pub.visible_conversations WHERE board=$1 ORDER BY pinned DESC,activity DESC,id DESC LIMIT 10 OFFSET $2`, slug, (page-1)*10)
	if err != nil {
		return nil, err
	}
	threads, err := collect(rows, scanThread)
	if err != nil {
		return nil, err
	}
	out := make([]model.ThreadWithOP, 0, len(threads))
	ids := []int64{}
	indices := map[int64]int{}
	for _, t := range threads {
		indices[t.ID] = len(out)
		ids = append(ids, t.ID)
		out = append(out, model.ThreadWithOP{Thread: t, Replies: []model.Post{}})
	}
	if len(ids) > 0 {
		rows, err = tx.QueryContext(ctx, `SELECT `+postFields+` FROM (SELECT p.*,row_number() OVER(PARTITION BY conversation_id ORDER BY is_op DESC,created_at DESC,id DESC) AS n FROM pub.visible_posts p WHERE conversation_id=ANY($1)) ranked WHERE n<=4 ORDER BY conversation_id,is_op DESC,created_at,id`, pq.Array(ids))
		if err != nil {
			return nil, err
		}
		posts, err := collect(rows, scanPost)
		if err != nil {
			return nil, err
		}
		for _, p := range posts {
			i := indices[p.ThreadID]
			if p.IsOP {
				out[i].OP = p
			} else {
				out[i].Replies = append(out[i].Replies, p)
			}
		}
	}
	for _, thread := range out {
		if thread.OP.ID == 0 { // A visible conversation always has its visible root in the same snapshot.
			return nil, fmt.Errorf("conversation %d has no opening post", thread.Thread.ID)
		}
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &model.ThreadsPage{Board: b, Threads: out, Pagination: model.Pagination{Page: page, Pages: last, Total: total}}, nil
}

func (s *Store) Thread(ctx context.Context, slug string, id int64, page int) (*model.ThreadPage, error) {
	tx, err := s.snapshot(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	b, err := readBoard(ctx, tx, slug)
	if err != nil {
		return nil, err
	}
	t, err := scanThread(tx.QueryRowContext(ctx, `SELECT `+threadFields+` FROM pub.visible_conversations WHERE board=$1 AND id=$2`, slug, id))
	if err != nil {
		return nil, err
	}
	root, err := scanPost(tx.QueryRowContext(ctx, `SELECT `+postFields+` FROM pub.visible_posts WHERE conversation_id=$1 AND is_op`, id))
	if err != nil {
		return nil, err
	}
	page, last := pages(page, t.ReplyCount, threadPageSize)
	rows, err := tx.QueryContext(ctx, `SELECT `+postFields+` FROM pub.visible_posts WHERE conversation_id=$1 AND NOT is_op ORDER BY created_at,id LIMIT $3 OFFSET $2`, id, (page-1)*threadPageSize, threadPageSize)
	if err != nil {
		return nil, err
	}
	replies, err := collect(rows, scanPost)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &model.ThreadPage{Board: b, Thread: t, OP: root, Replies: replies, Pagination: model.Pagination{Page: page, Pages: last, Total: t.ReplyCount}}, nil
}

func (s *Store) PostLocation(ctx context.Context, id int64) (string, error) {
	var board string
	var cid int64
	var page int
	err := s.reader.QueryRowContext(ctx, `SELECT p.board,p.conversation_id,CASE WHEN p.is_op THEN 1 ELSE
 ((SELECT count(*) FROM pub.visible_posts r WHERE r.conversation_id=p.conversation_id AND NOT r.is_op AND (r.created_at,r.id)<=(p.created_at,p.id))-1)/$2+1 END
 FROM pub.visible_posts p WHERE p.id=$1`, id, threadPageSize).Scan(&board, &cid, &page)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("/%s/%d?page=%d#post-%d", board, cid, page, id), nil
}
func (s *Store) Stats(ctx context.Context) (*model.Stats, error) {
	var stats model.Stats
	err := s.reader.QueryRowContext(ctx, `SELECT count(*),count(*) FILTER(WHERE is_op),count(*) FILTER(WHERE created_at>now()-interval '24 hours'),coalesce(sum(media_size),0) FROM pub.visible_posts`).Scan(&stats.TotalPosts, &stats.VisibleThreads, &stats.PostsLastDay, &stats.MediaBytes)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}
