package services

import (
	"context"
	"fmt"
	"os"
	"modern-social-media/internal/repository"
)

type StoryService struct {
	Repo repository.StoryRepository
}

func NewStoryService(repo repository.StoryRepository) *StoryService {
	return &StoryService{Repo: repo}
}

func (s *StoryService) CleanupExpiredStories(ctx context.Context) error {
	stories, err := s.Repo.GetExpiredStories(ctx, 24)
	if err != nil {
		return fmt.Errorf("failed to get expired stories: %w", err)
	}

	if len(stories) == 0 {
		return nil
	}

	for _, story := range stories {
		if story.MediaURL != "" {
			relativePath := story.MediaURL
			if len(relativePath) > 0 && relativePath[0] == '/' {
				relativePath = relativePath[1:]
			}

			if err := os.Remove(relativePath); err != nil {
				if !os.IsNotExist(err) {
					fmt.Printf("Warning: failed to remove file %s: %v\n", relativePath, err)
				}
			}
		}
	}
	
	return s.Repo.DeleteExpiredStories(ctx, 24)
}
