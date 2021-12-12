package lib

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

//Global definition will run before main() and init()
var R *gin.Engine
var cache *redis.Client
var Ctx = context.Background() //context.Context todo
var conn *pgx.Conn
var err error

const (
	HOST          = "127.0.0.1"
	PORT          = "5432"
	DATABASE      = "dbForFaceClone"
	USER          = "postgres"
	PASSWORD      = "123456"
	defaultAvatar = "img/my_avatar.png" //avatar relative path
)

type userCred struct {
	userID       int                `conn:"user_id"`
	username     string             `conn:"username"`
	password     string             `conn:"password"`
	email        string             `conn:"email"`
	name         pgtype.Text        `conn:"name"`
	lastname     pgtype.Varchar     `conn:"lastname"`
	gender       pgtype.Unknown     `conn:"gender"`
	mobileNumber pgtype.Varchar     `conn:"mobilenumber"`
	country      pgtype.Varchar     `conn:"country"`
	birthday     pgtype.Date        `conn:"birthday"`
	createdOn    pgtype.Timestamptz `conn:"createdon"`
	lastLogin    pgtype.Timestamp   `conn:"lastlogin"`
	avatarPath   pgtype.Text        `conn:"avatarpath"`

	//Rememberme bool
}
