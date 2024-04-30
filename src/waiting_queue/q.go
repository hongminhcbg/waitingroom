package waitingqueue

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/go-redis/redis/v9"
	"github.com/hongminhcbg/waitingroom/src/models"
	"github.com/hongminhcbg/waitingroom/src/utils"
)

type ISortedSet interface {
	Pop(ctx context.Context, n int) ([]*models.UserCtx, error)
	Zadd(ctx context.Context, score float64, userId string) error
	ZRank(ctx context.Context, userId string) (int, error)
	Len(ctx context.Context) (int, error)
	Zcard(ctx context.Context) (int, error)
	Zrem(ctx context.Context, userId string) error
}

type _waitingQueue struct {
	r       *redis.Client
	log     logr.Logger
	setName string
}

func NewWaitingQueue(r *redis.Client, log logr.Logger, setName string) ISortedSet {
	return &_waitingQueue{
		r:       r,
		log:     log,
		setName: setName,
	}
}

func (w *_waitingQueue) Pop(ctx context.Context, n int) ([]*models.UserCtx, error) {
	users := w.r.ZRange(ctx, w.setName, 0, int64(n-1)).Val()
	if len(users) == 0 {
		return nil, redis.Nil
	}

	ans := make([]*models.UserCtx, 0, len(users))
	for _, userid := range users {
		u, err := utils.LoadUserByUserId(w.r, ctx, userid)
		if err != nil {
			w.log.Error(err, "load user by user id error")
			continue
		}

		ans = append(ans, u)
	}

	err := w.r.ZRem(ctx, w.setName, users).Err()
	if err != nil {
		return nil, err // redis die
	}

	return ans, nil
}

func (w *_waitingQueue) Zadd(ctx context.Context, score float64, userId string) error {
	return w.r.ZAdd(ctx, w.setName, redis.Z{
		Score:  score,
		Member: userId,
	}).Err()
}

func (w *_waitingQueue) Len(ctx context.Context) (int, error) {
	a, err := w.r.ZCount(ctx, w.setName, "0", "+inf").Result()
	if err != nil {
		return 0, err
	}

	return int(a), nil
}

func (w *_waitingQueue) ZRank(ctx context.Context, userId string) (int, error) {
	a, err := w.r.ZRank(ctx, w.setName, userId).Result()
	if err != nil {
		return 0, err
	}

	return int(a), nil
}

func (w *_waitingQueue) Zcard(ctx context.Context) (int, error) {
	a, err := w.r.ZCard(ctx, w.setName).Result()
	if err != nil {
		return 0, err
	}

	return int(a), nil
}

func (w *_waitingQueue) Zrem(ctx context.Context, userId string) error {
	err := w.r.ZRem(ctx, w.setName, userId).Err()
	if err != nil {
		return err
	}

	return nil
}
