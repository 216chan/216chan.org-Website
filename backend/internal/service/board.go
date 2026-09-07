package service

import (
	"database/sql"
	"fmt"

	"216chan/backend/internal/model"
	"216chan/backend/internal/repository"
)

type BoardService struct {
	repo *repository.BoardRepo
}

func NewBoardService(repo *repository.BoardRepo) *BoardService {
	return &BoardService{repo: repo}
}

func (s *BoardService) GetBoard(slug string) (*model.Board, error) {
	return s.repo.GetBoard(slug)
}

func (s *BoardService) ListThreads(board string, page, limit int) (*model.ThreadsPage, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}
	b, err := s.repo.GetBoard(board)
	if err != nil {
		return nil, err
	}
	total, err := s.repo.CountThreads(board)
	if err != nil {
		return nil, err
	}
	pages := (total + limit - 1) / limit
	if pages == 0 {
		pages = 1
	}
	offset := (page - 1) * limit
	threads, err := s.repo.ListThreads(board, offset, limit)
	if err != nil {
		return nil, err
	}
	ids := make([]int64, len(threads))
	for i, t := range threads {
		ids[i] = t.ID
	}
	opsMap := make(map[int64]*model.Post)
	for _, tid := range ids {
		op, err := s.repo.GetOP(tid)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}
		if err == nil {
			opsMap[tid] = op
		}
	}
	repliesMap, err := s.repo.GetRepliesPreview(ids, 3)
	if err != nil {
		return nil, err
	}
	out := make([]model.ThreadWithOP, 0, len(threads))
	for _, t := range threads {
		tw := model.ThreadWithOP{Thread: t, OP: opsMap[t.ID], Replies: repliesMap[t.ID]}
		out = append(out, tw)
	}
	return &model.ThreadsPage{Board: *b, Threads: out, Page: page, Limit: limit, Total: total, Pages: pages}, nil
}

func (s *BoardService) GetThread(board string, id int64, page, limit int) (*model.ThreadPage, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 15
	}
	b, err := s.repo.GetBoard(board)
	if err != nil {
		return nil, err
	}
	t, err := s.repo.GetThread(board, id)
	if err != nil {
		return nil, err
	}
	op, err := s.repo.GetOP(id)
	if err != nil {
		return nil, fmt.Errorf("op: %w", err)
	}
	total, err := s.repo.CountReplies(id)
	if err != nil {
		return nil, err
	}
	pages := (total + limit - 1) / limit
	if pages == 0 {
		pages = 1
	}
	offset := (page - 1) * limit
	replies, err := s.repo.ListReplies(id, offset, limit)
	if err != nil {
		return nil, err
	}
	if replies == nil {
		replies = []model.Post{}
	}
	return &model.ThreadPage{Board: *b, Thread: *t, OP: *op, Replies: replies, Page: page, Limit: limit, Total: total, Pages: pages}, nil
}

func (s *BoardService) CreateThread(board, body string, imageURL, imageName *string, fileSize *int, country, ipHash string) (int64, error) {
	return s.repo.CreateThread(board, body, imageURL, imageName, fileSize, country, ipHash)
}

func (s *BoardService) CreateReply(threadID int64, board, body string, imageURL, imageName *string, fileSize *int, country, ipHash string) (int64, error) {
	return s.repo.CreateReply(threadID, board, body, imageURL, imageName, fileSize, country, ipHash)
}