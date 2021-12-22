package main

import (
	"log"
	"smallSteeps/lib"
)

func main() {
	lib.InitServer()
	lib.CreateRedisClient()
	lib.ConnectDB()

	//routes
	lib.R.NoRoute(lib.NoRoute404)
	lib.R.GET("/logout", lib.Auth(lib.GetLogout))
	lib.R.GET("/login", lib.Auth(lib.GetHome))
	lib.R.GET("/home", lib.Auth(lib.GetHome))
	lib.R.GET("/profile", lib.Auth(lib.GetProfile))
	lib.R.GET("/settings", lib.Auth(lib.GetEdit))
	lib.R.GET("/user/:profileID", lib.Auth(lib.GetProfileByID))
	lib.R.GET("/unfriend/:profileID", lib.Auth(lib.GetUnfriend))

	lib.R.POST("/addunfriend", lib.Auth(lib.PostAddUnfriend))
	lib.R.POST("/updateprofile", lib.Auth(lib.PostUpdateProfile))
	lib.R.POST("/updatepp", lib.Auth(lib.PostUpdateProfilePhoto))
	lib.R.POST("/changepassword", lib.Auth(lib.PostChangePassword))
	lib.R.POST("/checkAuthLog", lib.PostCheckAuth)
	lib.R.POST("/checkReg", lib.PostCheckReg)
	lib.R.POST("/deleteaccount", lib.Auth(lib.PostDeleteAccount))

	log.Fatalln("Router encountered and error while main.Run:", lib.R.Run(":8080"))

}
