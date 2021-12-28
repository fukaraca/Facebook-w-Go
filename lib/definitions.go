package lib

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"html/template"
	"time"
)

//Global definition will run before main() and init()
var R *gin.Engine
var cache *redis.Client
var Ctx = context.Background()
var conn *pgx.Conn
var err error

var GetEnv = setEnv()
var GetVars = setVars()

var highestRes = float32(GetEnv.GetInt("HIGHEST_RESOLUTION"))
var defaultAvatar = GetEnv.GetString("DEFAULT_AVATAR_PATH")
var TIMEOUT = GetEnv.GetDuration("TIMEOUT_CTX")
var db_Host = GetEnv.GetString("DB_HOST")
var db_Port = GetEnv.GetString("DB_PORT")
var db_Name = GetEnv.GetString("DB_NAME")
var db_User = GetEnv.GetString("DB_USER")
var db_Password = GetEnv.GetString("DB_PASSWORD")
var redis_Host = GetEnv.GetString("REDIS_HOST")
var redis_Port = GetEnv.GetString("REDIS_PORT")
var redis_Password = GetEnv.GetString("REDIS_PASSWORD")
var redis_DB = GetEnv.GetInt("REDIS_DB")

var Server_Host = GetEnv.GetString("SERVER_HOST")
var Server_Port = GetEnv.GetString("SERVER_PORT")

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

type Relationship struct {
	Username   string      `conn:"username" json:"username"`
	Friendname string      `conn:"friendname" json:"friendID"`
	Since      pgtype.Date `conn:"since" json:"since"`
}

type PostThatBeSaved struct {
	Postername        string    `json:"postername" conn:"postername"`
	PostId            string    `json:"post_id" conn:"post_id"`
	PostTime          time.Time `json:"post_time" conn:"post_time"`
	PostMessage       string    `json:"post_message" conn:"post_message"`
	PostImageFilepath string    `json:"post_image_filepath" conn:"post_image_filepath"`
	PostYtEmbedLink   string    `json:"post_yt_embed_link" conn:"post_yt_embed_link"`
}

type PostThatBeTemplated struct {
	Postername        string        `json:"postername" conn:"postername"`
	PostId            string        `json:"post_id" conn:"post_id"`
	PostTime          time.Time     `json:"post_time" conn:"post_time"`
	PostMessage       string        `json:"post_message" conn:"post_message"`
	PostImageFilepath template.HTML `json:"post_image_filepath" conn:"post_image_filepath"`
	PostYtEmbedLink   template.HTML `json:"post_yt_embed_link" conn:"post_yt_embed_link"`
}
type ToBeLoadedMore template.HTML
