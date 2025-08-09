package service

import (
	"context"
	"log/slog"
	"sort"
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
		tagPairs          [][2]string
	}
)

func NewTagService(tagRepo tagRepo) *TagService {
	ts := &TagService{
		tagRepo:           tagRepo,
		tagUsageCounts:    make(map[string]int),
		tagLastUsed:       make(map[string]time.Time),
		tagTrendingScores: make(map[string]float64),
		tagPairs:          make([][2]string, 0),
	}

	go ts.periodicGetTagUsageCount()
	go ts.periodicGetTagLastUsedAndCalculateTrendingScore()
	go ts.periodicGetTagPairs()

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
	time.Sleep(3 * time.Second) // Initial delay to allow periodicGetTagUsageCount to run

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

func (s *TagService) periodicGetTagPairs() {
	time.Sleep(6 * time.Second) // Initial delay to allow periodicGetTagLastUsedAndCalculateTrendingScore to run

	// run for the first time immediately
	tagPairs, err := s.getTagPairs(context.Background())
	if err != nil {
		slog.Error("failed to get tag pairs", "error", err)
	}
	if tagPairs != nil {
		s.tagPairs = tagPairs
	}

	newTicker := time.NewTicker(10 * time.Second)
	defer newTicker.Stop()
	for {
		select {
		case <-newTicker.C:
			tagPairs, err := s.getTagPairs(context.Background())
			if err != nil {
				slog.Error("failed to get tag pairs", "error", err)
			}
			if tagPairs != nil {
				s.tagPairs = tagPairs
			}
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

func (s *TagService) GetTags(ctx context.Context, req params.GetTagsRequest) ([]params.GetTagResponse, error) {
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

	sort.Slice(responses, func(i, j int) bool {
		switch req.SortValue {
		case "usage_count":
			if req.Direction == "asc" {
				return responses[i].UsageCount < responses[j].UsageCount
			}
			return responses[i].UsageCount > responses[j].UsageCount
		case "trending_score":
			if req.Direction == "asc" {
				return responses[i].TrendingScore < responses[j].TrendingScore
			}
			return responses[i].TrendingScore > responses[j].TrendingScore
		case "name":
			if req.Direction == "asc" {
				return responses[i].Name < responses[j].Name
			}
			return responses[i].Name > responses[j].Name
		}
		return false
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

func (s *TagService) getTagPairs(ctx context.Context) ([][2]string, error) {
	// timeout must be less than the periodic ticker duration
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	tags, err := s.GetTags(ctx, params.GetTagsRequest{
		SortValue: "name",
		Direction: "asc",
	})
	if err != nil {
		return nil, err
	}

	var pairs [][2]string
	for i := 0; i < len(tags); i++ {
		for j := i + 1; j < len(tags); j++ {
			pairs = append(pairs, [2]string{tags[i].Name, tags[j].Name})
		}
	}

	return pairs, nil
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
