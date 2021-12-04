package lib

import (
	con "context"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/gofrs/uuid"
	"log"
	"net/http"
)

//CheckCookie function checks validation of cookie. Return TRUE if it's valid
func CheckCookie(toBeChecked string) bool {
	cookieVal, err := cache.Do(con.Background(), "GET", toBeChecked).Result()
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
	}

	return true
}

//SetCookie creates cookie for users who logged in successfully and returns the Cookie Values(UUID)
func SetCookie(username string, c *gin.Context) string {
	sessionToken, err := uuid.NewV4()
	if err != nil {
		log.Println("new UUID could'nt assigned error:", err)
	}
	cache.Do(con.Background(), "SETEX", sessionToken.String(), "3600", username)

	c.SetCookie("session_token", sessionToken.String(), 3600, "/", "localhost", true, true)
	return sessionToken.String()
}

//CheckSession function checks validation of session. If a request has no cookie or cookie is not valid then returns FALSE
func CheckSession(c *gin.Context) bool {
	toBeChecked, err := c.Cookie("session_token")
	if err == http.ErrNoCookie {
		log.Println("No cookie error: ", err)
		return false
	}

	if isCookieValid := CheckCookie(toBeChecked); !isCookieValid {
		log.Println("Cookie is not valid", toBeChecked)
		return false
	}

	return true
}

func RefreshSession(c *gin.Context) {

}
