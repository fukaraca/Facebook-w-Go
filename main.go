package main

import (
	"log"
	"smallSteeps/lib"
)

func init() {
	lib.InitServer()
	lib.CreateRedisClient()
	lib.ConnectDB()
}

func routes() {
	lib.R.NoRoute(lib.NoRoute404)
	lib.R.GET("/chat", lib.Auth(lib.WsChat))
	lib.R.GET("/logout", lib.Auth(lib.GetLogout))
	lib.R.GET("/login", lib.Auth(lib.GetHome))
	lib.R.GET("/home", lib.Auth(lib.GetHome))
	lib.R.GET("/profile", lib.Auth(lib.GetProfile))
	lib.R.GET("/settings", lib.Auth(lib.GetEdit))
	lib.R.GET("/user/:profileID", lib.Auth(lib.GetProfileByID))
	lib.R.GET("/unfriend/:profileID", lib.Auth(lib.GetUnfriend))
	lib.R.GET("/loadmorehome/:page", lib.Auth(lib.GetLoadMoreAtHome))
	lib.R.GET("/loadmoreprofile/:page", lib.Auth(lib.GetLoadMoreAtProfile))
	lib.R.GET("/loadmore/:profileID/:page", lib.Auth(lib.GetLoadMoreByUsername))
	lib.R.GET("/delpost/:postId", lib.Auth(lib.DelPostId))

	lib.R.POST("/searchuser", lib.Auth(lib.PostSearchUser))
	lib.R.POST("/postIt", lib.Auth(lib.PostIt))
	lib.R.POST("/addunfriend", lib.Auth(lib.PostAddUnfriend))
	lib.R.POST("/updateprofile", lib.Auth(lib.PostUpdateProfile))
	lib.R.POST("/updatepp", lib.Auth(lib.PostUpdateProfilePhoto))
	lib.R.POST("/updateBio", lib.Auth(lib.PostUpdateBio))
	lib.R.POST("/changepassword", lib.Auth(lib.PostChangePassword))
	lib.R.POST("/checkAuthLog", lib.PostCheckAuth)
	lib.R.POST("/checkReg", lib.PostCheckReg)
	lib.R.POST("/deleteaccount", lib.Auth(lib.PostDeleteAccount))
}

func main() {
	routes()
	log.Fatalln("Router encountered and error while main.Run:", lib.R.Run(lib.Server_Port))

}
