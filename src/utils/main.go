package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/hongminhcbg/waitingroom/src/models"
)

func LoadUserByUserId(r *redis.Client, ctx context.Context, userId string) (*models.UserCtx, error) {
	d, err := r.Get(ctx, fmt.Sprintf("user:%s", userId)).Result()
	if err != nil {
		return nil, err
	}

	a := &models.UserCtx{}
	err = json.Unmarshal([]byte(d), a)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func SaveUserByUserId(r *redis.Client, ctx context.Context, u *models.UserCtx) error {
	b, err := json.Marshal(u)
	if err != nil {
		return err
	}

	return r.Set(ctx, fmt.Sprintf("user:%s", u.UserId), b, 10*time.Minute).Err()
}
