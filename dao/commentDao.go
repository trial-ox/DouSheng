package dao

import (
	"errors"
	"log"
	"time"

	"github.com/RaymondCode/simple-demo/config"
)

// 评论信息-数据库中的结构体
type Comment struct {
	Id          int64     //评论id
	UserId      int64     //评论用户id
	VideoId     int64     //视频id
	CommentText string    //评论内容
	CreateDate  time.Time //评论发布的日期mm-dd
	Cancel      int32     //取消评论为1，发布评论为0
}

// 完成数据库表名称的映射
func (Comment) TableName() string {
	return "comments"
}

//获取评论数量
func Count(videoId int64) (int64, error) {
	//Init()
	var count int64
	//数据库中查询评论数量
	err := Db.Model(Comment{}).Where(map[string]interface{}{"video_id": videoId, "cancel": config.ValidComment}).Count(&count).Error
	if err != nil {
		return -1, errors.New("find comments count failed")
	}
	return count, nil
}

//CommentIdList 根据视频id获取评论id 列表
func CommentIdList(videoId int64) ([]string, error) {
	var commentIdList []string
	err := Db.Model(Comment{}).Select("id").Where("video_id = ?", videoId).Find(&commentIdList).Error
	if err != nil {
		log.Println("CommentIdList:", err)
		return nil, err
	}
	return commentIdList, nil
}

// send 发表评论
func SendComment(comment Comment) (Comment, error) {
	//添加评论信息
	err := Db.Model(Comment{}).Create(&comment).Error
	if err != nil {
		//返回错误信息
		return Comment{}, errors.New("insert comment into database failed")
	}
	return comment, nil
}

//删除评论
func DeleteComment(id int64) error {
	//声明Comment
	var commentInfo Comment
	//查询是否为null评论
	result := Db.Model(Comment{}).Where(map[string]interface{}{"id": id, "cancel": config.ValidComment}).First(&commentInfo)
	//查询到此评论数量为0
	//则返回无此评论
	if result.RowsAffected == 0 {
		//提示错误
		return errors.New("need delete comment is not exist")
	}
	//删除评论
	//设置更新评论状态为 1
	err := Db.Model(Comment{}).Where("id = ?", id).Update("cancel", config.InvalidComment).Error
	if err != nil {
		//返回错误信息
		return errors.New("delete comment failed")
	}
	//执行成功
	return nil
}

// 根据视频id查询所属评论全部列表信息
func GetCommentList(videoId int64) ([]Comment, error) {

	//声明列表用于获取数据库中查询的评论信息
	var commentList []Comment
	result := Db.Model(Comment{}).Where(map[string]interface{}{"video_id": videoId, "cancel": config.ValidComment}).
		Order("create_date desc").Find(&commentList)
	//若此视频没有评论信息
	//返回
	if result.RowsAffected == 0 {
		//不存在视频下的评论
		return nil, nil
	}
	//获取评论出错
	if result.Error != nil {
		return commentList, errors.New("get comment list from database failed")
	}
	//获取评论成功
	return commentList, nil
}
