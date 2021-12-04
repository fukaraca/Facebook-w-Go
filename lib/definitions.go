package lib

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4"
	"time"
)

var R *gin.Engine
var cache *redis.Client
var Ctx context.Context
var conn *pgx.Conn
var err error

type User_Cred struct {
	Username string
	Password string
	Email string
	Name string
	Lastname string
	CreatedOn time.Time
	LastLogin time.Time
	MobileNumber string
	Country string
	//Rememberme bool
}

const (
	HOST="127.0.0.1"
	PORT="5432"
	DATABASE="dbForFaceClone"
	USER="postgres"
	PASSWORD="123456"
)



