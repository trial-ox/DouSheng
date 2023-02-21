package controller

import (
	"github.com/RaymondCode/simple-demo/service"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
)

type likeResponse struct {
	StatusCode int32  `json:"status_code"`
	StatusMsg  string `json:"status_msg,omitempty"`
}

type GetFavouriteListResponse struct {
	StatusCode int32           `json:"status_code"`
	StatusMsg  string          `json:"status_msg,omitempty"`
	VideoList  []service.Video `json:"video_list,omitempty"`
}

// FavoriteAction no practical effect, just check if token is valid
func FavoriteAction(c *gin.Context) {
	userId, _ := strconv.ParseInt(c.GetString("userId"), 10, 64)
	videoId, _ := strconv.ParseInt(c.Query("video_id"), 10, 64)
	actionType, _ := strconv.ParseInt(c.Query("action_type"), 10, 64)

	favoriteService := new(service.FavoriteServiceImpl)
	err := favoriteService.FavouriteAction(userId, videoId, int32(actionType))
	if err != nil {
		log.Println("FavoriteAction失败: #{err}")
		c.JSON(http.StatusOK, likeResponse{
			StatusCode: 1,
			StatusMsg:  "点赞操作失败",
		})
	} else {
		log.Printf("FavoriteAction成功")
		c.JSON(http.StatusOK, likeResponse{
			StatusCode: 0,
			StatusMsg:  "点赞操作成功",
		})
	}
}

// FavoriteList 获取点赞列表
func FavoriteList(c *gin.Context) {
	strUserId := c.Query("user_id")
	strCurId := c.GetString("userId")
	userId, _ := strconv.ParseInt(strUserId, 10, 64)
	curId, _ := strconv.ParseInt(strCurId, 10, 64)
	like := GetVideo()

	videos, err := like.GetFavouriteList(userId, curId)
	if err == nil {
		log.Printf("方法like.GetFavouriteList(userid) 成功")
		c.JSON(http.StatusOK, GetFavouriteListResponse{
			StatusCode: 0,
			StatusMsg:  "get favouriteList success",
			VideoList:  videos,
		})
	} else {
		log.Printf("方法like.GetFavouriteList(userid) 失败：%v", err)
		c.JSON(http.StatusOK, GetFavouriteListResponse{
			StatusCode: 1,
			StatusMsg:  "get favouriteList fail ",
		})
	}
}
