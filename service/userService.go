package service

import (
	"github.com/RaymondCode/simple-demo/dao"
)

type UserService interface {
	GetUserList() []dao.TableUser
	GetUserByName(name string) dao.TableUser
	GetUserById(id int64) (User, error)
	InsertUser(user *dao.TableUser) bool
}
type User struct {
	Id            int64  `json:"id,omitempty"`
	Name          string `json:"name,omitempty"`
	FollowCount   int64  `json:"follow_count,omitempty"`
	FollowerCount int64  `json:"follower_count,omitempty"`
	IsFollow      bool   `json:"is_follow,omitempty"`
}
