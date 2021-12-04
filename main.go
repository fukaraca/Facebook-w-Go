package main

import (
	"context"
	"log"
	"smallSteeps/lib"
)

func main() {
	lib.Ctx = context.Background()
	lib.InitServer()
	lib.CreateRedisClient()
	lib.ConnectDB()

	//routes
	lib.R.NoRoute(lib.NoRoute404)
	lib.R.GET("/login", lib.GetIndex)
	lib.R.GET("/home", lib.GetHome)
	lib.R.GET("/profile", lib.GetProfile)
	lib.R.POST("/checkAuthLog", lib.PostCheckAuth)
	lib.R.POST("/checkReg", lib.PostCheckReg)

	err := lib.R.Run(":8080")
	if err != nil {
		log.Println("Router encountered and error while main.Run:", err)
	}

}
