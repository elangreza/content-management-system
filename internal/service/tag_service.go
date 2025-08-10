package service

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"time"

	"github.com/elangreza/content-management-system/internal/constanta"
	"github.com/elangreza/content-management-system/internal/entity"
	"github.com/elangreza/content-management-system/internal/params"
)

type (
	// tagRepo defines the methods that the tag repository must implement.
	tagRepo interface {
		UpsertTags(ctx context.Context, names ...string) error
		GetTags(ctx context.Context) ([]string, error)
		GetTagUsageCounts(ctx context.Context) (map[string]int, error)
		GetTagLastUsage(ctx context.Context) (map[string]time.Time, error)
		GetTagUsage(ctx context.Context) (map[string]entity.TagUsage, error)
		GetArticleTags(ctx context.Context, status constanta.ArticleVersionStatus) ([]entity.ArticleVersionTag, error)
	}

	TagActionTrigger struct {
		Name    TagServiceAction
		Payload interface{}
	}

	TagService struct {
		articleRepo      articleRepo
		tagRepo          tagRepo
		tagUsage         *SafeMap[string, entity.TagUsage]
		tagPairFrequency *SafeMap[[2]string, int]
		actionTrigger    chan TagActionTrigger
	}
)

type TagServiceAction int8

const (
	// trigger when article version is published or archived
	calculateTagUsageAndPairFrequency TagServiceAction = iota
	// trigger when article version is drafted or published
	calculateArticleTagRelation
)

func NewTagService(articleRepo articleRepo, tagRepo tagRepo) *TagService {
	tagUsage := NewSafeMap[string, entity.TagUsage]()
	tagPairFrequency := NewSafeMap[[2]string, int]()
	ts := &TagService{
		articleRepo:      articleRepo,
		tagRepo:          tagRepo,
		tagUsage:         tagUsage,
		tagPairFrequency: tagPairFrequency,
		actionTrigger:    make(chan TagActionTrigger),
	}

	go ts.tagRoutine()

	return ts
}

func (s *TagService) tagRoutine() {
	// run for the first time immediately
	s.calculateTagUsageAndPairFrequency()

	newTicker := time.NewTicker(10 * time.Second)
	defer newTicker.Stop()
	for {
		select {
		case <-newTicker.C:
			// slog.Info("ticker triggered")
			s.calculateTagUsageAndPairFrequency()
		case action, ok := <-s.actionTrigger:
			if !ok {
				return
			}
			switch action.Name {
			case calculateTagUsageAndPairFrequency:
				slog.Info("calculateTagUsageAndPairFrequency")
				s.calculateTagUsageAndPairFrequency()
			case calculateArticleTagRelation:
				slog.Info("calculateArticleTagRelation")
				s.calculateTagUsageAndPairFrequency()
				if err := s.updateArticleVersionTagRelationshipScore(action.Payload); err != nil {
					slog.Error("failed to update relationship score", "error", err)
					continue
				}
			default:
				slog.Error("unknown action", "action", action.Name)
			}
		}
	}
}

func (s *TagService) updateArticleVersionTagRelationshipScore(actionPayload any) error {

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	payload, ok := actionPayload.(entity.CalculateArticleVersionTagRelationShipScorePayload)
	if !ok {
		return fmt.Errorf("invalid payload for calculateArticleTagRelation action")
	}

	score := s.calculateArticleVersionTagRelationShipScore(payload.Tags)

	if err := s.articleRepo.UpdateArticleVersionRelationshipScore(ctx, payload.ArticleVersionID, score); err != nil {
		return fmt.Errorf("failed to update relationship score: %w", err)
	}

	slog.Info("successfully updated relationship score", "score", score, "article_version_id", payload.ArticleVersionID)

	return nil
}

func (s *TagService) calculateTagUsageAndPairFrequency() {
	var err error
	s.tagUsage, err = s.getTagUsage(context.Background())
	if err != nil {
		slog.Error("failed to get tag usage counts", "error", err)
	}
	s.tagPairFrequency, err = s.getTagPairFrequency(context.Background())
	if err != nil {
		slog.Error("failed to get tag pairs", "error", err)
	}
}

func (s *TagService) CreateTag(ctx context.Context, tagNames ...string) error {
	if err := s.tagRepo.UpsertTags(ctx, tagNames...); err != nil {
		return err
	}

	// recalculate tag usage and pair frequency
	s.CreateTagTrigger(calculateTagUsageAndPairFrequency, nil)

	return nil
}

func (s *TagService) CreateTagTrigger(name TagServiceAction, payload any) {
	go func() {
		s.actionTrigger <- TagActionTrigger{
			Name:    name,
			Payload: payload,
		}
	}()
}

func (s *TagService) GetTags(ctx context.Context, req params.GetTagsRequest) ([]params.GetTagResponse, error) {
	tags, err := s.tagRepo.GetTags(ctx)
	if err != nil {
		return nil, err
	}

	var responses []params.GetTagResponse
	for _, tag := range tags {
		response := params.GetTagResponse{
			Name: tag,
		}
		ok := s.tagUsage.Exist(tag)
		if ok {
			usage := s.tagUsage.Get(tag)
			response.UsageCount = usage.Count
			response.TrendingScore = usage.TrendingScore
			response.LastUsed = usage.LastUsed
		}
		responses = append(responses, response)
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
		case "last_used":
			if req.Direction == "asc" {
				return responses[i].LastUsed.Unix() < responses[j].LastUsed.Unix()
			}
			return responses[i].LastUsed.Unix() > responses[j].LastUsed.Unix()
		}
		return false
	})

	return responses, nil
}

func (s *TagService) getTagUsage(ctx context.Context) (*SafeMap[string, entity.TagUsage], error) {
	// timeout must be less than the periodic ticker duration
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	tagUsages, err := s.tagRepo.GetTagUsage(ctx)
	if err != nil {
		return nil, err
	}

	sm := NewSafeMap[string, entity.TagUsage]()

	now := time.Now()
	interval := 24 * time.Hour
	for tag, usage := range tagUsages {
		if usage.LastUsed.IsZero() || usage.LastUsed.Add(interval).Before(now) {
			usageCopy := usage
			usageCopy.TrendingScore = 0
			sm.Set(tag, usageCopy)
			continue
		}

		score := float64(usage.Count) / float64(1+interval.Hours())
		usageCopy := usage
		usageCopy.TrendingScore = score
		sm.Set(tag, usageCopy)
	}

	return sm, nil
}

func (s *TagService) getTagPairFrequency(ctx context.Context) (*SafeMap[[2]string, int], error) {
	// timeout must be less than the periodic ticker duration
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	articleVersionTags, err := s.tagRepo.GetArticleTags(ctx, constanta.Published)
	if err != nil {
		return nil, err
	}

	tagPairs := NewSafeMap[[2]string, int]()

	for i := 0; i < len(articleVersionTags); i++ {
		for j := i + 1; j < len(articleVersionTags); j++ {
			pair := [2]string{articleVersionTags[i].TagName, articleVersionTags[j].TagName}
			tagPairs.Set(pair, tagPairs.Get(pair)+1)
		}
	}

	return tagPairs, nil
}

func (s *TagService) getTagPair(tags []entity.Tag) [][2]string {
	var pairs [][2]string
	for i := 0; i < len(tags); i++ {
		for j := i + 1; j < len(tags); j++ {
			pairs = append(pairs, [2]string{tags[i].Name, tags[j].Name})
		}
	}

	return pairs
}

func (s *TagService) calculateArticleVersionTagRelationShipScore(tags []entity.Tag) float64 {

	if len(tags) < 2 {
		return 0
	}

	var scoreSum float64
	var validPairs int
	pairs := s.getTagPair(tags)
	for _, pair := range pairs {
		coOccur := float64(s.tagPairFrequency.Get(pair))
		freq1 := float64(s.tagUsage.Get(pair[0]).Count)
		freq2 := float64(s.tagUsage.Get(pair[1]).Count)
		if freq1 == 0 || freq2 == 0 {
			continue
		}
		score := coOccur / math.Sqrt(freq1*freq2)
		scoreSum += score
		validPairs++
	}

	if validPairs == 0 {
		return 0
	}

	finalScore := math.Round(scoreSum/float64(validPairs)*10000) / 10000

	return finalScore
}
