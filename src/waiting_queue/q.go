package waitingqueue

import (
	"context"

	"github.com/go-redis/redis/v9"
	"github.com/hongminhcbg/waitingroom/src/models"
)

type IWaitingQueue interface {
	Pop(ctx context.Context, n int) ([]*models.UserCtx, error)
	Push(ctx context.Context, u *models.UserCtx) error
	ZRank(ctx context.Context, userId int64) (int, error)
}

type _waitingQueue struct {
	r *redis.Client
}

func (w *_waitingQueue) Pop(ctx context.Context, n int) ([]*models.UserCtx, error) {
	panic("implement me")
}
