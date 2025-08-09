package service

import (
	"context"
	"log/slog"
	"slices"
	"time"

	"github.com/elangreza/content-management-system/internal/params"
)

type (
	// tagRepo defines the methods that the tag repository must implement.
	tagRepo interface {
		UpsertTags(ctx context.Context, names ...string) error
		GetTags(ctx context.Context) ([]string, error)
		GetTagUsageCounts(ctx context.Context) (map[string]int, error)
		GetTagLastUsage(ctx context.Context) (map[string]time.Time, error)
	}

	TagService struct {
		tagRepo           tagRepo
		tagUsageCounts    map[string]int
		tagLastUsed       map[string]time.Time
		tagTrendingScores map[string]float64
	}
)

func NewTagService(tagRepo tagRepo) *TagService {
	ts := &TagService{
		tagRepo:           tagRepo,
		tagUsageCounts:    make(map[string]int),
		tagLastUsed:       make(map[string]time.Time),
		tagTrendingScores: make(map[string]float64),
	}

	go ts.periodicGetTagUsageCount()
	go ts.periodicGetTagLastUsedAndCalculateTrendingScore()

	return ts
}

func (s *TagService) periodicGetTagUsageCount() {

	// run for the first time immediately
	counts, err := s.getTagUsageCounts(context.Background())
	if err != nil {
		slog.Error("failed to get tag usage counts", "error", err)
	}
	if counts != nil {
		s.tagUsageCounts = counts
	}

	newTicker := time.NewTicker(10 * time.Second)
	defer newTicker.Stop()
	for {
		select {
		case <-newTicker.C:
			counts, err := s.getTagUsageCounts(context.Background())
			if err != nil {
				slog.Error("failed to get tag usage counts", "error", err)
			}
			if counts != nil {
				s.tagUsageCounts = counts
			}
		}
	}
}

func (s *TagService) periodicGetTagLastUsedAndCalculateTrendingScore() {
	time.Sleep(5 * time.Second) // Initial delay to allow periodicGetTagUsageCount to run

	// run for the first time immediately
	tagLastUsed, err := s.getTagLastUsed(context.Background())
	if err != nil {
		slog.Error("failed to get tag last used", "error", err)
	}
	if tagLastUsed != nil {
		s.tagLastUsed = tagLastUsed
	}
	s.calculateTrendingScore()

	newTicker := time.NewTicker(10 * time.Second)
	defer newTicker.Stop()
	for {
		select {
		case <-newTicker.C:
			tagLastUsed, err := s.getTagLastUsed(context.Background())
			if err != nil {
				slog.Error("failed to get tag last used", "error", err)
			}
			if tagLastUsed != nil {
				s.tagLastUsed = tagLastUsed
			}

			s.calculateTrendingScore()
		}
	}
}

func (s *TagService) CreateTag(ctx context.Context, tagNames ...string) error {
	for _, tagName := range tagNames {
		if err := s.tagRepo.UpsertTags(ctx, tagName); err != nil {
			return err
		}
	}
	return nil
}

func (s *TagService) GetTags(ctx context.Context) ([]params.GetTagResponse, error) {
	tags, err := s.tagRepo.GetTags(ctx)
	if err != nil {
		return nil, err
	}

	var responses []params.GetTagResponse
	for _, tag := range tags {
		responses = append(responses, params.GetTagResponse{
			Name:          tag,
			UsageCount:    s.tagUsageCounts[tag],
			TrendingScore: s.tagTrendingScores[tag],
		})
	}

	slices.SortFunc(responses, func(a, b params.GetTagResponse) int {
		if a.UsageCount == b.UsageCount {
			return 0
		} else if a.UsageCount > b.UsageCount {
			return -1
		}
		return 1
	})

	return responses, nil
}

func (s *TagService) getTagUsageCounts(ctx context.Context) (map[string]int, error) {
	// timeout must be less than the periodic ticker duration
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	counts, err := s.tagRepo.GetTagUsageCounts(ctx)
	if err != nil {
		return nil, err
	}

	return counts, nil
}

func (s *TagService) getTagLastUsed(ctx context.Context) (map[string]time.Time, error) {
	// timeout must be less than the periodic ticker duration
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	counts, err := s.tagRepo.GetTagLastUsage(ctx)
	if err != nil {
		return nil, err
	}

	return counts, nil
}

func (s *TagService) calculateTrendingScore() {
	now := time.Now()
	interval := 24 * time.Hour
	for tag, usage := range s.tagUsageCounts {
		lastUsed, ok := s.tagLastUsed[tag]
		if !ok {
			// If never used, set a low score
			s.tagTrendingScores[tag] = 0
			continue
		}

		if lastUsed.Add(interval).Before(now) {
			s.tagTrendingScores[tag] = 0
			continue
		}

		score := float64(usage) / float64(1+interval.Hours())
		s.tagTrendingScores[tag] = score
	}
}
