package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/RaymondCode/simple-demo/config"
	"github.com/RaymondCode/simple-demo/dao"
	"github.com/dgrijalva/jwt-go"
)

type UserServiceImpl struct {
	FavoriteService
	FollowService
}

func (usi *UserServiceImpl) GetUserList() []dao.TableUser {
	users, err := dao.GetTableUserList()
	if err != nil {
		log.Println("Err:", err.Error())
		return users
	}
	return users
}

func (usi *UserServiceImpl) GetUserByName(name string) dao.TableUser {
	user, err := dao.GetTableUserByUsername(name)
	if err != nil {
		log.Println("Err:", err.Error())
		log.Println("User Not Found")
		return user
	}
	log.Println("Query User Success")
	return user
}

func (usi *UserServiceImpl) InsertUser(user *dao.TableUser) bool {
	success := dao.InsertTableUser(user)
	return success
}

func (usi *UserServiceImpl) GetUserById(id int64) (User, error) {
	user := User{
		Id:            0,
		Name:          "",
		FollowCount:   0,
		FollowerCount: 0,
		IsFollow:      false,
	}
	tableUser, err := dao.GetTableUserById(id)
	if err != nil {
		log.Println("Err:", err.Error())
		return user, err
	}

	user.Id = tableUser.Id
	user.Name = tableUser.Name
	return user, nil
}

// 加密密码

func GetSha256(str string) string {
	srcByte := []byte(str)
	sha256New := sha256.New()
	sha256Bytes := sha256New.Sum(srcByte)
	sha256String := hex.EncodeToString(sha256Bytes)
	return sha256String
}

// 根据用户信息生成token
func NewToken(username string) string {
	userServiceImpl := new(UserServiceImpl)
	u := userServiceImpl.GetUserByName(username)
	expiresTime := time.Now().Unix() + int64(config.OneDay)
	fmt.Printf("expiresTime: %v\n", expiresTime)

	claims := jwt.StandardClaims{
		Audience:  u.Name,
		ExpiresAt: expiresTime,
		Id:        strconv.FormatInt(u.Id, 10),
		IssuedAt:  time.Now().Unix(),
		Issuer:    "ljy",
		NotBefore: time.Now().Unix(),
	}
	var jwtSecret = []byte(config.Secret)
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(jwtSecret)
	if err == nil {
		println("generate token success\n")
		return token
	} else {
		println("generate token fail\n")
		return "fail"
	}

}

// GetUserById 未登录情况下,根据user_id获得User对象
func (usi *UserServiceImpl) GetUserById(id int64) (User, error) {
	user := User{
		Id:             0,
		Name:           "",
		FollowCount:    0,
		FollowerCount:  0,
		IsFollow:       false,
		TotalFavorited: 0,
		FavoriteCount:  0,
	}
	tableUser, err := dao.GetTableUserById(id)
	if err != nil {
		log.Println("Err:", err.Error())
		log.Println("User Not Found")
		return user, err
	}
	log.Println("Query User Success")
	//followCount, _ := usi.GetFollowingCnt(id)
	//if err != nil {
	//	log.Println("Err:", err.Error())
	//}
	//followerCount, _ := usi.GetFollowerCnt(id)
	//if err != nil {
	//	log.Println("Err:", err.Error())
	//}
	u := GetLikeService() //解决循环依赖
	totalFavorited, _ := u.TotalFavourite(id)
	favoritedCount, _ := u.FavouriteVideoCount(id)
	user = User{
		Id:             id,
		Name:           tableUser.Name,
		FollowCount:    0,
		FollowerCount:  0,
		IsFollow:       false,
		TotalFavorited: totalFavorited,
		FavoriteCount:  favoritedCount,
	}
	return user, nil
}

// GetUserByIdWithCurId 已登录(curID)情况下,根据user_id获得User对象
func (usi *UserServiceImpl) GetUserByIdWithCurId(id int64, curId int64) (User, error) {
	user := User{
		Id:             0,
		Name:           "",
		FollowCount:    0,
		FollowerCount:  0,
		IsFollow:       false,
		TotalFavorited: 0,
		FavoriteCount:  0,
	}
	tableUser, err := dao.GetTableUserById(id)
	if err != nil {
		log.Println("Err:", err.Error())
		log.Println("User Not Found")
		return user, err
	}
	log.Println("Query User Success")

	//followCount, err := usi.GetFollowingCnt(id)
	//if err != nil {
	//	log.Println("Err:", err.Error())
	//}
	//followerCount, err := usi.GetFollowerCnt(id)
	//if err != nil {
	//	log.Println("Err:", err.Error())
	//}
	//isfollow, err := usi.IsFollowing(curId, id)
	//if err != nil {
	//	log.Println("Err:", err.Error())
	//}
	u := GetLikeService() //解决循环依赖
	totalFavorited, _ := u.TotalFavourite(id)
	favoritedCount, _ := u.FavouriteVideoCount(id)
	user = User{
		Id:   id,
		Name: tableUser.Name,
		//FollowCount:   followCount,
		//FollowerCount: followerCount,
		//IsFollow:      isfollow,
		TotalFavorited: totalFavorited,
		FavoriteCount:  favoritedCount,
	}
	return user, nil
}

// GetUserByIdWithCurId 已登录(curID)情况下,根据user_id获得User对象
func (usi *UserServiceImpl) GetUserByIdWithCurId(id int64, curId int64) (User, error) {
	user := User{
		Id:            0,
		Name:          "",
		FollowCount:   0,
		FollowerCount: 0,
		IsFollow:      false,
		//TotalFavorited: 0,
		//FavoriteCount:  0,
	}
	tableUser, err := dao.GetTableUserById(id)
	if err != nil {
		log.Println("Err:", err.Error())
		log.Println("User Not Found")
		return user, err
	}
	log.Println("Query User Success")

	//followCount, err := usi.GetFollowingCnt(id)
	//if err != nil {
	//	log.Println("Err:", err.Error())
	//}
	//followerCount, err := usi.GetFollowerCnt(id)
	//if err != nil {
	//	log.Println("Err:", err.Error())
	//}
	//isfollow, err := usi.IsFollowing(curId, id)
	//if err != nil {
	//	log.Println("Err:", err.Error())
	//}
	//u := GetLikeService() //解决循环依赖
	//totalFavorited, _ := u.TotalFavourite(id)
	//favoritedCount, _ := u.FavouriteVideoCount(id)
	user = User{
		Id:   id,
		Name: tableUser.Name,
		//FollowCount:   followCount,
		//FollowerCount: followerCount,
		//IsFollow:      isfollow,
		//TotalFavorited: totalFavorited,
		//FavoriteCount:  favoritedCount,
	}
	return user, nil
}
