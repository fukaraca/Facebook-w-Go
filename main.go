package main

import (
	"context"
	"smallSteeps/lib"
)



func main()  {
	lib.Ctx =context.Background()
	lib.InitServer()
	lib.CreateRedisClient()
	lib.ConnectDB()


	//routes
	lib.R.NoRoute(lib.NoRoute404)
	lib.R.GET("/index", lib.GetIndex)
	lib.R.GET("/home", lib.GetHome)
	lib.R.POST("/checkAuthLog", lib.PostCheckAuth)
	lib.R.POST("/checkReg", lib.PostCheckReg)

	lib.R.Run(":8080")

}
