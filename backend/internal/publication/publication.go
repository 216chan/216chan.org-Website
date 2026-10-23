package publication

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/lib/pq"
)

type Store struct{ reader, writer *sql.DB }

func New(reader, writer *sql.DB) *Store { return &Store{reader, writer} }

type Error struct {
	Status  int
	Message string
}

func (e *Error) Error() string { return e.Message }
func Status(err error) int {
	var e *Error
	if errors.As(err, &e) {
		return e.Status
	}
	if errors.Is(err, sql.ErrNoRows) {
		return 404
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return 503
	}
	var pg *pq.Error
	if errors.As(err, &pg) && (pg.Code == "55P03" || pg.Code.Class() == "08" || pg.Code.Class() == "53" || pg.Code.Class() == "57") {
		return 503
	}
	return 500
}
func Message(err error) string {
	var e *Error
	if errors.As(err, &e) {
		return e.Message
	}
	if Status(err) == 404 {
		return "Not found."
	}
	return "Request could not be completed."
}

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func ValidReceipt(receipt string) bool { return uuidPattern.MatchString(receipt) }

func NewReceipt() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = (b[6] & 15) | 64
	b[8] = (b[8] & 63) | 128
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}

type PostID int64
type ConversationID int64

type Submission struct {
	Receipt, Board, Body string
	Conversation         ConversationID
	Media                *ApprovedMedia
	CountryCode          string // Server observation; not part of receipt identity.
}

type Accepted struct {
	PostID *PostID `json:"post_id"` // Tagged so clients decode what MarshalJSON writes.
}

func (a Accepted) MarshalJSON() ([]byte, error) {
	state := "accepted_unavailable"
	if a.PostID != nil {
		state = "accepted_visible"
	}
	return json.Marshal(struct {
		State  string  `json:"state"`
		PostID *PostID `json:"post_id"`
	}{state, a.PostID})
}

func commandError(ctx context.Context, err error, codes map[pq.ErrorCode]*Error) error {
	if err == nil {
		return nil
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	var pg *pq.Error
	if errors.As(err, &pg) {
		if mapped, ok := codes[pg.Code]; ok {
			return mapped
		}
	}
	return err
}

var ErrReceiptConflict = &Error{409, "Receipt conflicts with a different submission."}

var publishErrors = map[pq.ErrorCode]*Error{
	"22023": {400, "Invalid submission."},
	"P0002": {404, "Conversation not found."},
	"P2001": ErrReceiptConflict,
	"23503": {404, "Board not found."},
	"P0001": {409, "Conversation is unavailable or locked."},
}

func (s *Store) Publish(ctx context.Context, in Submission) (Accepted, error) {
	in.Body = strings.TrimSpace(in.Body)
	if in.Media != nil && (len(in.Media.bytes) == 0 || len(in.Media.bytes) > MaxMediaBytes) {
		return Accepted{}, &Error{400, "Approved media is empty or oversized."}
	}
	if !ValidReceipt(in.Receipt) || in.Conversation < 0 || !utf8.ValidString(in.Body) || strings.ContainsRune(in.Body, 0) || utf8.RuneCountInString(in.Body) > 5000 || (in.Body == "" && in.Media == nil) {
		return Accepted{}, &Error{400, "A receipt and body of 1–5000 characters are required."}
	}
	var target any
	if in.Conversation != 0 {
		target = in.Conversation
	}
	var media, source any
	if in.Media != nil {
		media = in.Media.bytes
		source = in.Media.original[:]
	}
	var id sql.NullInt64
	var country any
	if in.CountryCode != "" && in.CountryCode != "XX" {
		country = in.CountryCode
	}
	err := s.writer.QueryRowContext(ctx, `SELECT pub.publish($1,$2,$3,$4,$5,$6,$7)`, in.Receipt, in.Board, target, in.Body, media, source, country).Scan(&id)
	if err != nil {
		return Accepted{}, commandError(ctx, err, publishErrors)
	}
	if !id.Valid {
		return Accepted{}, nil // Acceptance survives withdrawal.
	}
	post := PostID(id.Int64)
	return Accepted{&post}, nil
}

func CheckRoles(ctx context.Context, reader, writer, staff *sql.DB) error {
	names := map[string]bool{}
	for i, db := range [3]*sql.DB{reader, writer, staff} {
		name, err := checkRole(ctx, db, roles[i])
		if err != nil {
			return err
		}
		if names[name] {
			return fmt.Errorf("request roles must differ")
		}
		names[name] = true
	}
	return nil
}

func (s *Store) Ready(ctx context.Context) error {
	if err := s.reader.PingContext(ctx); err != nil {
		return err
	}
	return s.writer.PingContext(ctx)
}
