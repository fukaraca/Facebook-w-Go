package lib

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"time"
)

//Global definition will run before main() and init()
var R *gin.Engine
var cache *redis.Client
var Ctx = context.Background()
var conn *pgx.Conn
var err error

const (
	HOST          = "127.0.0.1"
	PORT          = "5432"
	DATABASE      = "dbForFaceClone"
	USER          = "postgres"
	PASSWORD      = "123456"
	defaultAvatar = "img/my_avatar.png" //avatar relative path
	TIMEOUT       = 5 * time.Second
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
	location     pgtype.Varchar     `conn:"location"`
	relationship pgtype.Varchar     `conn:"relationship"`
	school       pgtype.Varchar     `conn:"school"`
	workplace    pgtype.Varchar     `conn:"workplace"`
	bio          pgtype.Varchar     `conn:"bio"`
}

//YtEmbedJson is used for Youtube embed link requests from oembed api
type YtEmbedJson struct {
	Html string                 `json:"html"`
	X    map[string]interface{} `json:"-"`
}

//Must be exported for struct level assignments while templating
type EuserCred struct {
	UserID       int                `conn:"user_id"`
	Username     string             `conn:"username"`
	Password     string             `conn:"password"`
	Email        string             `conn:"email"`
	Name         pgtype.Text        `conn:"name"`
	Lastname     pgtype.Varchar     `conn:"lastname"`
	Gender       pgtype.Unknown     `conn:"gender"`
	MobileNumber pgtype.Varchar     `conn:"mobilenumber"`
	Country      pgtype.Varchar     `conn:"country"`
	Birthday     pgtype.Date        `conn:"birthday"`
	CreatedOn    pgtype.Timestamptz `conn:"createdon"`
	LastLogin    pgtype.Timestamp   `conn:"lastlogin"`
	AvatarPath   pgtype.Text        `conn:"avatarpath"`
	Location     pgtype.Varchar     `conn:"location"`
	Relationship pgtype.Varchar     `conn:"relationship"`
	School       pgtype.Varchar     `conn:"school"`
	Workplace    pgtype.Varchar     `conn:"workplace"`
	Bio          pgtype.Varchar     `conn:"bio"`
}

type FriendWhoToBeAdded struct {
	FriendID string `json:"friendid"`
	Since    string `json:"since"`
}
