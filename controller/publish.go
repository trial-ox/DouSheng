package controller

import (
	"fmt"
	"github.com/RaymondCode/simple-demo/service"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
)

type VideoListResponse struct {
	Response
	VideoList []service.Video `json:"video_list"`
}

// Publish check token then save upload file to public directory
func Publish(c *gin.Context) {
	token := c.PostForm("token")

	if _, exist := usersLoginInfo[token]; !exist {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "User doesn't exist"})
		return
	}

	data, err := c.FormFile("data")
	if err != nil {
		c.JSON(http.StatusOK, Response{
			StatusCode: 1,
			StatusMsg:  err.Error(),
		})
		return
	}

	filename := filepath.Base(data.Filename)
	user := usersLoginInfo[token]
	finalName := fmt.Sprintf("%d_%s", user.Id, filename)
	saveFile := filepath.Join("./public/", finalName)
	if err := c.SaveUploadedFile(data, saveFile); err != nil {
		c.JSON(http.StatusOK, Response{
			StatusCode: 1,
			StatusMsg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		StatusCode: 0,
		StatusMsg:  finalName + " uploaded successfully",
	})
}

// PublishList all users have same publish video list
func PublishList(c *gin.Context) {
	// 获取访问的用户id
	user_id, _ := c.GetQuery("user_id")
	userId, _ := strconv.ParseInt(user_id, 10, 64)
	log.Printf("获取到用户id:%v\n", userId)

	// 获取当前用户id，这里直接写死
	//curId, _ := strconv.ParseInt(c.GetString("userId"), 10, 64)
	var curId int64 = 2
	log.Printf("获取到当前用户id:%v\n", curId)

	videoService := service.VideoServiceImpl{}
	list, err := videoService.List(userId, curId)
	if err != nil {
		log.Printf("调用videoService.List(%v)出现错误：%v\n", userId, err)
		c.JSON(http.StatusOK, VideoListResponse{
			Response:  Response{StatusCode: 1, StatusMsg: "获取视频列表失败"},
			VideoList: nil,
		})
		return
	}

	log.Printf("调用videoService.List(%v)成功", userId)
	c.JSON(http.StatusOK, VideoListResponse{
		Response: Response{
			StatusCode: 0,
		},
		VideoList: list,
	})

}
