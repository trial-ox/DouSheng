package controller

import (
	"github.com/RaymondCode/simple-demo/service"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"time"
)

type FeedResponse struct {
	Response
	VideoList []service.Video `json:"video_list,omitempty"`
	NextTime  int64           `json:"next_time,omitempty"`
}

// Feed same demo video list for every request
func Feed(c *gin.Context) {
	inputTime := c.Query("latest_time")
	log.Printf("传入的时间" + inputTime)
	var lastTime time.Time
	if inputTime != "" {
		me, _ := strconv.ParseInt(inputTime, 10, 64)
		lastTime = time.Unix(me, 0)
	} else {
		lastTime = time.Now()
	}
	log.Printf("获取到时间戳%v", lastTime)
	userId, _ := strconv.ParseInt(c.GetString("userId"), 10, 64)
	log.Printf("获取到用户id:%v\n", userId)
	// videoService := GetVideo()
	var videoService service.VideoServiceImpl
	feed, nextTime, err := videoService.Feed(lastTime, userId)
	if err != nil {
		log.Printf("方法videoService.Feed(lastTime, userId) 失败：%v", err)
		c.JSON(http.StatusOK, FeedResponse{
			Response: Response{StatusCode: 1, StatusMsg: "获取视频流失败"},
		})
		return
	}
	log.Printf("test======== %v", feed)
	log.Printf("方法videoService.Feed(lastTime, userId) 成功")
	log.Printf("final test %v")
	c.JSON(http.StatusOK, FeedResponse{
		Response:  Response{StatusCode: 0},
		VideoList: feed,
		NextTime:  nextTime.Unix(),
	})
}

// GetVideo 拼装videoService
func GetVideo() service.VideoServiceImpl {
	var userService service.UserServiceImpl
	var videoService service.VideoServiceImpl
	videoService.UserService = &userService
	// var followService service.FollowServiceImp

	var favoriteService service.FavoriteServiceImpl
	// var commentService service.CommentServiceImpl
	// userService.FollowService = &followService
	userService.FavoriteService = &favoriteService
	// followService.UserService = &userService
	favoriteService.VideoService = &videoService
	// commentService.UserService = &userService
	// videoService.CommentService = &commentService
	videoService.FavoriteService = &favoriteService
	return videoService
}
