package video

import (
	"context"
	"fmt"

	"github.com/go_video_subs/internal/domain/video"
	"github.com/jmoiron/sqlx"
)

type videoRepository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) video.Repository {
	return &videoRepository{db: db}
}

func (r *videoRepository) FindByTiers(ctx context.Context, tiers []string) ([]video.Video, error) {
	if len(tiers) == 0 {
		return []video.Video{}, nil
	}

	query, args, err := sqlx.In(`
		SELECT v.id, v.category_id, v.title, v.description, v.url, v.duration_seconds, v.created_at, v.updated_at
		FROM videos v
		JOIN video_categories vc ON v.category_id = vc.id
		WHERE vc.name IN (?)
		ORDER BY v.created_at DESC
	`, tiers)
	if err != nil {
		return nil, fmt.Errorf("repository: build find videos by tiers query: %w", err)
	}

	query = r.db.Rebind(query)

	videos := make([]video.Video, 0)
	if err := r.db.SelectContext(ctx, &videos, query, args...); err != nil {
		return nil, fmt.Errorf("repository: find videos by tiers: %w", err)
	}
	return videos, nil
}
