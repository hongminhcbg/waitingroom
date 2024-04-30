package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/hongminhcbg/waitingroom/config"
	"github.com/hongminhcbg/waitingroom/src/models"
	"github.com/hongminhcbg/waitingroom/src/store"
	"github.com/hongminhcbg/waitingroom/src/utils"
	waitingqueue "github.com/hongminhcbg/waitingroom/src/waiting_queue"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"github.com/go-redis/redis/v9"
	"github.com/robfig/cron/v3"
)

type Service struct {
	store         *store.UserStore
	log           logr.Logger
	activeSession waitingqueue.ISortedSet
	queue         waitingqueue.ISortedSet
	redis         *redis.Client
}

func (s *Service) cronJobPickUpUserToPool() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.log.Info("cron job pick up user to pool")
	totalUserInActiveSession, err := s.activeSession.Zcard(ctx)
	if err != nil {
		s.log.Error(err, "get total user in active session error")
		return
	}

	MAX_USER_IN_POOL := 1000
	if totalUserInActiveSession >= MAX_USER_IN_POOL {
		s.log.Info("active session is full")
		return
	}

	delta := MAX_USER_IN_POOL - totalUserInActiveSession
	for i := 0; i < delta; i++ {
		users, err := s.queue.Pop(ctx, 1)
		if errors.Is(err, redis.Nil) || len(users) == 0 {
			s.log.Info("queue is empty")
			return
		}

		if err != nil {
			s.log.Error(err, "pop user from queue error")
			// TODO: force remove user from queue
			continue
		}

		users[0].TsInActivePool = time.Now().UnixMilli()
		utils.SaveUserByUserId(s.redis, ctx, users[0])
		err = s.activeSession.Zadd(ctx, float64(users[0].TsInActivePool), users[0].UserId)
		if err != nil {
			s.log.Error(err, "add user to queue error")
			// TODO: force remove user from queue
			continue
		}

		s.log.Info("add user to active session success", "user_id", users[0].UserId)
	}
}

func NewService(cfg *config.Config, store *store.UserStore, r *redis.Client, log logr.Logger) *Service {
	queue := waitingqueue.NewWaitingQueue(r, log, "queue")
	activeSession := waitingqueue.NewWaitingQueue(r, log, "active_pool")
	svc := &Service{
		store:         store,
		log:           log,
		activeSession: activeSession,
		queue:         queue,
		redis:         r,
	}

	// Seconds field, required
	c := cron.New(cron.WithSeconds(), cron.WithChain(
		cron.Recover(log), // or use cron.DefaultLogger
	))

	c.AddFunc("*/5 * * * * *", svc.cronJobPickUpUserToPool)
	c.Start()

	return svc
}

func (s *Service) createNewUser(ctx *gin.Context, req *models.CreateUserRequest) {
	if req.ReqId == "" {
		s.log.Info("req_id id empty, auto generate")
		req.ReqId = fmt.Sprint(time.Now().UnixMilli())
	}

	r := models.User{
		Name:      req.Name,
		ReqId:     req.ReqId,
		RetryTime: 0,
		Status:    models.USER_INIT,
	}

	err := s.store.Save(ctx.Request.Context(), &r)
	if err != nil {
		s.log.Error(err, "save to record error")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	s.log.Info("create new user success")
	ctx.JSON(http.StatusOK, r)
}

func (s *Service) handleExistedReqId(ctx *gin.Context, u *models.User, req *models.CreateUserRequest) {
	ctx.JSON(http.StatusOK, u)
}

func (s *Service) CreateUser(ctx *gin.Context) {
	var req models.CreateUserRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		s.log.Error(err, "json unmarshal error")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	if req.ReqId == "" {
		s.log.Info("Will create new user")
		s.createNewUser(ctx, &req)
		return
	}

	u, _ := s.store.GetByReqId(ctx.Request.Context(), req.ReqId)
	if u == nil {
		s.createNewUser(ctx, &req)
		return
	}

	s.handleExistedReqId(ctx, u, &req)
}
