package service

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v9"
	"github.com/hongminhcbg/waitingroom/src/models"
	"github.com/hongminhcbg/waitingroom/src/utils"
)

type SlotReq struct {
	UserID string `json:"user_id"`
}

type SlotCheckResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Rank      int    `json:"rank,omitempty"`
		SlotToken string `json:"slot_token,omitempty"`
	} `json:"data,omitempty"`
}

func (s *Service) slotCheckInQueue(ctx *gin.Context, userId string) {
	rank, err := s.queue.ZRank(ctx.Request.Context(), userId)
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		s.log.Info("user not in queue, onboarding new user", "user_id", userId)
		rank, err := s.onboardingNewUser(ctx.Request.Context(), userId)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, SlotCheckResponse{
			Code:    1,
			Message: "success",
			Data: struct {
				Rank      int    `json:"rank,omitempty"`
				SlotToken string `json:"slot_token,omitempty"`
			}{
				Rank:      rank,
				SlotToken: "",
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, SlotCheckResponse{
		Code:    1,
		Message: "success",
		Data: struct {
			Rank      int    `json:"rank,omitempty"`
			SlotToken string `json:"slot_token,omitempty"`
		}{
			Rank:      rank,
			SlotToken: "",
		},
	})
	return
}

func (s *Service) onboardingNewUser(ctx context.Context, userId string) (rank int, err error) {
	u := &models.UserCtx{
		UserId:         userId,
		TsEnroll:       time.Now().UnixMilli(),
		TsInActivePool: 0,
	}

	err = utils.SaveUserByUserId(s.redis, ctx, u)
	if err != nil {
		s.log.Error(err, "save user by user id error")
		return
	}

	err = s.queue.Zadd(ctx, float64(u.TsEnroll), u.UserId)
	if err != nil {
		return
	}

	rank, err = s.queue.ZRank(ctx, u.UserId)
	if err != nil {
		s.log.Error(err, "get rank in queue error")
		return
	}

	rank += 1
	s.log.Info("add user to queue success", "user_id", u.UserId, "rank", rank)
	return
}

func (s *Service) SlotCheck(ctx *gin.Context) {
	req := &SlotReq{}
	err := ctx.BindJSON(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err = s.activeSession.ZRank(ctx.Request.Context(), req.UserID)
	if errors.Is(err, redis.Nil) {
		// in active user, check rank in queue
		s.log.Info("user not in active pool, check in queue", "user_id", req.UserID)
		s.slotCheckInQueue(ctx, req.UserID)
		return
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	u, err := utils.LoadUserByUserId(s.redis, ctx.Request.Context(), req.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if time.Now().Unix()-u.TsInActivePool > 10*60 {
		// expired in active user, check rank in queue
		err := s.activeSession.Zrem(ctx.Request.Context(), req.UserID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		rank, err := s.onboardingNewUser(ctx.Request.Context(), req.UserID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, SlotCheckResponse{
			Code:    1,
			Message: "success",
			Data: struct {
				Rank      int    `json:"rank,omitempty"`
				SlotToken string `json:"slot_token,omitempty"`
			}{
				Rank:      rank,
				SlotToken: "",
			},
		})

		return
	}

	token := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d:%d", u.UserId, u.TsInActivePool)))
	ctx.JSON(http.StatusOK, SlotCheckResponse{
		Code:    1,
		Message: "success",
		Data: struct {
			Rank      int    `json:"rank,omitempty"`
			SlotToken string `json:"slot_token,omitempty"`
		}{
			Rank:      0,
			SlotToken: token,
		},
	})

	return
}

func (s *Service) SlotRelease(ctx *gin.Context) {
	req := &SlotReq{}
	err := ctx.BindJSON(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func(ctx context.Context, userId string) {
		defer wg.Done()
		err := s.queue.Zrem(ctx, userId)
		if err != nil {
			s.log.Error(err, "remove user from queue error")
		}
	}(ctx.Request.Context(), req.UserID)
	go func(ctx context.Context, userId string) {
		defer wg.Done()
		err := s.activeSession.Zrem(ctx, userId)
		if err != nil {
			s.log.Error(err, "remove user from queue error")
		}
	}(ctx.Request.Context(), req.UserID)
	wg.Wait()

	ctx.JSON(http.StatusOK, gin.H{"message": "success"})
}
