package lib

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/gofrs/uuid"
	"golang.org/x/net/context"
	"log"
	"net/http"
	"os"
	"time"
)

//CheckCookie function checks validation of cookie. Return TRUE if it's valid
func CheckCookie(c *gin.Context, toBeChecked, userId string) bool {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	cookieVal, err := cache.Do(ctx, "GET", toBeChecked).Result()

	switch {
	case err == redis.Nil:
		log.Println("Cookie does not exist!")
		return false
	case err != nil:
		log.Println("Get Failed:", err)
		return false
	case cookieVal == "":
		log.Println("Cookie value is empty!")
		return false
	case userId != cookieVal.(string):
		return false
	}
	return true
}

//todo secure flag!!
//CreateSession creates cookie for users who logged in successfully and returns the Cookie Values(UUID)
func CreateSession(username string, c *gin.Context) {
	sessionToken, err := uuid.NewV4()
	if err != nil {
		log.Println("new UUID could'nt assigned error:", err)
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	cache.Do(ctx, "SETEX", sessionToken.String(), "3600", username)

	c.SetCookie("session_token", sessionToken.String(), 3600, "/", "localhost", false, true)
	c.SetCookie("uid", username, 3600, "/", "localhost", false, true)

}

//CheckSession function checks validation of session. If a request has no cookie or cookie is not valid then returns FALSE
func CheckSession(c *gin.Context) bool {
	toBeChecked, err := c.Cookie("session_token")
	if err == http.ErrNoCookie {
		log.Println("No cookie error: ", err)
		return false
	}

	toBeCheckedId, err := c.Cookie("uid")
	if err == http.ErrNoCookie {
		log.Println("No cookie error: ", err)
		return false
	}

	if isCookieValid := CheckCookie(c, toBeChecked, toBeCheckedId); !isCookieValid {
		log.Println("Cookie is not valid", toBeChecked)
		return false
	}

	return true
}

func RefreshSession(c *gin.Context) {

}

//DeleteSession deletes the session as named
func DeleteSession(c *gin.Context) (bool, error) {
	toBeChecked, err := c.Cookie("session_token")
	if err != nil {
		return false, fmt.Errorf("no cookie to delete:%s", err.Error())
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	res, err := cache.Do(ctx, "DEL", toBeChecked).Result()
	if err != nil {
		log.Println("session deletion err:", err)
		return false, fmt.Errorf("session deletion err:%s", err.Error())
	}

	log.Printf("%v item removed in order to delete session.\n", res)

	c.SetCookie("session_token", "", -1, "/", "localhost", false, true)
	c.SetCookie("uid", "", -1, "/", "localhost", false, true)
	c.SetCookie("short_status_message", "", -1, "/", "localhost", false, true)
	err = os.RemoveAll("./web/asset/avatars")
	if err != nil {
		log.Println("deleting the public folders due to logout failed", err)
	}
	return true, nil
}
