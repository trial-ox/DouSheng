package service

import (
	"sync"

	"github.com/RaymondCode/simple-demo/dao"
)

type FollowServiceImp struct {
	UserService
}

var (
	followServiceImp  *FollowServiceImp //controller层通过该实例变量调用service的所有业务方法。
	followServiceOnce sync.Once         //限定该service对象为单例，节约内存。
)

// 生成并返回FollowServiceImp结构体单例变量。
func NewFSIInstance() *FollowServiceImp {
	followServiceOnce.Do(
		func() {
			followServiceImp = &FollowServiceImp{
				UserService: &UserServiceImpl{
					FollowService: &FollowServiceImp{},
				},
			}
		})
	return followServiceImp
}

//根据当前用户id和目标用户id来判断当前用户是否关注了目标用户
func (*FollowServiceImp) IsFollowing(userId int64, targetId int64) (bool, error) {
	// SQL 查询。
	relation, err := dao.NewFollowDaoInstance().FindRelation(userId, targetId)

	if nil != err {
		return false, err
	}
	if nil == relation {
		return false, nil
	}
	return true, nil
}

//根据用户id来查询用户被多少其他用户关注
func (*FollowServiceImp) GetFollowerCnt(userId int64) (int64, error) {

	// SQL中查询。
	ids, err := dao.NewFollowDaoInstance().GetFollowersIds(userId)
	if nil != err {
		return 0, err
	}
	return int64(len(ids)), err
}

// GetFollowingCnt 给定当前用户id，查询其关注者数量。
func (*FollowServiceImp) GetFollowingCnt(userId int64) (int64, error) {
	// 用SQL查询。
	ids, err := dao.NewFollowDaoInstance().GetFollowingIds(userId)
	if nil != err {
		return 0, err
	}
	return int64(len(ids)), err
}

// GetFollowing 根据当前用户id来查询他的关注者列表。
func (f *FollowServiceImp) getFollowing(userId int64) ([]User, error) {
	// 获取关注对象的id数组。
	ids, err := dao.NewFollowDaoInstance().GetFollowingIds(userId)
	// 查询出错
	if nil != err {
		return nil, err
	}
	// 没得关注者
	if nil == ids {
		return nil, nil
	}
	// 根据每个id来查询用户信息。
	len := len(ids)
	if len > 0 {
		len -= 1
	}
	var wg sync.WaitGroup
	wg.Add(len)
	users := make([]User, len)
	i, j := 0, 0
	for ; i < len; j++ {
		if ids[j] == -1 {
			continue
		}
		go func(i int, idx int64) {
			defer wg.Done()
			users[i], _ = f.GetUserByIdWithCurId(idx, userId)
		}(i, ids[i])
		i++
	}
	wg.Wait()
	// 返回关注对象列表。
	return users, nil
}

//根据当前用户id来查询他的粉丝列表。
func (f *FollowServiceImp) GetFollowers(userId int64) ([]User, error) {
	return getFollowers(userId)
}

// 从数据库查所有粉丝信息。
func getFollowers(userId int64) ([]User, error) {
	users := make([]User, 1)

	if err := dao.Db.Raw("select T.id,T.name,T.follow_cnt follow_count,T.follower_cnt follower_count,if(f.cancel is null,'false','true') is_follow"+
		"\nfrom follows f right join"+
		"\n(select fid,id,`name`,"+
		"\ncount(if(tag = 'follower' and cancel is not null,1,null)) follower_cnt,"+
		"\ncount(if(tag = 'follow' and cancel is not null,1,null)) follow_cnt"+
		"\nfrom("+
		"\nselect f1.user_id fid,u.id,`name`,f2.cancel,'follower' tag"+
		"\nfrom follows f1 join users u on f1.follower_id = u.id and f1.cancel = 0"+
		"\nleft join follows f2 on u.id = f2.user_id and f2.cancel = 0"+
		"\nunion all"+
		"\nselect f1.user_id fid,u.id,`name`,f2.cancel,'follow' tag"+
		"\nfrom follows f1 join users u on f1.follower_id = u.id and f1.cancel = 0"+
		"\nleft join follows f2 on u.id = f2.follower_id and f2.cancel = 0"+
		"\n) T group by fid,id,`name`"+
		"\n) T on f.user_id = T.id and f.follower_id = T.fid and f.cancel = 0 where fid = ?", userId).
		Scan(&users).Error; nil != err {
		// 查询出错。
		return nil, err
	}
	// 查询成功。
	return users, nil
}
