package service

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v9"
	"github.com/hongminhcbg/waitingroom/src/models"
)

type SlotCheckRequest struct {
	UserID int64 `json:"user_id"`
}

type SlotCheckResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Rank      int    `json:"rank,omitempty"`
		SlotToken string `json:"slot_token,omitempty"`
	} `json:"data,omitempty"`
}

func (s *Service) slotCheckInQueue(ctx *gin.Context, userId int64) {
	rank, err := s.queue.ZRank(ctx.Request.Context(), userId)
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
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

func (s *Service) onboardingNewUser(ctx context.Context, userId int64) (rank int, err error) {
	err = s.queue.Push(ctx, &models.UserCtx{
		UserId:         userId,
		TsEnroll:       time.Now().UnixMilli(),
		TsInActivePool: 0,
	})

	if err != nil {
		return
	}

	rank, err = s.queue.ZRank(ctx, userId)
	return
}

func (s *Service) SlotCheck(ctx *gin.Context) {
	req := &SlotCheckRequest{}
	err := ctx.BindJSON(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := s.activeSession.Get(req.UserID)
	if errors.Is(err, redis.Nil) {
		// in active user, check rank in queue
		s.slotCheckInQueue(ctx, req.UserID)
		return
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if time.Now().Unix()-u.TsInActivePool > 10*60 {
		// expired in active user, check rank in queue
		err := s.activeSession.Del(req.UserID)
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
	ctx.JSON(http.StatusOK, gin.H{"message": "slot release"})
}
