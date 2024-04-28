package router

import (
	"github.com/gin-gonic/gin"
	"github.com/hongminhcbg/waitingroom/src/service"
)

func InitGin(e *gin.Engine, s *service.Service) {
	e.POST("/api/v1/users", s.CreateUser)
	e.POST("/api/v1/slot-check", s.SlotCheck)
	e.POST("/api/v1/slot-release", s.SlotRelease)
}
