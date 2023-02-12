package jwt

import (
	"fmt"
	"github.com/RaymondCode/simple-demo/config"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type Response struct {
	StatusCode int32  `json:"status_code"`
	StatusMsg  string `json:"status_msg,omitempty"`
}

// 若用户携带的token正确,解析token,将userId放入上下文context中并放行;否则,返回错误信息
func Auth() gin.HandlerFunc {
	return func(context *gin.Context) {
		//auth := context.Request.Header.Get("Authorization")
		auth := context.Query("token")
		if len(auth) == 0 {
			context.Abort()
			context.JSON(http.StatusUnauthorized, Response{
				StatusCode: -1,
				StatusMsg:  "Unauthorized",
			})
		}

		token, err := parseToken(auth)
		if err != nil {
			context.Abort()
			context.JSON(http.StatusUnauthorized, Response{
				StatusCode: -1,
				StatusMsg:  "Token Error",
			})
		} else {
			log.Println("token 正确")
		}
		if token != nil { // 判断，防止退出登录时销毁token，发生空指针错误
			context.Set("userId", token.Id)
		}
		context.Next()
	}
}

// 未登录情况下,若携带token,则解析出用户id并放入context;若未携带,则放入用户id默认值0
func AuthWithoutLogin() gin.HandlerFunc {
	return func(context *gin.Context) {
		auth := context.Query("token")
		var userId string
		if len(auth) == 0 {
			userId = "0"
		} else {

			token, err := parseToken(auth)
			if err != nil {
				context.Abort()
				context.JSON(http.StatusUnauthorized, Response{
					StatusCode: -1,
					StatusMsg:  "Token 错误",
				})
			} else {
				userId = token.Id
				println("token 正确")
			}
		}
		context.Set("userId", userId)
		context.Next()
	}
}

// 解析token,判断token是否正确
func parseToken(token string) (*jwt.StandardClaims, error) {
	jwtToken, err := jwt.ParseWithClaims(token, &jwt.StandardClaims{}, func(token *jwt.Token) (i interface{}, e error) {
		return []byte(config.Secret), nil
	})
	if err == nil && jwtToken != nil {
		if claim, ok := jwtToken.Claims.(*jwt.StandardClaims); ok && jwtToken.Valid {
			return claim, nil
		}
	}
	return nil, err
}

// AuthBody 鉴权中间件
// 若用户携带的token正确,解析token,将userId放入上下文context中并放行;否则,返回错误信息
func AuthBody() gin.HandlerFunc {
	return func(context *gin.Context) {
		auth := context.Request.PostFormValue("token")
		fmt.Printf("%v \n", auth)

		if len(auth) == 0 {
			context.Abort()
			context.JSON(http.StatusUnauthorized, Response{
				StatusCode: -1,
				StatusMsg:  "Unauthorized",
			})
		}

		token, err := parseToken(auth)
		if err != nil {
			context.Abort()
			context.JSON(http.StatusUnauthorized, Response{
				StatusCode: -1,
				StatusMsg:  "Token 错误",
			})
		} else {
			println("token 正确")
		}
		context.Set("userId", token.Id)
		context.Next()
	}
}
