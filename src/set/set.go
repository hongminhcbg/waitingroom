package set

import "github.com/hongminhcbg/waitingroom/src/models"

type ISet interface {
	Get(userId int64) (*models.UserCtx, error)
	Add(u *models.UserCtx) error
	Del(userId int64) error
}
