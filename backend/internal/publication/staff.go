package publication

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

type Staff struct{ db *sql.DB }

func NewStaff(db *sql.DB) *Staff { return &Staff{db} }

type Review struct {
	PostID      int64
	Revision    int64
	State       Exposure
	Body, Board string
	Exposed     bool
}
type Detail struct {
	ID                       int64
	Body                     string
	State                    Exposure
	Revision, ConversationID int64
	Board                    string
	Locked, Pinned           bool
	ConversationRevision     int64
	Exposed                  bool
	HasMedia                 bool
}
type Exposure string

const (
	Visible Exposure = "visible"
	Hidden  Exposure = "hidden"
	Erased  Exposure = "erased"
)

type PostDecision struct {
	Post     PostID
	Expected int64
	State    Exposure
	Reason   string
}
type ConversationDecision struct {
	Conversation   ConversationID
	Expected       int64
	Board          string
	Locked, Pinned bool
	Reason         string
}
type ReviewDecision struct {
	Post     PostID
	Expected int64
	Reason   string
}

func (s *Store) Report(ctx context.Context, id PostID) error {
	_, err := s.reader.ExecContext(ctx, `SELECT pub.report($1)`, id)
	return err
}
func (s *Staff) Queue(ctx context.Context, page int) ([]Review, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT post_id,revision,state,body,board,exposed FROM pub.review_queue ORDER BY post_id LIMIT 50 OFFSET $1`, (max(1, min(page, 100000))-1)*50)
	if err != nil {
		return nil, err
	}
	return collect(rows, func(row scanner) (r Review, err error) {
		err = row.Scan(&r.PostID, &r.Revision, &r.State, &r.Body, &r.Board, &r.Exposed)
		return
	})
}
func (s *Staff) Detail(ctx context.Context, id int64) (Detail, error) {
	var d Detail
	err := s.db.QueryRowContext(ctx, `SELECT id,body,state,revision,conversation_id,board,locked,pinned,conversation_revision,exposed,has_media FROM pub.staff_posts WHERE id=$1`, id).Scan(&d.ID, &d.Body, &d.State, &d.Revision, &d.ConversationID, &d.Board, &d.Locked, &d.Pinned, &d.ConversationRevision, &d.Exposed, &d.HasMedia)
	return d, err
}

var decisionErrors = map[pq.ErrorCode]*Error{
	"22023": {400, "Invalid decision input."},
	"23503": {400, "Invalid decision input."},
	"P0001": {409, "Decision conflicts with current state. Reload and review."},
	"P0002": {404, "Not found."},
}

func (s *Staff) executeDecision(ctx context.Context, actor, reason, statement string, args ...any) error {
	if actor == "" || reason == "" {
		return &Error{400, "Actor and reason required."}
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return commandError(ctx, err, decisionErrors)
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `SELECT set_config('app.actor',$1,true)`, actor); err != nil {
		return commandError(ctx, err, decisionErrors)
	}
	if _, err = tx.ExecContext(ctx, statement, args...); err != nil {
		return commandError(ctx, err, decisionErrors)
	}
	return commandError(ctx, tx.Commit(), decisionErrors)
}
func (s *Staff) SetPost(ctx context.Context, actor string, d PostDecision) error {
	if d.Post < 1 || d.Expected < 0 || (d.State != Visible && d.State != Hidden && d.State != Erased) {
		return &Error{400, "Invalid post decision."}
	}
	return s.executeDecision(ctx, actor, d.Reason, `SELECT pub.set_post($1,$2,$3,$4)`, d.Post, d.Expected, d.State, d.Reason)
}
func (s *Staff) Configure(ctx context.Context, actor string, d ConversationDecision) error {
	if d.Conversation < 1 || d.Expected < 0 {
		return &Error{400, "Invalid conversation decision."}
	}
	return s.executeDecision(ctx, actor, d.Reason, `SELECT pub.configure($1,$2,$3,$4,$5,$6)`, d.Conversation, d.Expected, d.Reason, d.Board, d.Locked, d.Pinned)
}
func (s *Staff) CloseReview(ctx context.Context, actor string, d ReviewDecision) error {
	if d.Post < 1 || d.Expected < 0 {
		return &Error{400, "Invalid review decision."}
	}
	return s.executeDecision(ctx, actor, d.Reason, `SELECT pub.close_report($1,$2,$3)`, d.Post, d.Expected, d.Reason)
}
func (s *Staff) Media(ctx context.Context, id PostID) ([]byte, error) {
	var data []byte
	err := s.db.QueryRowContext(ctx, `SELECT bytes FROM pub.staff_media WHERE post_id=$1`, id).Scan(&data)
	return data, err
}

func (s *Staff) Ready(ctx context.Context) error { return s.db.PingContext(ctx) }
