package controller

import (
	"fmt"
	"github.com/RaymondCode/simple-demo/dao"
	"github.com/RaymondCode/simple-demo/service"
	"github.com/RaymondCode/simple-demo/utils"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"log"
	"net/http"
	"strconv"
	"time"
)

type VideoListResponse struct {
	Response
	VideoList []service.Video `json:"video_list"`
}

// Publish check token then save upload file to public directory
func Publish(c *gin.Context) {
	title := c.PostForm("title")
	userId, _ := strconv.ParseInt(c.GetString("userId"), 10, 64)
	log.Printf("获取到用户id:%v\n", userId)

	data, err := c.FormFile("data")

	if err != nil {
		c.JSON(http.StatusOK, Response{
			StatusCode: 1,
			StatusMsg:  err.Error(),
		})
		return
	}

	//生成视频地址
	videoUUID, _ := uuid.NewV4()
	videoDir := time.Now().Format("2006-01-02") + "/" + videoUUID.String() + ".mp4"
	videoUrl := "https://" + "douyin-ljy" + ".oss-cn-beijing.aliyuncs.com/" + videoDir
	fmt.Println("上传视频地址是" + videoUrl)
	//生成图片地址
	pictureUUID, _ := uuid.NewV4()
	pictureDir := time.Now().Format("2006-01-02") + "/" + pictureUUID.String() + ".jpg"
	coverUrl := "https://" + "douyin-ljy" + ".oss-cn-beijing.aliyuncs.com/" + pictureDir
	fmt.Println("上传视频封面的地址是" + coverUrl)

	//开启协程上传
	go func() {
		//上传视频
		_ = utils.UploadVideo(videoDir, data)
		//time.Sleep(2*time.Second)
		//获取封面
		coverBytes, _ := utils.ReadFrameAsJpeg(videoUrl)
		//上传封面
		_ = utils.UploadPicture(pictureDir, coverBytes)
	}()

	err2 := dao.Save(videoUrl, coverUrl, userId, title)
	if err2 != nil {
		log.Printf("方法videoService.Publish(data, userId) 失败：%v", err)
		c.JSON(http.StatusOK, Response{
			StatusCode: 1,
			StatusMsg:  err.Error(),
		})
		return
	}
	log.Printf("方法videoDao save 成功")

	c.JSON(http.StatusOK, Response{
		StatusCode: 0,
		StatusMsg:  "uploaded successfully",
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

	videoService := GetVideo()
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
