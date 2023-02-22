package controller

import (
	"github.com/RaymondCode/simple-demo/dao"
	"github.com/RaymondCode/simple-demo/service"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
)

// usersLoginInfo use map to store user info, and key is username+password for demo
// user data will be cleared every time the server starts
// test data: username=zhanglei, password=douyin

var userIdSequence = int64(1)

// 返回的用户信息
type UserLoginResponse struct {
	Response
	UserId int64  `json:"user_id,omitempty"`
	Token  string `json:"token"`
}

type UserResponse struct {
	Response
	service.User `json:"user"`
}

var usersLoginInfo = map[string]User{
	"zhangleidouyin": {
		Id:            1,
		Name:          "zhanglei",
		FollowCount:   10,
		FollowerCount: 5,
		IsFollow:      true,
	},
}

// douyin/user/register/ 用户注册
func Register(c *gin.Context) {
	username := c.Query("username")
	password := c.Query("password")

	userService := service.UserServiceImpl{}
	un := userService.GetUserByName(username)
	if un.Name == username {
		c.JSON(http.StatusOK, UserLoginResponse{
			Response: Response{
				StatusCode: 1,
				StatusMsg:  "该用户名已使用",
			},
		})
	} else {
		user := dao.TableUser{
			Name:     username,
			Password: service.GetSha256(password),
		}
		success := userService.InsertUser(&user)
		if !success {
			println("插入失败")
		}
		u := userService.GetUserByName(username)
		token := service.NewToken(username)
		log.Println("注册用户的id：", u.Id)
		c.JSON(http.StatusOK, UserLoginResponse{
			Response: Response{StatusCode: 0, StatusMsg: username + "注册成功"},
			UserId:   u.Id,
			Token:    token,
		})
	}
}

// douyin/user/login/ 用户登录
func Login(c *gin.Context) {
	username := c.Query("username")
	password := c.Query("password")

	userService := service.UserServiceImpl{}
	user := userService.GetUserByName(username)

	if service.GetSha256(password) == user.Password {
		token := service.NewToken(username)
		c.JSON(http.StatusOK, UserLoginResponse{
			Response: Response{StatusCode: 0, StatusMsg: "登录成功"},
			UserId:   user.Id,
			Token:    token,
		})
	} else {
		c.JSON(http.StatusOK, UserLoginResponse{
			Response: Response{StatusCode: 1, StatusMsg: "用户名或密码错误"},
		})
	}

}

func UserInfo(c *gin.Context) {

	userid, _ := strconv.ParseInt(c.Query("user_id"), 10, 64)

	userService := service.UserServiceImpl{
		FavoriteService: &service.FavoriteServiceImpl{},
	}

	if user, err := userService.GetUserById(userid); err != nil {
		c.JSON(http.StatusOK, UserResponse{
			Response: Response{StatusCode: 1, StatusMsg: "用户不存在"},
		})
	} else {
		c.JSON(http.StatusOK, UserResponse{
			Response: Response{StatusCode: 0},
			User:     user,
		})
	}
}
