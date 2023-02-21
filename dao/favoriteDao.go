package dao

import (
	"errors"
	"github.com/RaymondCode/simple-demo/config"
	"log"
)

// Like 表的结构。
type Like struct {
	Id      int64 //自增主键
	UserId  int64 //点赞用户id
	VideoId int64 //视频id
	Cancel  int8  //是否点赞，0为点赞，1为取消赞
}

// TableName 修改表名映射
func (Like) TableName() string {
	return "likes"
}

//根据videoId获取点赞userId
func GetLikeUserIdList(videoId int64) ([]int64, error) {

	var likeUserIdList []int64
	err := Db.Model(Like{}).Where(map[string]interface{}{"video_id": videoId, "cancel": config.IsLike}).
		Pluck("user_id", &likeUserIdList).Error
	if err != nil {
		log.Println(err.Error())
		return nil, errors.New("获取userId失败")
	} else {
		return likeUserIdList, nil
	}
}

// UpdateLike 根据userId，videoId,actionType点赞或者取消赞
func UpdateLike(userId int64, videoId int64, actionType int32) error {
	//更新当前用户观看视频的点赞状态“cancel”，返回错误结果
	err := Db.Model(Like{}).Where(map[string]interface{}{"user_id": userId, "video_id": videoId}).
		Update("cancel", actionType).Error

	if err != nil {
		log.Println(err.Error())
		return errors.New("更新失败")
	}
	//更新操作成功
	return nil
}

// InsertLike 插入点赞数据
func InsertLike(likeData Like) error {
	//创建点赞数据，默认为点赞，cancel为0，返回错误结果
	err := Db.Model(Like{}).Create(&likeData).Error

	if err != nil {
		log.Println(err.Error())
		return errors.New("插入失败")
	}
	return nil
}

// GetLikeInfo 根据userId,videoId查询点赞信息
func GetLikeInfo(userId int64, videoId int64) (Like, error) {
	//创建一条空like结构体，用来存储查询到的信息
	var likeInfo Like
	//根据userid,videoId查询是否有该条信息，如果有，存储在likeInfo,返回查询结果
	err := Db.Model(Like{}).Where(map[string]interface{}{"user_id": userId, "video_id": videoId}).
		First(&likeInfo).Error
	if err != nil {
		//查询数据为0，打印"can't find data"，返回空结构体，这时候就应该要考虑是否插入这条数据了
		if "record not found" == err.Error() {
			log.Println("can't find data")
			return Like{}, nil
		} else {
			//如果查询数据库失败，返回获取likeInfo信息失败
			log.Println(err.Error())
			return likeInfo, errors.New("get likeInfo failed")
		}
	}
	return likeInfo, nil
}

// GetLikeVideoIdList 根据userId查询所属点赞全部videoId
func GetLikeVideoIdList(userId int64) ([]int64, error) {
	var likeVideoIdList []int64
	err := Db.Where(map[string]interface{}{"user_id": userId, "cancel": config.IsLike}).
		Pluck("video_id", &likeVideoIdList).Error
	if err != nil {
		//查询数据为0，返回空likeVideoIdList切片，以及返回无错误
		if len(likeVideoIdList) == 0 {
			log.Println("查询数据为0")
			return likeVideoIdList, nil
		} else {
			//如果查询数据库失败，返回获取likeVideoIdList失败
			log.Println(err.Error())
			return likeVideoIdList, errors.New("GetLikeVideoIdList查询失败")
		}
	}
	return likeVideoIdList, nil
}
