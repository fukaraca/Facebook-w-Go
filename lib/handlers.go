package lib

import (
	"context"
	"fmt"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

//NoRoute404 is handler function for 404 Page
func NoRoute404(c *gin.Context) {
	if c.Request.URL.Path == "/" {
		c.Redirect(http.StatusMovedPermanently, "/login")
		return
	}
	log.Println("Req ID:", requestid.Get(c))
	c.HTML(http.StatusNotFound, "404.html", nil)

}

//Auth is the authentication middleware
func Auth(fn gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("Req ID:", requestid.Get(c))
		if !CheckSession(c) {
			c.HTML(http.StatusUnauthorized, "login.html", nil)
			return
		}

		//c.Writer.Header().Set("X-Custom-Header", "value")

		fn(c)
	}
}

//GetIndex is handler function for homepage or start page todo gerek kalmadı
func GetIndex(c *gin.Context) {
	if !CheckSession(c) {
		c.HTML(http.StatusUnauthorized, "login.html", nil)
		return
	}

	GetHome(c)
}

//PostCheckAuth is the handler function for sign in check : Authentication
func PostCheckAuth(c *gin.Context) {
	if CheckSession(c) {
		return
	}
	logUserName := *Striper(c.PostForm("usernameL"))
	logPassword := *Striper(c.PostForm("passwordL"))
	//check login credential
	checkLoginQuery := fmt.Sprintf("SELECT password FROM user_creds WHERE username LIKE '%s' ;", logUserName)
	hashedToBeChecked, err := QueryLog(c, checkLoginQuery)
	if err != nil {
		loginMessage := err.Error() + " Something went wrong with DB/Query!"
		c.HTML(http.StatusBadGateway, "login.html", gin.H{
			"messageL": loginMessage,
		})
		return
	}
	if !CheckPasswordHash(logPassword, hashedToBeChecked) {
		loginMessage := "Password or Username is incorrect"
		c.HTML(http.StatusForbidden, "login.html", gin.H{
			"messageL": loginMessage,
		})
		return
	}
	lastLogTime := time.Now().Format(time.RFC3339)
	insertLastLog := fmt.Sprintf("UPDATE user_creds SET lastlogin = '%s' WHERE username = '%s';", lastLogTime, logUserName)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	_, err = conn.Exec(ctx, insertLastLog)
	if err != nil {
		log.Println("last login time update has error:", err)
	}

	CreateSession(logUserName, c)
	/*	c.HTML(http.StatusOK, "home.html", gin.H{
		"message": "Welcome " + logUserName,
	})*/
	log.Println("Req ID:", requestid.Get(c))
	c.Redirect(http.StatusMovedPermanently, "/home")
}

//PostCheckReg is handler function to Register new user
func PostCheckReg(c *gin.Context) {
	if CheckSession(c) {
		return
	}
	regUserName := *Striper(c.PostForm("usernameReg"))
	tempPass := *Striper(c.PostForm("passwordReg"))
	regPassword, err := HashPassword(tempPass)
	regEmail := *Striper(c.PostForm("emailReg"))
	regCreatedOn := time.Now().Format(time.RFC3339)
	//insert new users if not exist
	checkRegQuery := fmt.Sprintf("INSERT INTO user_creds (user_id,username,password,email,createdon) VALUES (nextval('user_id_seq'),'%s','%s','%s','%s');", regUserName, regPassword, regEmail, regCreatedOn)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	res, err := conn.Exec(ctx, checkRegQuery)

	messageR := "New account created successfully! You can login anytime."
	//if any error occurs
	if err != nil {
		messageR = QueryErr(err)
		log.Println("Registration error :", messageR)
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{
			"messageR": messageR,
		})
		return
	}
	//if it's OK
	log.Println("Req ID:", requestid.Get(c))
	log.Println("register Query:", res.String())

	c.HTML(http.StatusOK, "login.html", gin.H{
		"messageR": messageR,
	})

}

func GetHome(c *gin.Context) {

	c.HTML(200, "home.html", nil)

}

func GetProfile(c *gin.Context) {

	c.HTML(200, "profile.html", gin.H{
		".profileId": c.Request.URL.Path,
	})

}

//GetEdit function is handler for settings and edit user information page
func GetEdit(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	username, err := c.Cookie("uid")
	if err == http.ErrNoCookie {
		log.Println("No cookie error: ", err)
		return
	}
	//short status message like "update executed successfully"
	statusMessage, _ := c.Cookie("short_status_message")
	c.SetCookie("short_status_message", "", -1, "/", "localhost", false, true)

	querriedProfile, err := BringMeProfile(ctx, username)
	if err != nil {
		log.Println(err)
	}
	relPath, err := BringMeAvatar(querriedProfile.avatarPath.String, username)
	if err != nil {
		log.Println(err)
	}
	c.HTML(http.StatusOK, "settings.html", gin.H{
		"firstname":     querriedProfile.name.String,
		"lastname":      querriedProfile.lastname.String,
		"gender":        querriedProfile.gender.String,
		"birthday":      querriedProfile.birthday.Time.Format("2006-02-01"),
		"mobilenumber":  querriedProfile.mobileNumber.String,
		"country":       querriedProfile.country.String,
		"statusMessage": statusMessage,
		"avatarPath":    relPath,

		/*		"userStatus":,
				"userName":username,
				"avatarPath":*/
	})
}

//PostUpdateProfile function handles settings page in order to update user informations
func PostUpdateProfile(c *gin.Context) {

	upFirstname := *Striper(c.PostForm("firstname"))
	upLastname := *Striper(c.PostForm("lastname"))
	upGender := *Striper(c.PostForm("gender"))
	upBirthday := *Striper(c.PostForm("birthday"))
	upMobile := *Striper(c.PostForm("mobilenumber"))
	upCountry := *Striper(c.PostForm("country"))

	username, err := c.Cookie("uid")
	if err == http.ErrNoCookie {
		log.Println("No cookie error: ", err)
		return
	}
	updateQueryString := fmt.Sprintf("UPDATE user_creds SET name = '%s',lastname = '%s',gender = '%s',birthday = '%s',mobilenumber = '%s',country = '%s' WHERE username = '%s';", upFirstname, upLastname, upGender, upBirthday, upMobile, upCountry, username)
	fmt.Println(updateQueryString) //todo silinecek
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	_, err = conn.Exec(ctx, updateQueryString)
	if err != nil {
		log.Println("user profile update has failed:", err)
		return
	}
	shortStatusMessage := "Updated successfully!"
	c.SetCookie("short_status_message", shortStatusMessage, 60, "/", "localhost", false, true)
	c.Redirect(303, "/settings")

}

//PostUpdateProfilePhoto is handling function for changing profile picture
func PostUpdateProfilePhoto(c *gin.Context) {

	picToBeUploaded, header, err := c.Request.FormFile("change_pp")
	if err != nil {
		log.Println("Avatar photo couldn't be uploaded", err)
		c.SetCookie("short_status_message", "Avatar photo couldn't be uploaded:"+err.Error(), 60, "/", "localhost", false, true)
		return
	}
	defer picToBeUploaded.Close()

	username, err := c.Cookie("uid")
	if err == http.ErrNoCookie {
		log.Println("No uid cookie error: ", err)
		c.SetCookie("short_status_message", "avatar photo couldn't be uploaded:"+err.Error(), 60, "/", "localhost", false, true)
		return
	}

	filename := username + RandomString(4) + filepath.Ext(header.Filename)
	filePathString := fmt.Sprintf("./private/assets/avatars/%s/", username)
	err = os.MkdirAll(filePathString, 0666)
	if err != nil {
		log.Println("filepath for avatar couldn't be created")
		c.SetCookie("short_status_message", "avatar photo couldn't be uploaded:"+err.Error(), 60, "/", "localhost", false, true)
		return
	}

	//resize and save file
	err = ResizeAndSave(picToBeUploaded, filePathString, filename)
	if err != nil {
		c.SetCookie("short_status_message", "avatar photo couldn't be uploaded:"+err.Error(), 60, "/", "localhost", false, true)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	insertyString := fmt.Sprintf("UPDATE user_creds SET avatarpath = '%s' WHERE username= '%s';", filePathString+filename, username)
	_, err = conn.Exec(ctx, insertyString)
	if err != nil {
		log.Println("avatarpath update to DB was failed:", err)
		c.SetCookie("short_status_message", "avatar path couldn't be saved:"+err.Error(), 60, "/", "localhost", false, true)
		return
	}
	c.SetCookie("short_status_message", "Profile picture has been updated successfully!", 30, "/", "localhost", false, true)
	c.Redirect(http.StatusFound, "/settings")

}

func PostChangePassword(c *gin.Context) {
	newPassword := *Striper(c.PostForm("newpassword"))
	oldPassword := *Striper(c.PostForm("oldpassword"))
	username, err := c.Cookie("uid")
	if err != nil {
		if err == http.ErrNoCookie {
			c.SetCookie("short_status_message", "Changing password failed:"+err.Error(), 30, "/", "localhost", false, true)
			c.Redirect(http.StatusFound, "/settings")
			return
		}
	}

	checkLoginQuery := fmt.Sprintf("SELECT password FROM user_creds WHERE username LIKE '%s' ;", username)
	hashedToBeChecked, err := QueryLog(c, checkLoginQuery)
	if err != nil {
		c.SetCookie("short_status_message", "Something went wrong with DB/Query:!"+err.Error(), 30, "/", "localhost", false, true)
		c.Redirect(http.StatusFound, "/settings")
		return
	}

	if !CheckPasswordHash(oldPassword, hashedToBeChecked) {
		c.SetCookie("short_status_message", "Password is incorrect!", 30, "/", "localhost", false, true)
		c.Redirect(http.StatusFound, "/settings")
		return
	}

	newHashed, err := HashPassword(newPassword)
	if err != nil {
		c.SetCookie("short_status_message", "Something went wrong with encryption but don't worry, it's not you, it's me!"+err.Error(), 30, "/", "localhost", false, true)
		c.Redirect(http.StatusFound, "/settings")
		return
	}

	updatePassword := fmt.Sprintf("UPDATE user_creds SET password = '%s' WHERE username = '%s';", newHashed, username)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	_, err = conn.Exec(ctx, updatePassword)
	if err != nil {
		c.SetCookie("short_status_message", "Password change failed!"+err.Error(), 30, "/", "localhost", false, true)
		c.Redirect(http.StatusFound, "/settings")
		return
	}
	ok, err := DeleteSession(c)
	if !ok || err != nil {
		log.Println("session couldn't be deleted")
	}
	c.HTML(http.StatusFound, "login.html", gin.H{
		"messageL": "Password changed successfully, please login with the new password",
	})

}

//todo chrome da bazen cikis yapmıyor

//GetLogout is handler function for logout
func GetLogout(c *gin.Context) {
	ok, err := DeleteSession(c)
	if err != nil || !ok {
		log.Println("session couldn't be deleted:", err)
		return
	}
	c.Redirect(http.StatusFound, "/login")
}
