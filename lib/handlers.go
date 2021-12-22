package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"math/rand"
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

//GetHome function is handler function for GET' ting home page
func GetHome(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	username, err := c.Cookie("uid")
	if err == http.ErrNoCookie {
		log.Println("No cookie error: ", err)
		return
	}
	queryAllUser := fmt.Sprintf("select username from user_creds;")
	res, err := conn.Query(ctx, queryAllUser)

	defer res.Close()
	users := []string{}
	for res.Next() {
		tempUsername := ""
		err = res.Scan(&tempUsername)
		if err != nil {
			log.Println("scan all friend query failed:", err)
			return
		}
		users = append(users, tempUsername)
	}
	if err = res.Err(); err != nil {
		log.Println("query error:", err)
	}
	//suggest me 3 random user as friend
	suggestibles := []string{} //non-friends
	for _, user := range users {
		if user != username && !QueryFriendship(ctx, username, user) {
			suggestibles = append(suggestibles, user)
		}
	}
	rand.Seed(time.Now().UnixNano())
	toBeSuggested := make(map[int]string)

	for i := 0; i < 3; i++ {
		randIndex := rand.Intn(len(suggestibles))
		toBeSuggested[randIndex] = suggestibles[randIndex]
		if len(toBeSuggested) < 3 && i == 2 {
			i--
		}
	}
	//show me one of my friends
	friendsList, err := BringMeFriends(ctx, username)

	c.HTML(http.StatusOK, "home.html", gin.H{
		"suggestibles": toBeSuggested,
		"randomFriend": friendsList[rand.Intn(len(friendsList))].FriendID,
	})

}

//GetProfile is the function for clients profile page
func GetProfile(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	username, err := c.Cookie("uid")
	if err == http.ErrNoCookie {
		log.Println("No cookie error: ", err)
		return
	}
	querriedFriendList, err := BringMeFriends(ctx, username)
	if err != nil {
		log.Println("Bringmefriends query failed:", err)
	}

	querriedProfile, err := BringMeThatProfile(ctx, username)
	if err != nil {
		log.Println(err)
	}
	relPath, err := BringMeAvatar(querriedProfile.AvatarPath.String, username)
	if err != nil {
		log.Println(err)
	}
	relPath, _ = filepath.Rel("./user", relPath)
	relPath = filepath.ToSlash(relPath)
	//Exported = querriedProfile //todo silince
	c.HTML(http.StatusOK, "profile.html", gin.H{
		"profileID":     username,
		"avatarPath":    relPath,
		"profilestruct": querriedProfile,
		"friends":       querriedFriendList,
	})

}

//GetProfileByID is the handler function for certain profile page
func GetProfileByID(c *gin.Context) {
	username, err := c.Cookie("uid")
	if err == http.ErrNoCookie {
		log.Println("No cookie error: ", err)
		return
	}
	profileID := c.Param("profileID")
	if username == profileID {
		c.Redirect(http.StatusFound, "/profile")
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), TIMEOUT)
	defer cancel()
	queryStr := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM user_creds WHERE username='%s');", profileID)
	ok := QueryUsername(ctx, queryStr)
	if !ok {
		log.Println("ProfileID couldn't be found in DB")
		c.Redirect(http.StatusFound, "/userNotFound")
	}

	querriedFriendList, err := BringMeFriends(ctx, profileID)
	if err != nil {
		log.Println("Bringmefriends query failed:", err)
	}
	querriedFriendship := QueryFriendship(ctx, username, profileID)
	addButValue := "Add Friend"
	if querriedFriendship {
		addButValue = "Unfriend"
	}
	querriedProfile, err := BringMeThatProfile(ctx, profileID)
	if err != nil {
		log.Println(err)
	}
	relPath, err := BringMeAvatar(querriedProfile.AvatarPath.String, profileID)
	if err != nil {
		log.Println(err)
	}
	relPath, _ = filepath.Rel("./user", relPath)
	relPath = filepath.ToSlash(relPath)
	c.HTML(http.StatusOK, "otheruserprofile.html", gin.H{
		"profileID":      profileID,
		"avatarPath":     relPath,
		"profilestruct":  querriedProfile,
		"addButtonValue": addButValue,
		"friends":        querriedFriendList,
	})

}

//GetUnfriend unfriend handler function for profile page
func GetUnfriend(c *gin.Context) {
	username, err := c.Cookie("uid")
	if err == http.ErrNoCookie {
		log.Println("No cookie error: ", err)
		return
	}
	profileID := c.Param("profileID")
	if username == profileID {
		c.Redirect(http.StatusFound, "/profile")
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), TIMEOUT)
	defer cancel()

	err = UnfriendQuery(ctx, username, profileID)
	if err != nil {
		log.Println("unfriend error1: ", err)
		return
	}
	err = UnfriendQuery(ctx, profileID, username)
	if err != nil {
		log.Println("unfriend error2: ", err)
		return
	}
	c.Redirect(http.StatusFound, "/profile")
}

//PostAddUnfriend is handler func for add or delete some from friends list
func PostAddUnfriend(c *gin.Context) {

	username, err := c.Cookie("uid")
	if err == http.ErrNoCookie {
		log.Println("No cookie error: ", err)
		return
	}
	friend := FriendWhoToBeAdded{}

	err = c.Bind(&friend)
	if err != nil && err != io.EOF {
		log.Println("binding json failed:", err)
		return
	}
	if username == friend.FriendID {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), TIMEOUT)
	defer cancel()
	ok := QueryFriendship(ctx, username, friend.FriendID)
	if !ok { //add as friend
		friend.Since = time.Now().Format("2006-01-02")
		jsonFriend, err := json.Marshal(friend)
		if err != nil {
			log.Println("json marshalling failed1:", err)
			return
		}
		err = InsertNewFriendRowForUser(ctx, username) //todo silinecek asagısı
		err = InsertNewFriendRowForUser(ctx, friend.FriendID)
		if err != nil {
			log.Println("insert username if not exist query was failed:", err)
			return
		}
		//update friends list for both username and thatProfile
		addFriendString := fmt.Sprintf("UPDATE friends SET friendsince = COALESCE(friendsince, '[]' ::JSONB ) || '[%s]'::JSONB WHERE username='%s';", jsonFriend, username) //jsonb array için {}kullanıyor. JSONB İÇİN []  COALESCE
		_, err = conn.Exec(ctx, addFriendString)
		if err != nil {
			log.Println("update friends list was failed1:", err)
			return
		}
		tempName := friend.FriendID
		friend.FriendID = username
		jsonFriend, err = json.Marshal(friend)
		if err != nil {
			log.Println("json marshalling failed2:", err)
			return
		}
		addFriendString = fmt.Sprintf("UPDATE friends SET friendsince = COALESCE(friendsince, '[]' ::JSONB ) || '[%s]'::JSONB WHERE username='%s';", jsonFriend, tempName) //jsonb array için {}kullanıyor. JSONB İÇİN []  COALESCE
		_, err = conn.Exec(ctx, addFriendString)
		if err != nil {
			log.Println("update friends list was failed2:", err)
			return
		}

		c.String(http.StatusOK, "true")
		return
	}
	//unfriend for both username and thatProfile
	err = UnfriendQuery(ctx, username, friend.FriendID)
	if err != nil {
		log.Println("unfriend failed1", err)
	}
	err = UnfriendQuery(ctx, friend.FriendID, username)
	if err != nil {
		log.Println("unfriend failed1", err)
	}

	c.String(http.StatusOK, "false")
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
		"birthday":      querriedProfile.birthday.Time.Format("2006-01-02"),
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
	upBirthday := *Striper(c.PostForm("birthday")) //year month date is as sequence
	upMobile := *Striper(c.PostForm("mobilenumber"))
	upCountry := *Striper(c.PostForm("country"))

	username, err := c.Cookie("uid")
	if err == http.ErrNoCookie {
		log.Println("No cookie error: ", err)
		return
	}
	updateQueryString := fmt.Sprintf("UPDATE user_creds SET name = '%s',lastname = '%s',gender = '%s',birthday = '%s',mobilenumber = '%s',country = '%s' WHERE username = '%s';", upFirstname, upLastname, upGender, upBirthday, upMobile, upCountry, username)

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

//PostChangePassword is the function for changing login password.
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

//PostDeleteAccount function simply deletes the request related account
func PostDeleteAccount(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	confPass := *Striper(c.PostForm("deletepw"))

	username, err := c.Cookie("uid")
	if err != nil {
		if err == http.ErrNoCookie {
			c.SetCookie("short_status_message", "Delete failed:"+err.Error(), 30, "/", "localhost", false, true)
			c.Redirect(http.StatusFound, "/settings")
			return
		}
	}

	checkPassQuery := fmt.Sprintf("SELECT password FROM user_creds WHERE username LIKE '%s' ;", username)
	hashedToBeChecked, err := QueryLog(c, checkPassQuery)
	if err != nil {
		c.SetCookie("short_status_message", "Something went wrong with DB/Query(pd):!"+err.Error(), 30, "/", "localhost", false, true)
		c.Redirect(http.StatusFound, "/settings")
		return
	}
	if !CheckPasswordHash(confPass, hashedToBeChecked) {
		c.SetCookie("short_status_message", "Password is incorrect!", 30, "/", "localhost", false, true)
		c.Redirect(http.StatusFound, "/settings")
		return
	}

	deleteAccQuery := fmt.Sprintf("DELETE FROM user_creds WHERE username LIKE '%s' ;", username)
	_, err = conn.Exec(ctx, deleteAccQuery)
	if err != nil {
		c.SetCookie("short_status_message", "Delete failed!"+err.Error(), 30, "/", "localhost", false, true)
		c.Redirect(http.StatusFound, "/settings")
		return
	}

	c.Redirect(http.StatusFound, "/logout")

}

//GetLogout is handler function for logout
func GetLogout(c *gin.Context) {
	ok, err := DeleteSession(c)
	if err != nil || !ok {
		log.Println("session couldn't be deleted:", err)
		return
	}
	c.Redirect(http.StatusFound, "/login")
}
