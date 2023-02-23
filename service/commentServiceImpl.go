package service

import (
	"log"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/RaymondCode/simple-demo/config"
	"github.com/RaymondCode/simple-demo/dao"
	"github.com/RaymondCode/simple-demo/middleware/redis"
)

//引入userservice
type CommentServiceImpl struct {
	UserService
}

//使用video id 查询Comment数量
func (c CommentServiceImpl) CountFromVideoId(videoId int64) (int64, error) {
	//先在缓存中查视频id
	cnt, err := redis.RdbVCid.SCard(redis.Ctx, strconv.FormatInt(videoId, 10)).Result()
	if err != nil { //若查询缓存出错，则打印log
		//return 0, err
		log.Println("count from redis error:", err)
	}
	//缓存中查到了数量，则返回数量值-1（去除0值）
	if cnt != 0 {
		return cnt - 1, nil
	}
	//缓存中查不到则去请求数据库
	cntDao, err1 := dao.Count(videoId)
	if err1 != nil {
		log.Println("comment count dao err:", err1)
		return 0, nil
	}
	//将评论id切片存入redis
	go func() {
		//查询评论id-list
		cList, _ := dao.CommentIdList(videoId)
		//先在redis中存储一个值,防止脏读
		_, _err := redis.RdbVCid.SAdd(redis.Ctx, strconv.Itoa(int(videoId)), config.DefaultRedisValue).Result()
		if _err != nil { //若存储redis失败
			return
		}
		//设置key值过期时间
		_, err := redis.RdbVCid.Expire(redis.Ctx, strconv.Itoa(int(videoId)),
			time.Duration(config.OneMonth)*time.Second).Result()
		if err != nil {
			log.Println("redis save one vId - cId expire failed")
		}
		//评论id循环存入redis
		for _, commentId := range cList {
			insertRedisVideoCommentId(strconv.Itoa(int(videoId)), commentId)
		}
	}()
	//返回结果
	return cntDao, nil
}

//在redis中存储video_id对应的comment_id
func insertRedisVideoCommentId(videoId string, commentId string) {
	//在redis-RdbVCid中存储video_id对应的comment_id
	_, err := redis.RdbVCid.SAdd(redis.Ctx, videoId, commentId).Result()
	if err != nil { //若存储redis失败-1 err，则直接删除key
		redis.RdbVCid.Del(redis.Ctx, videoId)
		return
	}
	//在redis-RdbCVid中存储comment_id对应的video_id
	_, err = redis.RdbCVid.Set(redis.Ctx, commentId, videoId, 0).Result()
	if err != nil {
		log.Println("redis save one cId - vId failed")
	}
}

// 发表评论接口实现
func (c CommentServiceImpl) Send(comment dao.Comment) (CommentInfo, error) {
	//初始化数据
	var commentInfo dao.Comment
	commentInfo.VideoId = comment.VideoId         //评论视频id传入
	commentInfo.UserId = comment.UserId           //评论用户id传入
	commentInfo.CommentText = comment.CommentText //评论内容传入
	commentInfo.Cancel = config.ValidComment      //评论状态=0(有效),=1(无效)
	commentInfo.CreateDate = comment.CreateDate   //评论时间
	//发送添加评论
	commentRtn, err := dao.SendComment(commentInfo)
	//若不成功则返回提示报错
	if err != nil {
		return CommentInfo{}, err
	}
	//查询用户信息
	impl := UserServiceImpl{
		FollowService: &FollowServiceImp{},
	}
	userData, err2 := impl.GetUserByIdWithCurId(comment.UserId, comment.UserId)
	if err2 != nil {
		return CommentInfo{}, err2
	}
	//3.拼接
	commentData := CommentInfo{
		Id:         commentRtn.Id,
		UserInfo:   userData,
		Content:    commentRtn.CommentText,
		CreateDate: commentRtn.CreateDate.Format(config.DateTime),
	}
	//返回结果
	return commentData, nil
}

// 3、删除评论，传入评论id
func (c CommentServiceImpl) DelComment(commentId int64) error {
	//不在内存中，则直接走数据库删除
	return dao.DeleteComment(commentId)
}

// 查看评论列表-返回评论list
func (c CommentServiceImpl) GetList(videoId int64, userId int64) ([]CommentInfo, error) {
	commentList, err := dao.GetCommentList(videoId)
	if err != nil {
		return nil, err
	}
	//当前有0条评论
	if commentList == nil {
		return nil, nil
	}

	//提前定义好切片长度
	commentInfoList := make([]CommentInfo, len(commentList))

	wg := &sync.WaitGroup{}
	wg.Add(len(commentList))
	idx := 0
	for _, comment := range commentList {
		//2.调用方法组装评论信息，再append
		var commentData CommentInfo
		//将评论信息进行组装，添加想要的信息,插入从数据库中查到的数据
		go func(comment dao.Comment) {
			oneComment(&commentData, &comment, userId)
			//3.组装list
			//commentInfoList = append(commentInfoList, commentData)
			commentInfoList[idx] = commentData
			idx = idx + 1
			wg.Done()
		}(comment)
	}
	wg.Wait()
	//评论排序-按照主键排序
	sort.Sort(CommentSlice(commentInfoList))
	return commentInfoList, nil
}

//此函数用于给一个评论赋值：评论信息+用户信息 填充
func oneComment(comment *CommentInfo, com *dao.Comment, userId int64) {
	var wg sync.WaitGroup
	wg.Add(1)
	//根据评论用户id和当前用户id，查询评论用户信息
	impl := UserServiceImpl{
		FollowService: &FollowServiceImp{},
	}
	var err error
	comment.Id = com.Id
	comment.Content = com.CommentText
	comment.CreateDate = com.CreateDate.Format(config.DateTime)
	comment.UserInfo, err = impl.GetUserByIdWithCurId(com.UserId, userId)
	if err != nil {
		log.Println("CommentService-GetList: GetUserByIdWithCurId return err: " + err.Error()) //函数返回提示错误信息
	}
	wg.Done()
	wg.Wait()
}

// CommentSlice 此变量以及以下三个函数都是做排序-准备工作
type CommentSlice []CommentInfo

func (a CommentSlice) Len() int { //重写Len()方法
	return len(a)
}
func (a CommentSlice) Swap(i, j int) { //重写Swap()方法
	a[i], a[j] = a[j], a[i]
}
func (a CommentSlice) Less(i, j int) bool { //重写Less()方法
	return a[i].Id > a[j].Id
}
