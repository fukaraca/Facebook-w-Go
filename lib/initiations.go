package lib

import (
	con "context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4"
	"log"
)

//InitServer function initiates general stuff
func InitServer() {
	R =gin.Default()
	R.LoadHTMLGlob("template/*.html")
	R.Static("/css","template/css") //static files icin kullanilir
}

//CreateRedisClient function creates a Redis Client
func CreateRedisClient(){
	client:=redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		Password: "",
		DB: 0,
	})
	pong,err:=client.Ping(con.Background()).Result()
	if err != nil {
		log.Println("redis ping error:",err)
	}
	log.Println(pong," redis activated")
	cache =client
}
//ConnectDB function opens a connection to PSQL DB
func ConnectDB()  {
	var databaseURL =fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", HOST, PORT, USER, PASSWORD, DATABASE)
	conn, err =pgx.Connect(Ctx,databaseURL)
	if err != nil {
		log.Println("DB connection error:",err)
	}
	//check whether connection is ok or not
	err = conn.Ping(Ctx)
	if err != nil {
		log.Println("Ping to DB error:",err)
	}


}

