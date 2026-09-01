package service

import (
	"216chan/backend/internal/model"
	"216chan/backend/internal/repository"
)

type StatsService struct {
	repo *repository.StatsRepo
}

func NewStatsService(repo *repository.StatsRepo) *StatsService {
	return &StatsService{repo: repo}
}

func (s *StatsService) GetStats() (*model.Stats, error) {
	return s.repo.GetStats()
}
