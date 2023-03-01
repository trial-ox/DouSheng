package main

import (
	"github.com/RaymondCode/simple-demo/controller"
	"github.com/RaymondCode/simple-demo/middleware/jwt"
	"github.com/gin-gonic/gin"
)

func initRouter(r *gin.Engine) {
	// public directory is used to serve static resources
	r.Static("/static", "./public")

	apiRouter := r.Group("/douyin")

	// basic apis
	apiRouter.GET("/feed/", jwt.AuthWithoutLogin(), controller.Feed)
	apiRouter.GET("/user/", jwt.Auth(), controller.UserInfo)
	apiRouter.POST("/user/register/", controller.Register)
	apiRouter.POST("/user/login/", controller.Login)
	apiRouter.POST("/publish/action/", jwt.AuthBody(), controller.Publish)
	apiRouter.GET("/publish/list/", jwt.Auth(), controller.PublishList)

	// extra apis - I
	apiRouter.POST("/favorite/action/", jwt.Auth(), controller.FavoriteAction)
	apiRouter.GET("/favorite/list/", jwt.Auth(), controller.FavoriteList)
	apiRouter.POST("/comment/action/", jwt.Auth(), controller.CommentAction)
	apiRouter.GET("/comment/list/", jwt.AuthWithoutLogin(), controller.CommentList)

	// extra apis - II
	apiRouter.POST("/relation/action/", jwt.Auth(), controller.RelationActionFuction)
	apiRouter.GET("/relation/follow/list/", jwt.Auth(), controller.GetFollowing)
	apiRouter.GET("/relation/follower/list/", jwt.Auth(), controller.GetFollowers)
	apiRouter.GET("/relation/friend/list/", jwt.Auth(), controller.FriendList)
	apiRouter.GET("/message/chat/", jwt.Auth(), controller.MessageChat)
	apiRouter.POST("/message/action/", jwt.Auth(), controller.MessageAction)
}
