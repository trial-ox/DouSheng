package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/RaymondCode/simple-demo/dao"
	"github.com/RaymondCode/simple-demo/service"
	"github.com/gin-gonic/gin"
)

type CommentListResponse struct {
	StatusCode  int32                 `json:"status_code"`
	StatusMsg   string                `json:"status_msg,omitempty"`
	CommentList []service.CommentInfo `json:"comment_list,omitempty"`
}

type CommentActionResponse struct {
	StatusCode int32               `json:"status_code"`
	StatusMsg  string              `json:"status_msg,omitempty"`
	Comment    service.CommentInfo `json:"comment,omitempty"`
}

func CommentAction(c *gin.Context) {
	//获取userId
	id, _ := c.Get("userId")
	userid, _ := id.(string)
	userId, err := strconv.ParseInt(userid, 10, 64)
	//错误处理
	if err != nil {
		c.JSON(http.StatusOK, CommentActionResponse{
			StatusCode: -1,
			StatusMsg:  "comment userId json invalid",
		})
		return
	}
	//获取videoId
	videoId, err := strconv.ParseInt(c.Query("video_id"), 10, 64)
	//错误处理
	if err != nil {
		c.JSON(http.StatusOK, CommentActionResponse{
			StatusCode: -1,
			StatusMsg:  "comment videoId json invalid",
		})
		return
	}
	//获取操作类型
	actionType, err := strconv.ParseInt(c.Query("action_type"), 10, 32)
	//错误处理
	if err != nil || actionType < 1 || actionType > 2 {
		c.JSON(http.StatusOK, CommentActionResponse{
			StatusCode: -1,
			StatusMsg:  "comment actionType json invalid",
		})
		return
	}
	//调用service层评论函数
	commentService := new(service.CommentServiceImpl)
	if actionType == 1 { //actionType为1，则进行发表评论操作
		content := c.Query("comment_text")

		//发表评论数据准备
		var sendComment dao.Comment
		sendComment.UserId = userId
		sendComment.VideoId = videoId
		sendComment.CommentText = content
		timeNow := time.Now()
		sendComment.CreateDate = timeNow
		//发表评论
		commentInfo, err := commentService.Send(sendComment)
		//发表评论失败
		if err != nil {
			c.JSON(http.StatusOK, CommentActionResponse{
				StatusCode: -1,
				StatusMsg:  "send comment failed",
			})
			return
		}
		//发表评论成功:
		//返回结果
		c.JSON(http.StatusOK, CommentActionResponse{
			StatusCode: 0,
			StatusMsg:  "send comment success",
			Comment:    commentInfo,
		})
		return
	} else { //actionType为2，则进行删除评论操作
		//获取要删除的评论的id
		commentId, err := strconv.ParseInt(c.Query("comment_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusOK, CommentActionResponse{
				StatusCode: -1,
				StatusMsg:  "delete commentId invalid",
			})
			return
		}
		//删除评论操作
		err = commentService.DelComment(commentId)
		if err != nil { //删除评论失败
			str := err.Error()
			c.JSON(http.StatusOK, CommentActionResponse{
				StatusCode: -1,
				StatusMsg:  str,
			})
			return
		}
		//删除评论成功
		c.JSON(http.StatusOK, CommentActionResponse{
			StatusCode: 0,
			StatusMsg:  "delete comment success",
		})
		return
	}
}

// 查看评论列表 comment/list/
func CommentList(c *gin.Context) {
	//获取userId
	id, _ := c.Get("userId")
	userid, _ := id.(string)
	userId, err := strconv.ParseInt(userid, 10, 64)

	//获取videoId
	videoId, err := strconv.ParseInt(c.Query("video_id"), 10, 64)
	//错误处理
	if err != nil {
		c.JSON(http.StatusOK, CommentListResponse{
			StatusCode: -1,
			StatusMsg:  "comment videoId json invalid",
		})
		return
	}

	//调用service层评论函数
	commentService := new(service.CommentServiceImpl)
	commentList, err := commentService.GetList(videoId, userId)
	if err != nil {
		c.JSON(http.StatusOK, CommentListResponse{
			StatusCode: -1,
			StatusMsg:  err.Error(),
		})
		return
	}
	//获取评论列表成功
	c.JSON(http.StatusOK, CommentListResponse{
		StatusCode:  0,
		StatusMsg:   "get comment list success",
		CommentList: commentList,
	})
	return
}
