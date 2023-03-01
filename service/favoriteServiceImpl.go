package service

import (
	"errors"
	"github.com/RaymondCode/simple-demo/config"
	"github.com/RaymondCode/simple-demo/dao"
	"github.com/RaymondCode/simple-demo/middleware/redis"
	"log"
	"strconv"
	"sync"
	"time"
)

type FavoriteServiceImpl struct {
	VideoService
	UserService
}

func (favorite *FavoriteServiceImpl) IsFavourite(videoId int64, userId int64) (bool, error) {
	strUserId := strconv.FormatInt(userId, 10)
	strVideoId := strconv.FormatInt(videoId, 10)
	//查询Redis key：strUserId中是否存在value:videoId
	if num, err := redis.RdbLikeUserId.Exists(redis.Ctx, strUserId).Result(); num > 0 {

		if err != nil {
			log.Printf("方法:IsFavourite RedisLikeUserId失败：%v", err)
			return false, err
		}
		exist, err1 := redis.RdbLikeUserId.SIsMember(redis.Ctx, strUserId, videoId).Result()
		if err != nil {
			log.Printf("方法IsFavourite RedisLikeUserId失败：%v", err1)

		}
		return exist, nil
	} else {
		//LikeUserId不存在key,查询Redis LikeVideoId,逻辑同上
		if num, err := redis.RdbLikeVideoId.Exists(redis.Ctx, strVideoId).Result(); num > 0 {
			if err != nil {
				log.Printf("方法IsFavourite RedisLikeVideoId失败：%v", err)
				return false, err
			}
			exist, err1 := redis.RdbLikeVideoId.SIsMember(redis.Ctx, strVideoId, userId).Result()

			if err1 != nil {
				log.Printf("方法IsFavourite RedisLikeVideoId 失败：%v", err1)
				return false, err1
			}
			return exist, nil
		} else {
			//都没有，创建set，添加一个默认值,维持set不为空，避免缓存击穿
			if err := NewRedisSet(strUserId); err != nil {
				return false, err
			}
			//LikeUserId LikeVideoId中都没有对应key,通过userId查询likes表,返回所有点赞videoId，并维护到Redis LikeUserId
			videoList, err1 := dao.GetLikeVideoIdList(userId)
			if err1 != nil {
				log.Printf(err1.Error())
				return false, err1
			}
			for _, likeVideoId := range videoList {
				redis.RdbLikeUserId.SAdd(redis.Ctx, strUserId, likeVideoId)
			}
			//加入后再查询redis
			exist, err2 := redis.RdbLikeUserId.SIsMember(redis.Ctx, strUserId, videoId).Result()
			if err2 != nil {
				log.Printf("方法:IsFavourite RedisLikeUserId 失败：%v", err2)
				return false, err2
			}
			return exist, nil
		}
	}

}

//根据视频id获取点赞数量
func (favorite *FavoriteServiceImpl) FavouriteCount(videoId int64) (int64, error) {
	strVideoId := strconv.FormatInt(videoId, 10)
	//查询redis
	if num, err := redis.RdbLikeVideoId.Exists(redis.Ctx, strVideoId).Result(); num > 0 {
		if err != nil {
			log.Printf("方法:FavouriteCount RedisLikeVideoId失败：%v", err)
			return 0, err
		}
		//获取userId数量
		count, err1 := redis.RdbLikeVideoId.SCard(redis.Ctx, strVideoId).Result()
		if err1 != nil {
			log.Printf("方法:FavouriteCount RedisLikeVideoId失败：%v", err1)
			return 0, err1
		}
		return count - 1, nil //去掉初始化的默认元素
	} else {
		//redis中没有videoId的set，设置初始化，逻辑同上方法
		if err := NewRedisSet(strVideoId); err != nil {
			return 0, err
		}
		//通过videoId查表，返回信息并维护到redis中
		userIdList, err1 := dao.GetLikeUserIdList(videoId)
		if err1 != nil {
			log.Printf(err1.Error())
			return 0, err1
		}
		for _, likeUserId := range userIdList {
			redis.RdbLikeVideoId.SAdd(redis.Ctx, strVideoId, likeUserId)
		}

		count, err2 := redis.RdbLikeVideoId.SCard(redis.Ctx, strVideoId).Result()
		if err2 != nil {
			log.Printf("方法:FavouriteCount RedisLikeVideoId失败：%v", err2)
			return 0, err2
		}
		return count - 1, nil
	}
}

// FavouriteAction 根据userId，videoId,actionType对视频进行点赞或者取消赞操作;
func (favorite *FavoriteServiceImpl) FavouriteAction(userId int64, videoId int64, actionType int32) error {
	strUserId := strconv.FormatInt(userId, 10)

	strVideoId := strconv.FormatInt(videoId, 10)
	//更新redis
	if actionType == config.LikeAction {
		//查询redis
		if num, err := redis.RdbLikeUserId.Exists(redis.Ctx, strUserId).Result(); num > 0 {
			if err != nil {
				log.Printf("方法:FavouriteAction RedisLikeUserId失败：%v", err)
				return err
			}
			//优先更新redis，确保redis里的数据为正确
			if _, err1 := redis.RdbLikeUserId.SAdd(redis.Ctx, strUserId, videoId).Result(); err1 != nil {
				log.Printf("方法:FavouriteAction RedisLikeUserId 失败：%v", err1)
				return err1
			} else {
				//若redis更新无异常再来更新操作数据库

				//如果查询没有数据，用来生成该条点赞信息，存储在likeData中
				var likeData dao.Like
				//先查询是否有这条数据
				likeInfo, err := dao.GetLikeInfo(userId, videoId)

				if err != nil {
					log.Printf(err.Error())

				} else {
					if likeInfo == (dao.Like{}) { //没查到这条数据，则新建这条数据；
						likeData.UserId = userId        //插入userId
						likeData.VideoId = videoId      //插入videoId
						likeData.Cancel = config.IsLike //插入点赞cancel=0

						if err := dao.InsertLike(likeData); err != nil {
							log.Printf(err.Error())

						}
					} else { //查到这条数据,更新即可;
						//如果有问题，说明插入数据库失败，打印错误信息err:"update data fail"
						if err := dao.UpdateLike(userId, videoId, config.IsLike); err != nil {
							log.Printf(err.Error())
						}
					}
				}
			}
		}
		//若redis中不存在，维护redis，新建key:strUserId

		if err := NewRedisSet(strUserId); err != nil {
			return err
		}
		//查询数据库，添加到redis中
		videoIdList, err1 := dao.GetLikeVideoIdList(userId)
		if err1 != nil {
			return err1
		}

		for _, likeVideoId := range videoIdList {
			if _, err1 := redis.RdbLikeUserId.SAdd(redis.Ctx, strUserId, likeVideoId).Result(); err1 != nil {
				log.Printf("方法:FavouriteAction RedisLikeUserId add value失败")
				redis.RdbLikeUserId.Del(redis.Ctx, strUserId)
				return err1
			}
		}
		//redis添加无异常后修改数据库
		if _, err2 := redis.RdbLikeUserId.SAdd(redis.Ctx, strUserId, videoId).Result(); err2 != nil {
			log.Printf("方法:FavouriteAction RedisLikeUserId add value失败：%v", err2)
			return err2
		} else {
			//如果查询没有数据，用来生成该条点赞信息，存储在likeData中
			var likeData dao.Like
			//先查询是否有这条数据
			likeInfo, err := dao.GetLikeInfo(userId, videoId)

			if err != nil {
				log.Printf(err.Error())

			} else {
				if likeInfo == (dao.Like{}) { //没查到这条数据，则新建这条数据；
					likeData.UserId = userId        //插入userId
					likeData.VideoId = videoId      //插入videoId
					likeData.Cancel = config.IsLike //插入点赞cancel=0

					if err := dao.InsertLike(likeData); err != nil {
						log.Printf(err.Error())

					}
				} else { //查到这条数据,更新即可;
					//如果有问题，说明插入数据库失败，打印错误信息err:"update data fail"
					if err := dao.UpdateLike(userId, videoId, config.IsLike); err != nil {
						log.Printf(err.Error())
					}
				}
			}
		}
		//初始化维护 redis key：strVideoId
		//查询是否已经加载
		if n, err := redis.RdbLikeVideoId.Exists(redis.Ctx, strVideoId).Result(); n > 0 {

			if err != nil {
				log.Printf("方法:FavouriteAction RedisLikeVideoId query key失败：%v", err)
				return err
			} //如果加载过此信息key:strVideoId，则加入value:userId
			//如果redis LikeVideoId 添加失败，返回错误信息
			if _, err1 := redis.RdbLikeVideoId.SAdd(redis.Ctx, strVideoId, userId).Result(); err1 != nil {
				log.Printf("方法:FavouriteAction RedisLikeVideoId add value失败：%v", err1)
				return err1
			}
		} else {
			//不存在 strVideoId，新建，逻辑同上
			if err := NewRedisSet(strVideoId); err != nil {
				return err
			}
			userIdList, err1 := dao.GetLikeUserIdList(videoId)
			//如果有问题，说明查询失败，返回错误信息："get likeUserIdList failed"
			if err1 != nil {
				return err1
			}
			//遍历userIdList,添加进key的集合中，若失败，删除key，并返回错误信息，这么做的原因是防止脏读，
			//保证redis与mysql数据一致性
			for _, likeUserId := range userIdList {
				if _, err1 := redis.RdbLikeVideoId.SAdd(redis.Ctx, strVideoId, likeUserId).Result(); err1 != nil {
					log.Printf("方法:FavouriteAction RedisLikeVideoId失败")
					redis.RdbLikeVideoId.Del(redis.Ctx, strVideoId)
					return err1
				}
			}
			//这样操作理由同上
			if _, err2 := redis.RdbLikeVideoId.SAdd(redis.Ctx, strVideoId, userId).Result(); err2 != nil {
				log.Printf("方法:FavouriteAction RedisLikeVideoId失败：%v", err2)
				return err2
			}
		}
	} else {
		//取消点赞
		//查询key：strUserId是否存在
		if n, err := redis.RdbLikeUserId.Exists(redis.Ctx, strUserId).Result(); n > 0 {

			if err != nil {
				log.Printf("方法:FavouriteAction RedisLikeUserId query key失败：%v", err)
				return err
			} //防止出现redis数据不一致情况，当redis删除操作成功，才执行数据库更新操作
			if _, err1 := redis.RdbLikeUserId.SRem(redis.Ctx, strUserId, videoId).Result(); err1 != nil {
				log.Printf("方法:FavouriteAction RedisLikeUserId del value失败：%v", err1)
				return err1
			} else {
				//操作数据库
				likeInfo, err := dao.GetLikeInfo(userId, videoId)
				//如果有问题，说明查询数据库失败，返回错误信息err:"get likeInfo failed"
				if err != nil {
					log.Printf(err.Error())
				} else {
					if likeInfo == (dao.Like{}) { //只有当前是点赞状态才能取消点赞这个行为
						// 所以如果查询不到数据则返回错误信息:"can't find data,this action invalid"
						log.Printf(errors.New("can't find data,this action invalid").Error())
					} else {
						//如果查询到数据，则更新为取消赞状态
						//如果有问题，说明插入数据库失败
						if err := dao.UpdateLike(userId, videoId, config.Unlike); err != nil {
							log.Printf(err.Error())
						}
					}
				}

			}
		} else {
			//redis不存在key：userId，则新建，逻辑同上
			if err := NewRedisSet(strUserId); err != nil {
				return err
			}
			videoIdList, err1 := dao.GetLikeVideoIdList(userId)
			//如果有问题，说明查询失败，返回错误信息："get likeVideoIdList failed"
			if err1 != nil {
				return err1
			}
			//遍历videoIdList,添加进key的集合中，若失败，删除key，并返回错误信息，这么做的原因是防止脏读，
			//保证redis与mysql 数据原子性
			for _, likeVideoId := range videoIdList {
				if _, err1 := redis.RdbLikeUserId.SAdd(redis.Ctx, strUserId, likeVideoId).Result(); err1 != nil {
					log.Printf("方法:FavouriteAction RedisLikeUserId add value失败")
					redis.RdbLikeUserId.Del(redis.Ctx, strUserId)
					return err1
				}
			}
			//这样操作理由同上
			if _, err2 := redis.RdbLikeUserId.SRem(redis.Ctx, strUserId, videoId).Result(); err2 != nil {
				log.Printf("方法:FavouriteAction RedisLikeUserId del value失败：%v", err2)
				return err2
			} else {
				//操作数据库
				likeInfo, err := dao.GetLikeInfo(userId, videoId)
				//如果有问题，说明查询数据库失败，返回错误信息err:"get likeInfo failed"
				if err != nil {
					log.Printf(err.Error())
				} else {
					if likeInfo == (dao.Like{}) { //只有当前是点赞状态才能取消点赞这个行为
						// 所以如果查询不到数据则返回错误信息:"can't find data,this action invalid"
						log.Printf(errors.New("can't find data,this action invalid").Error())
					} else {
						//如果查询到数据，则更新为取消赞状态
						//如果有问题，说明插入数据库失败
						if err := dao.UpdateLike(userId, videoId, config.Unlike); err != nil {
							log.Printf(err.Error())
						}
					}
				}
			}
		}
		//查询Redis LikeVideoId(key:strVideoId)是否已经加载过此信息
		if n, err := redis.RdbLikeVideoId.Exists(redis.Ctx, strVideoId).Result(); n > 0 {
			//如果有问题，说明查询redis失败,返回错误信息
			if err != nil {
				log.Printf("方法:FavouriteAction RedisLikeVideoId query key失败：%v", err)
				return err
			} //如果加载过此信息key:strVideoId，则删除value:userId
			//如果redis LikeVideoId 删除失败，返回错误信息
			if _, err1 := redis.RdbLikeVideoId.SRem(redis.Ctx, strVideoId, userId).Result(); err1 != nil {
				log.Printf("方法:FavouriteAction RedisLikeVideoId del value失败：%v", err1)
				return err1
			}
		} else { //redis不存在key：videoId，则新建，逻辑同上
			if err := NewRedisSet(strVideoId); err != nil {
				return err
			}
			//if _, err := redis.RdbLikeVideoId.SAdd(redis.Ctx, strVideoId, -1).Result(); err != nil {
			//	log.Printf("方法:FavouriteAction RedisLikeVideoId add value失败")
			//	redis.RdbLikeVideoId.Del(redis.Ctx, strVideoId)
			//	return err
			//}
			////给键值设置有效期
			//_, err := redis.RdbLikeVideoId.Expire(redis.Ctx, strVideoId,
			//	time.Duration(config.OneMonth)*time.Second).Result()
			//if err != nil {
			//	log.Printf("方法:FavouriteAction RedisLikeVideoId 设置有效期失败")
			//	redis.RdbLikeVideoId.Del(redis.Ctx, strVideoId)
			//	return err
			//}

			userIdList, err1 := dao.GetLikeUserIdList(videoId)
			//如果有问题，说明查询失败，返回错误信息："get likeUserIdList failed"
			if err1 != nil {
				redis.RdbLikeVideoId.Del(redis.Ctx, strVideoId)
				return err1
			}
			//遍历userIdList,添加进key的集合中，若失败，删除key，并返回错误信息，这么做的原因是防止脏读，
			//保证redis与mysql数据一致性
			for _, likeUserId := range userIdList {
				if _, err1 := redis.RdbLikeVideoId.SAdd(redis.Ctx, strVideoId, likeUserId).Result(); err1 != nil {
					log.Printf("方法:FavouriteAction RedisLikeVideoId失败")
					redis.RdbLikeVideoId.Del(redis.Ctx, strVideoId)
					return err1
				}
			}
			//这样操作理由同上
			if _, err2 := redis.RdbLikeVideoId.SRem(redis.Ctx, strVideoId, userId).Result(); err2 != nil {
				log.Printf("方法:FavouriteAction RedisLikeVideoId del value失败：%v", err2)
				return err2
			}
		}
	}
	return nil
}

// GetFavouriteList 据userId，curId(当前用户Id),返回userId的点赞列表;
func (favorite *FavoriteServiceImpl) GetFavouriteList(userId int64, curId int64) ([]Video, error) {
	strUserId := strconv.FormatInt(userId, 10)
	//查询redis UserId,如果key：strUserId存在,则获取全部videoId
	if n, err := redis.RdbLikeUserId.Exists(redis.Ctx, strUserId).Result(); n > 0 {
		if err != nil {
			log.Printf("方法:GetFavouriteList RedisLikeVideoId query key失败：%v", err)
			return nil, err
		}
		videoIdList, err1 := redis.RdbLikeUserId.SMembers(redis.Ctx, strUserId).Result()
		if err1 != nil {
			log.Printf("方法:GetFavouriteList RedisLikeVideoId get values失败：%v", err1)
			return nil, err1
		}
		favoriteVideoList := new([]Video)
		//采用协程并发添加video到集合中去
		length := len(videoIdList) - 1 //减去默认设置的-1
		if length == 0 {
			return *favoriteVideoList, nil
		}
		var wg sync.WaitGroup
		wg.Add(length) //将WaitGroup的计数值设为video的数量
		for i := 0; i <= length; i++ {
			videoId, _ := strconv.ParseInt(videoIdList[i], 10, 64)
			if videoId == -1 { //跳过默认的-1
				continue
			}
			go favorite.addFavouriteVideoList(videoId, curId, favoriteVideoList, &wg)
		}
		wg.Wait()
		return *favoriteVideoList, nil
	} else {
		//redis中没有userId的key，则查询数据库
		//新建set key：userId
		if err := NewRedisSet(strUserId); err != nil {
			return nil, err
		}
		videoIdList, err1 := dao.GetLikeVideoIdList(userId)
		if err1 != nil {
			log.Println(err1.Error())
			redis.RdbLikeUserId.Del(redis.Ctx, strUserId)
			return nil, err1
		}
		//保证redis与mysql数据一致性
		for _, likeVideoId := range videoIdList {
			if _, err2 := redis.RdbLikeUserId.SAdd(redis.Ctx, strUserId, likeVideoId).Result(); err2 != nil {
				log.Printf("方法:GetFavouriteList RedisLikeUserId add value失败")
				redis.RdbLikeUserId.Del(redis.Ctx, strUserId)
				return nil, err2
			}
		}
		//采用协程并发添加video到集合中去
		favoriteVideoList := new([]Video)
		length := len(videoIdList) //减去默认设置的-1
		if length == 0 {
			return *favoriteVideoList, nil
		}
		var wg sync.WaitGroup
		wg.Add(length) //将WaitGroup的计数值设为video的数量
		for i := 0; i <= length; i++ {

			if videoIdList[i] == -1 { //跳过默认的-1
				continue
			}
			go favorite.addFavouriteVideoList(videoIdList[i], curId, favoriteVideoList, &wg)
		}
		wg.Wait()
		return *favoriteVideoList, nil
	}

}

// TotalFavourite 获取用户被点赞总数
func (favorite *FavoriteServiceImpl) TotalFavourite(userId int64) (int64, error) {
	videoIdList, err := favorite.GetVideoIdList(userId)
	if err != nil {
		log.Printf(err.Error())
		return 0, err
	}
	var sum int64 //该用户的总被点赞数
	//提前开辟空间,存取每个视频的点赞数
	videoLikeCountList := new([]int64)
	//采用协程并发将对应videoId的点赞数添加到集合中去
	i := len(videoIdList)
	var wg sync.WaitGroup
	wg.Add(i)
	for j := 0; j < i; j++ {
		go favorite.addVideoLikeCount(videoIdList[j], videoLikeCountList, &wg)
	}
	wg.Wait()
	//遍历累加，求总被点赞数
	for _, count := range *videoLikeCountList {
		sum += count
	}
	return sum, nil
}

//FavouriteVideoCount 根据userId获取这个用户点赞视频数量
func (favorite *FavoriteServiceImpl) FavouriteVideoCount(userId int64) (int64, error) {
	strUserId := strconv.FormatInt(userId, 10)

	if num, err := redis.RdbLikeUserId.Exists(redis.Ctx, strUserId).Result(); num > 0 {
		if err != nil {
			log.Printf("方法:FavouriteVideoCount RdbLikeUserId query key失败：%v", err)
			return 0, err
		} else {
			count, err1 := redis.RdbLikeUserId.SCard(redis.Ctx, strUserId).Result()
			if err1 != nil {
				log.Printf("方法:FavouriteVideoCount RdbLikeUserId query count 失败：%v", err1)
				return 0, err1
			}
			return count - 1, nil //去掉DefaultRedisValue

		}
	} else {
		if err := NewRedisSet(strUserId); err != nil {
			return 0, err
		}
		videoIdList, err1 := dao.GetLikeVideoIdList(userId)
		if err1 != nil {
			log.Printf(err1.Error())
			return 0, err1
		}
		//维护Redis LikeUserId(key:strUserId)，遍历videoIdList加入
		for _, likeVideoId := range videoIdList {
			if _, err1 := redis.RdbLikeUserId.SAdd(redis.Ctx, strUserId, likeVideoId).Result(); err1 != nil {
				log.Printf("方法:FavouriteVideoCount RedisLikeUserId add value失败")
				redis.RdbLikeUserId.Del(redis.Ctx, strUserId)
				return 0, err1
			}
		}
		count, err2 := redis.RdbLikeUserId.SCard(redis.Ctx, strUserId).Result()
		if err2 != nil {
			log.Printf("方法:FavouriteVideoCount RdbLikeUserId query count 失败：%v", err2)
			return 0, err2
		}
		return count - 1, nil
	}
}

// NewRedisSet NewRedisSet 在查询redis中的set时若不存在该key，则新建并设默认值防止缓存击穿
func NewRedisSet(strId string) error {
	if _, err := redis.RdbLikeVideoId.SAdd(redis.Ctx, strId, -1).Result(); err != nil {
		log.Printf("方法:FavouriteAction RedisLikeVideoId add value失败")
		redis.RdbLikeVideoId.Del(redis.Ctx, strId)
		return err
	}
	//给键值设置有效期
	_, err := redis.RdbLikeVideoId.Expire(redis.Ctx, strId,
		time.Duration(config.OneMonth)*time.Second).Result()
	if err != nil {
		log.Printf("方法:FavouriteAction RedisLikeVideoId 设置有效期失败")
		redis.RdbLikeVideoId.Del(redis.Ctx, strId)
		return err
	}
	return nil
}

//addFavouriteVideoList 根据videoId,登录用户curId，添加视频对象到点赞列表空间
func (favorite *FavoriteServiceImpl) addFavouriteVideoList(videoId int64, curId int64, favoriteVideoList *[]Video, wg *sync.WaitGroup) {
	defer wg.Done()
	//调用videoService接口，GetVideo：根据videoId，当前用户id:curId，返回Video类型对象
	video, err := favorite.GetVideo(videoId, curId)
	//如果没有获取这个video_id的视频，视频可能被删除了,打印异常,并且不加入此视频
	if err != nil {
		log.Println(errors.New("this favourite video is miss"))
		return
	}
	//将Video类型对象添加到集合中去
	*favoriteVideoList = append(*favoriteVideoList, video)
}

//addVideoLikeCount 根据videoId，将该视频点赞数加入对应提前开辟好的空间内
func (favorite *FavoriteServiceImpl) addVideoLikeCount(videoId int64, videoLikeCountList *[]int64, wg *sync.WaitGroup) {
	defer wg.Done()
	//调用FavouriteCount：根据videoId,获取点赞数
	count, err := favorite.FavouriteCount(videoId)
	if err != nil {

		log.Printf(err.Error())
		return
	}
	*videoLikeCountList = append(*videoLikeCountList, count)
}

//GetLikeService 解决likeService调videoService,videoService调userService,useService调likeService循环依赖的问题
func GetLikeService() FavoriteServiceImpl {
	var userService UserServiceImpl
	var videoService VideoServiceImpl
	var favoriteService FavoriteServiceImpl
	userService.FavoriteService = &favoriteService
	favoriteService.VideoService = &videoService
	videoService.UserService = &userService
	return favoriteService
}
