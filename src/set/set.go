package set

import "github.com/hongminhcbg/waitingroom/src/models"

type ISet interface {
	Get(userId string) (*models.UserCtx, error)
	Add(u *models.UserCtx) error
	Del(userId string) error
}
