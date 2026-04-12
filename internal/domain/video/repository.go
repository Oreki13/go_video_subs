package video

import "context"

type Repository interface {
	FindByTiers(ctx context.Context, tiers []string) ([]Video, error)
}
