package lib

import (
	"context"
	"fmt"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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
			c.Abort()
			return
		}

		fn(c)
	}
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

//GetHome function is handler function for GET' ting home page
func GetHome(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), TIMEOUT)
	defer cancel()
	username, err := c.Cookie("uid")
	if err == http.ErrNoCookie {
		log.Println("No cookie error: ", err)
		return
	}
	statusMessage, _ := c.Cookie("short_status_message")
	c.SetCookie("short_status_message", "", -1, "/", Server_Host, false, true)

	//suggest me 3 random user as friend and also bring one of my friends
	toBeSuggested, randomFriend, err := FindMeSuggestibleFriendsAndAlsoOneOfMine(ctx, username)
	if err != nil {
		log.Println("Suggestible friend query failed", err)
		randomFriend = "above for future friend"
	}

	latestPosts, err := BringMeSomePosts(ctx, username)

	if err != nil {
		log.Println("loading of latest posts was failed", err)
		statusMessage = fmt.Sprintf("%s\nloading of latest posts was failed:%s", statusMessage, err.Error())
	}

	c.HTML(http.StatusOK, "home.html", gin.H{
		"profileId":     username,
		"suggestibles":  toBeSuggested,
		"randomFriend":  randomFriend,
		"statusMessage": statusMessage,
		"posts":         latestPosts,
	})
}

//GetProfile is the function for clients profile page
func GetProfile(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), TIMEOUT)
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
	relPath := defaultAvatar
	if querriedProfile.AvatarPath.String != "" {
		relPath, err = BringMeAvatar(querriedProfile.AvatarPath.String, username)
		if err != nil {
			log.Println(err)
		}
		relPath, _ = filepath.Rel("./user", relPath)
		relPath = filepath.ToSlash(relPath)
	}

	myLatestPosts, err := BringMeSomeMyPosts(ctx, username)
	if err != nil {
		log.Println("loading of my latest posts was failed", err)
	}

	gallery := BringHisGallery(username, 10)
	c.HTML(http.StatusOK, "profile.html", gin.H{
		"profileID":     username,
		"avatarPath":    relPath,
		"profilestruct": querriedProfile,
		"friends":       querriedFriendList,
		"posts":         myLatestPosts,
		"gallery":       gallery,
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
	ok := QueryUsername(ctx, profileID)
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
	//bringmeavatar for requested profileId
	relPath, err := BringMeAvatar(querriedProfile.AvatarPath.String, profileID)
	if err != nil {
		log.Println(err)
	}
	relPath, _ = filepath.Rel("./user", relPath)
	relPath = filepath.ToSlash(relPath)

	relPathUsername := ""
	//bring me usernames avatar for chat thumbnail
	row := conn.QueryRow(ctx, "SELECT avatarpath FROM user_creds WHERE username=$1;", username)
	_ = row.Scan(&relPathUsername)
	if relPathUsername == "" {
		relPathUsername = defaultAvatar
	} else {
		relPathUsername, err = BringMeAvatar(relPathUsername, username)
		if err != nil {
			log.Println("bring me avatar failed in getprofilebyid:", err)
		}
		relPathUsername, _ = filepath.Rel("./user", relPathUsername)
		relPathUsername = filepath.ToSlash(relPathUsername)
	}

	hisLatestPosts, err := BringMeSomeMyPosts(ctx, profileID)

	if err != nil {
		log.Println("loading of my latest posts was failed", err)
	}
	gallery := []string{}
	tempGallery := BringHisGallery(profileID, 10)
	for _, oneImagePath := range tempGallery {
		oneImagePath, _ = filepath.Rel("./user", oneImagePath)
		oneImagePath = filepath.ToSlash(oneImagePath)
		gallery = append(gallery, oneImagePath)
	}

	c.HTML(http.StatusOK, "otheruserprofile.html", gin.H{
		"profileID":      profileID,
		"avatarPath":     relPath,
		"avatarUsername": relPathUsername, //for chat thumbnail
		"profilestruct":  querriedProfile,
		"addButtonValue": addButValue,
		"friends":        querriedFriendList,
		"posts":          hisLatestPosts,
		"username":       username,
		"gallery":        gallery,
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

	c.Redirect(http.StatusFound, "/profile")
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
		"username":      username,
		"bio":           querriedProfile.bio.String,
		"lastname":      querriedProfile.lastname.String,
		"gender":        querriedProfile.gender.String,
		"birthday":      querriedProfile.birthday.Time.Format("2006-01-02"),
		"mobilenumber":  querriedProfile.mobileNumber.String,
		"country":       querriedProfile.country.String,
		"statusMessage": statusMessage,
		"avatarPath":    relPath,
	})
}

//GetLoadMoreAtHome is loadmore button handler at homepage
func GetLoadMoreAtHome(c *gin.Context) {

	username, err := c.Cookie("uid")
	if err == http.ErrNoCookie {
		log.Println("No cookie error: ", err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), TIMEOUT)
	defer cancel()

	pageStr := c.Param("page")
	pageNum, err := strconv.Atoi(pageStr)
	if err != nil {
		log.Println("strconvAtoi at GetLoadMore failed", err)
		return
	}

	loadMorePost, err := LoadMoreWithOffset(ctx, username, pageNum)
	if err != nil {
		log.Println("load more post failed:", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"LoadMorePost": loadMorePost,
	})

}

//GetLoadMoreAtProfile function handles load more post feature for my profile page via XMLHTTPRequest
func GetLoadMoreAtProfile(c *gin.Context) {

	username, err := c.Cookie("uid")
	if err == http.ErrNoCookie {
		log.Println("No cookie error: ", err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), TIMEOUT)
	defer cancel()

	pageStr := c.Param("page")
	pageNum, err := strconv.Atoi(pageStr)
	if err != nil {
		log.Println("strconvAtoi at GetLoadMorehome failed", err)
		return
	}
	loadMorePost, err := LoadMoreWithOffsetAllSameUsername(ctx, username, pageNum)
	if err != nil {
		log.Println("load more post failed:", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"LoadMorePost": loadMorePost,
	})

}

//GetLoadMoreByUsername function handles load more post feature for other user profile page via XMLHTTPRequest
func GetLoadMoreByUsername(c *gin.Context) {

	_, err := c.Cookie("uid")
	if err == http.ErrNoCookie {
		log.Println("No cookie error: ", err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), TIMEOUT)
	defer cancel()

	pageStr := c.Param("page")
	pageNum, err := strconv.Atoi(pageStr)
	if err != nil {
		log.Println("strconvAtoi at Get Load More username failed", err)
		return
	}
	thirdUsername := c.Param("profileID")
	loadMorePost, err := LoadMoreWithOffsetAllSameUsername(ctx, thirdUsername, pageNum)
	if err != nil {
		log.Println("load more post failed:", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"LoadMorePost": loadMorePost,
	})

}

//DelPostId function deletes the related post
func DelPostId(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), TIMEOUT)
	defer cancel()
	username, err := c.Cookie("uid")
	if err == http.ErrNoCookie {
		log.Println("No cookie error: ", err)
		return
	}
	postIdToBeDeleted := c.Param("postId")
	err = DeleteThisPost(ctx, username, postIdToBeDeleted)
	if err != nil {
		log.Println("Post couldn't be deleted:", err)
		return
	}
	c.String(http.StatusOK, "")
	return
}

//PostSearchUser is handler for user search requests
func PostSearchUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), TIMEOUT)
	defer cancel()
	lettersToBeSearched := SearchLetters{}
	err = c.Bind(&lettersToBeSearched)
	if err != nil && err != io.EOF {
		log.Println("binding json failed post seach user:", err)
		return
	}
	filteredQuery, err := FilterUsersByLetters(ctx, lettersToBeSearched)
	if err != nil {
		log.Println("filtering failed:")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"filtered": filteredQuery,
	})
}

//PostIt function handles posting service
func PostIt(c *gin.Context) {
	username, err := c.Cookie("uid")
	if err == http.ErrNoCookie {
		log.Println("No cookie error: ", err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), TIMEOUT)
	defer cancel()
	shortStatusMessage := ""

	post := PostThatBeSaved{}
	post.Postername = username
	tempPostID, err := uuid.NewV4()
	if err != nil {
		log.Println("v4 uuid generate failed:", err)
	}
	post.PostId = tempPostID.String()
	post.PostTime = time.Now()
	//.Format("2006-01-02 03:04:05")
	post.PostMessage = c.PostForm("postmessage")

	//get youtube embed link if any exist in post message
	if ytLink, ok := SearchForYoutube(post.PostMessage); ok {
		tempLink, oki := GetYtEmbed(ytLink)
		if !oki && tempLink == "invalidurl" {
			shortStatusMessage = fmt.Sprintf("%sinvalid youtube url:", shortStatusMessage)
			c.SetCookie("short_status_message", shortStatusMessage, 60, "/", Server_Host, false, true)
		} else {
			post.PostYtEmbedLink = tempLink
		}
	}

	//process and upload image if there is any
	imgToBeUploaded, header, err := c.Request.FormFile("postimage")
	if err != nil || header.Size == 0 {
		log.Println("image upload failed or no such file:", err)
	} else { //if there is image to upload
		filename := username + RandomString(4) + filepath.Ext(header.Filename)
		filePathString := fmt.Sprintf("./private/assets/postImages/%s/", username)
		if err = os.MkdirAll(filePathString, 0666); err != nil {
			log.Println("filepath for post images couldn't be created")
			shortStatusMessage = fmt.Sprintf("%s\npost image couldn't be uploaded:%s", shortStatusMessage, err.Error())
			c.SetCookie("short_status_message", shortStatusMessage, 60, "/", Server_Host, false, true)
			c.Redirect(http.StatusFound, "/home")
		}
		if err = ResizeAndSave(imgToBeUploaded, filePathString, filename); err != nil {
			shortStatusMessage = fmt.Sprintf("%s\npost image couldn't be uploaded:%s", shortStatusMessage, err.Error())
			c.SetCookie("short_status_message", "post image couldn't be uploaded:"+err.Error(), 60, "/", Server_Host, false, true)
			c.Redirect(http.StatusFound, "/home")
		}
		post.PostImageFilepath = filePathString + filename

		defer imgToBeUploaded.Close()
	}
	if _, err = conn.Exec(ctx, "INSERT INTO posts (post_id, postername, post_message, post_time, post_yt_embed_link, post_image_filepath) VALUES ($1,$2,$3,$4,$5,$6);", post.PostId, post.Postername, post.PostMessage, post.PostTime.Format(time.RFC3339), post.PostYtEmbedLink, post.PostImageFilepath); err != nil {
		log.Println("add post to DB was failed:", err)
		return
	}

	c.Redirect(http.StatusFound, "/home")
}

//PostAddUnfriend is handler func for add or delete some from friends list
func PostAddUnfriend(c *gin.Context) {

	username, err := c.Cookie("uid")
	if err == http.ErrNoCookie {
		log.Println("No cookie error: ", err)
		return
	}
	friend := Relationship{}

	err = c.Bind(&friend)
	if err != nil && err != io.EOF {
		log.Println("binding json failed:", err)
		return
	}
	if username == friend.Friendname {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), TIMEOUT)
	defer cancel()
	ok := QueryFriendship(ctx, username, friend.Friendname)
	if !ok { //add as friend
		friend.Since.Time = time.Now()
		friend.Username = username

		_, err = conn.Exec(ctx, "INSERT INTO relations (username,friendname,since) VALUES ($1,$2,$3);", friend.Username, friend.Friendname, friend.Since.Time.Format("2006-01-02")) //jsonb array için {}kullanıyor. JSONB İÇİN []  COALESCE
		if err != nil {
			log.Println("update friends list was failed1:", err)
			return
		}
		// ;
		_, err = conn.Exec(ctx, "INSERT INTO relations (username,friendname,since) VALUES ($1,$2,$3)", friend.Friendname, friend.Username, friend.Since.Time.Format("2006-01-02")) //jsonb array için {}kullanıyor. JSONB İÇİN []  COALESCE
		if err != nil {
			log.Println("update friends list was failed1:", err)
			return
		}

		c.String(http.StatusOK, "true")
		return
	}
	//unfriend for both username and thatProfile
	err = UnfriendQuery(ctx, username, friend.Friendname)
	if err != nil {
		log.Println("unfriend failed1", err)
	}

	c.String(http.StatusOK, "false")
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

	ctx, cancel := context.WithTimeout(c.Request.Context(), TIMEOUT)
	defer cancel()
	_, err = conn.Exec(ctx, "UPDATE user_creds SET name = $1,lastname = $2,gender = $3,birthday = $4,mobilenumber = $5,country = $6 WHERE username = $7;", upFirstname, upLastname, upGender, upBirthday, upMobile, upCountry, username)
	if err != nil {
		log.Println("user profile update has failed:", err)
		return
	}
	shortStatusMessage := "Updated successfully!"
	c.SetCookie("short_status_message", shortStatusMessage, 60, "/", "localhost", false, true)

	c.Redirect(http.StatusFound, "/settings")

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
		c.SetCookie("short_status_message", "avatar photo couldn't be uploaded:"+err.Error(), 60, "/", Server_Host, false, true)
		return
	}

	//resize and save file
	err = ResizeAndSave(picToBeUploaded, filePathString, filename)
	if err != nil {
		c.SetCookie("short_status_message", "avatar photo couldn't be uploaded:"+err.Error(), 60, "/", Server_Host, false, true)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	_, err = conn.Exec(ctx, "UPDATE user_creds SET avatarpath = $1 WHERE username= $2;", filePathString+filename, username)
	if err != nil {
		log.Println("avatarpath update to DB was failed:", err)
		c.SetCookie("short_status_message", "avatar path couldn't be saved:"+err.Error(), 60, "/", "localhost", false, true)
		return
	}
	c.SetCookie("short_status_message", "Profile picture has been updated successfully!", 30, "/", "localhost", false, true)
	c.Redirect(http.StatusFound, "/settings")

}

//PostUpdateBio func is for handling edit profile bio
func PostUpdateBio(c *gin.Context) {
	newBioMessage := *Striper(c.PostForm("bioMessage"))
	username, err := c.Cookie("uid")
	if err == http.ErrNoCookie {
		log.Println("No cookie error: ", err)
		return
	}
	err = UpdateMyBio(c.Request.Context(), newBioMessage, username)
	if err != nil {
		c.SetCookie("short_status_message", "bio couldn't be updated", 60, "/", Server_Host, false, true)
	}
	c.SetCookie("short_status_message", "bio updates succesfully...", 60, "/", Server_Host, false, true)
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

	//checkLoginQuery := fmt.Sprintf("SELECT password FROM user_creds WHERE username LIKE '%s' ;", username)
	hashedToBeChecked, err := QueryLog(c, username)
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

	//updatePassword := fmt.Sprintf("UPDATE user_creds SET password = '%s' WHERE username = '%s';", newHashed, username)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	_, err = conn.Exec(ctx, "UPDATE user_creds SET password = $1 WHERE username = $2 ;", newHashed, username)
	if err != nil {
		c.SetCookie("short_status_message", "Password change failed!"+err.Error(), 30, "/", Server_Host, false, true)
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
			c.SetCookie("short_status_message", "Delete failed:"+err.Error(), 30, "/", Server_Host, false, true)
			c.Redirect(http.StatusFound, "/settings")
			return
		}
	}

	hashedToBeChecked, err := QueryLog(c, username)
	if err != nil {
		c.SetCookie("short_status_message", "Something went wrong with DB/Query(pd):!"+err.Error(), 30, "/", Server_Host, false, true)
		c.Redirect(http.StatusFound, "/settings")
		return
	}
	if !CheckPasswordHash(confPass, hashedToBeChecked) {
		c.SetCookie("short_status_message", "Password is incorrect!", 30, "/", Server_Host, false, true)
		c.Redirect(http.StatusFound, "/settings")
		return
	}

	_, err = conn.Exec(ctx, "DELETE FROM user_creds WHERE username = $1 ;", username)
	if err != nil {
		c.SetCookie("short_status_message", "Delete failed!"+err.Error(), 30, "/", Server_Host, false, true)
		c.Redirect(http.StatusFound, "/settings")
		return
	}
	//delete from relations table
	_, err = conn.Exec(ctx, "DELETE FROM relations WHERE username = $1 or friendname = $2 ;", username, username)
	if err != nil {
		c.SetCookie("short_status_message", "Delete failed!"+err.Error(), 30, "/", Server_Host, false, true)
		c.Redirect(http.StatusFound, "/settings")
		return
	}

	c.Redirect(http.StatusFound, "/logout")

}

//PostCheckAuth is the handler function for sign in check : Authentication
func PostCheckAuth(c *gin.Context) {
	if CheckSession(c) {
		return
	}
	logUserName := *Striper(c.PostForm("usernameL"))
	logPassword := *Striper(c.PostForm("passwordL"))
	//check login credential
	hashedToBeChecked, err := QueryLog(c, logUserName)
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

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	_, err = conn.Exec(ctx, "UPDATE user_creds SET lastlogin = $1 WHERE username = $2;", lastLogTime, logUserName)
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

	ctx, cancel := context.WithTimeout(c.Request.Context(), TIMEOUT)
	defer cancel()
	res, err := conn.Exec(ctx, "INSERT INTO user_creds (user_id,username,password,createdon,email) VALUES (nextval('user_id_seq'),$1,$2,$3,$4);", regUserName, regPassword, regCreatedOn, regEmail)

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
