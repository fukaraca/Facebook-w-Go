package lib

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//QueryErr Function for converting query errors into much human readible strings
func QueryErr(err error) string {
	var returnee string
	pgerr, _ := err.(*pgconn.PgError)

	switch {
	case strings.Contains(pgerr.Message, "null value") && strings.Contains(pgerr.Message, "password"):
		returnee = "Password can't be left empty"

	case strings.Contains(pgerr.Message, "null value") && strings.Contains(pgerr.Message, "email"):
		returnee = "E-mail can't be left empty"

	case strings.Contains(pgerr.Message, "null value") && strings.Contains(pgerr.Message, "username"):
		returnee = "Username can't be left empty"

	case strings.Contains(pgerr.ConstraintName, "username"):
		returnee = "Username alredy exist"

	case strings.Contains(pgerr.ConstraintName, "email"):
		returnee = "Email already exist"

	default:
		returnee = pgerr.Message
	}
	return returnee
}

//QueryLog Function for querying login credentials from DB
func QueryLog(c *gin.Context, username string) (string, error) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	result, err := conn.Query(ctx, "SELECT password FROM user_creds WHERE username LIKE $1 ;", username)
	defer result.Close()
	if err != nil {
		log.Println("Query Log conn.query:", err)
	}

	password := ""
	for result.Next() {
		if err := result.Scan(&password); err == pgx.ErrNoRows {
			return "", fmt.Errorf("username not found")
		} else if err == nil {
			return password, nil
		} else {
			return "", err
		}

	}
	return "", err
}

//QueryUsername func checks if given username is exist in DB
func QueryUsername(ctx context.Context, username string) bool {
	result, err := conn.Query(ctx, "SELECT EXISTS(SELECT 1 FROM user_creds WHERE username=$1);", username)
	defer result.Close()
	if err != nil {
		log.Println("Query username conn.query:", err)
		return false
	}
	ok := false
	for result.Next() {
		if err := result.Scan(&ok); err != nil || !ok {
			return false
		}
	}
	return ok
}

//BringMeProfile function simply brings the profile from DB
func BringMeProfile(ctx context.Context, username string) (userCred, error) {
	tempQurryed := userCred{}
	tempBirthday := pgtype.Date{}
	bringy := conn.QueryRow(ctx, "SELECT name,lastname,mobilenumber,country,birthday,gender,avatarpath,bio FROM user_creds WHERE username=$1;", username)
	err = bringy.Scan(&tempQurryed.name, &tempQurryed.lastname, &tempQurryed.mobileNumber, &tempQurryed.country, &tempBirthday, &tempQurryed.gender, &tempQurryed.avatarPath, &tempQurryed.bio)
	if err != nil {
		return userCred{}, fmt.Errorf("scanning the reloaded user infos from DB is failed:%s", err.Error())
	}
	tempQurryed.birthday.Time, err = time.Parse("2006-01-02", tempBirthday.Time.Format("2006-01-02"))

	return tempQurryed, err
}

//BringMeAvatar function brings avatar from private filesystem
func BringMeAvatar(privateavtPath, username string) (string, error) {

	_, filename := filepath.Split(privateavtPath)
	publicAvatarPath := fmt.Sprintf("./web/asset/avatars/%s/%s", username, filename)
	publicAvatarDirPath := fmt.Sprintf("./web/asset/avatars/%s", username)
	//file exists
	if _, err := os.Stat(publicAvatarPath); err == nil {
		relPath, _ := filepath.Rel("./web", publicAvatarPath)
		relPath = filepath.ToSlash(relPath)
		return relPath, nil
	}
	// file does *not* exist
	avatarImage, err := os.OpenFile(privateavtPath, os.O_RDONLY, 0444)
	if err != nil {
		return defaultAvatar, fmt.Errorf("private avatar file couldn't be opened:%s", err.Error())
	}
	defer avatarImage.Close()

	err = os.MkdirAll(publicAvatarDirPath, 0666)
	if err != nil {
		return defaultAvatar, fmt.Errorf("public avatar dir couldn't be created:%s", err.Error())
	}

	tempAvatar, err := os.Create(publicAvatarPath) //0 boyutlu file oluşturuluyor
	if err != nil {
		return defaultAvatar, fmt.Errorf("public avatar zero-file couldn't be created:%s", err.Error())
	}
	defer tempAvatar.Close()

	_, err = io.Copy(tempAvatar, avatarImage) //0 boyutlu dosyaya kopyalanır
	if err != nil {
		return defaultAvatar, fmt.Errorf("avatar file couldn't be copied:%s", err.Error())
	}

	//get relative path for avatar
	relPath, _ := filepath.Rel("./web", publicAvatarPath)
	relPath = filepath.ToSlash(relPath)

	return relPath, nil
}

//BringMeImage function brings gallery from private filesystem with n number of item
func BringHisGallery(username string, n int) []string {
	gallery := []string{}
	//is gallery copied into public directory readily
	/*publicGalleryPath := fmt.Sprintf("./web/asset/postImages/%s", username)
	if _, err := os.Stat(publicGalleryPath); !os.IsNotExist(err) {
		if gallery, err = ListFilesInDir(publicGalleryPath, n); err != nil {
			return nil
		}
		return gallery
	}*/
	privateGalleryPath := fmt.Sprintf("./private/assets/postImages/%s", username)
	if _, err := os.Stat(privateGalleryPath); os.IsNotExist(err) {
		//log.Printf("there is no image that %s uploaded: %s\n", username, err.Error())
		return nil
	}
	privateGallery, err := ListFilesInDir(privateGalleryPath, n)
	if err != nil {
		log.Println("gallery list couldn't be fetched from private folder", err)
		return nil
	}
	for _, image := range privateGallery {
		tempPrivatePathStr := fmt.Sprintf("./private/assets/postImages/%s/%s", username, image)
		tempPublicImgPath, err := BringMeImage(tempPrivatePathStr, username)
		if err != nil {
			log.Println("image couldn't be migrated from private to public", err)
			return nil
		}
		gallery = append(gallery, tempPublicImgPath)
	}

	return gallery
}

//BringMeImage function brings images from private filesystem
func BringMeImage(privateImgPath, username string) (string, error) {

	_, filename := filepath.Split(privateImgPath)
	publicImagePath := fmt.Sprintf("./web/asset/postImages/%s/%s", username, filename)
	publicImageDirPath := fmt.Sprintf("./web/asset/postImages/%s", username)
	//file exists
	if _, err := os.Stat(publicImagePath); err == nil {
		relPath, _ := filepath.Rel("./web", publicImagePath)
		relPath = filepath.ToSlash(relPath)
		return relPath, nil
	}
	// file does *not* exist
	postImage, err := os.OpenFile(privateImgPath, os.O_RDONLY, 0444)
	if err != nil {
		return "", fmt.Errorf("private image file couldn't be opened:%s", err.Error())
	}
	defer postImage.Close()

	//filePathString := fmt.Sprintf("./web/asset/avatars/%s/", username)
	err = os.MkdirAll(publicImageDirPath, 0666)
	if err != nil {
		return "", fmt.Errorf("public image dir couldn't be created:%s", err.Error())
	}

	tempImage, err := os.Create(publicImagePath) //0 boyutlu file oluşturuluyor
	if err != nil {
		return "", fmt.Errorf("public avatar zero-file couldn't be created:%s", err.Error())
	}
	defer tempImage.Close()

	_, err = io.Copy(tempImage, postImage) //0 boyutlu dosyaya kopyalanır
	if err != nil {
		return "", fmt.Errorf("image file couldn't be copied:%s", err.Error())
	}

	//get relative path for image
	relPath, _ := filepath.Rel("./web", publicImagePath)
	relPath = filepath.ToSlash(relPath)
	return relPath, nil
}

//BringMeThatProfile function brings the given profile from DB.
func BringMeThatProfile(ctx context.Context, profileID string) (EuserCred, error) {
	tempQurryed := EuserCred{}
	tempBirthday := pgtype.Date{}
	bringy := conn.QueryRow(ctx, "SELECT name, lastname, relationship,school,location,workplace,birthday,bio,avatarpath FROM user_creds WHERE username=$1;", profileID)
	err = bringy.Scan(&tempQurryed.Name, &tempQurryed.Lastname, &tempQurryed.Relationship, &tempQurryed.School, &tempQurryed.Location, &tempQurryed.Workplace, &tempBirthday, &tempQurryed.Bio, &tempQurryed.AvatarPath)
	if err != nil {
		return EuserCred{}, fmt.Errorf("scanning the reloaded user infos from DB is failed:%s", err.Error())
	}
	tempQurryed.Birthday.Time, err = time.Parse("2006-01-02", tempBirthday.Time.Format("2006-01-02"))

	return tempQurryed, err
}

//UnfriendQuery unfriends given thatProfile from username and vice-versa
func UnfriendQuery(ctx context.Context, username, thatProfile string) error {

	_, err = conn.Exec(ctx, "DELETE FROM relations WHERE username=$1 AND friendname=$2; DELETE FROM relations WHERE username=$3 AND friendname=$4;", username, thatProfile, thatProfile, username)
	if err != nil {
		log.Println("delete from relations table failed", err)
		return err
	}
	return nil
}

//DeleteThisPost deletes given post if username ise equal postername for related post
func DeleteThisPost(ctx context.Context, username, postId string) error {

	_, err := conn.Exec(ctx, "DO $do$ BEGIN IF (SELECT EXISTS(SELECT 1 FROM posts WHERE post_id = $1 AND postername = $2)) THEN DELETE FROM posts where post_id= $3 ; END IF;  END $do$", postId, username, postId)
	if err != nil {
		log.Println("delete post from posts table failed", err)
		return err
	}
	return nil

}

//BringMeSomeHisPosts function returns posts from certain username
func BringMeSomeMyPosts(ctx context.Context, username string) ([]PostThatBeTemplated, error) {

	rows, err := conn.Query(ctx, "SELECT * FROM posts WHERE postername=$1 ORDER BY post_time DESC LIMIT 10;", username)
	defer rows.Close()
	if err != nil {
		log.Println("get my posts query failed", err)
		return nil, err
	}
	tempBroughtPosts := []PostThatBeTemplated{}
	for i := 0; rows.Next(); i++ {
		tempPost := PostThatBeTemplated{}

		err = rows.Scan(&tempPost.PostId, &tempPost.Postername, &tempPost.PostMessage, &tempPost.PostTime, &tempPost.PostYtEmbedLink, &tempPost.PostImageFilepath)
		if err != nil {
			log.Println("scan row bring me some of my posts failed:", err)
			return nil, err
		}
		if tempPost.PostImageFilepath != "" {
			tempImagePath, err := BringMeImage(string(tempPost.PostImageFilepath), username)
			if err == nil {
				tempPost.PostImageFilepath = template.HTML(fmt.Sprintf("<img src=\"\\%s\" class=\"media-object\" alt=\"Failed to load post image :(\" style=\"max-width: 95%%; max-height: 95%%;\">", tempImagePath))

			}
		}
		tempBroughtPosts = append(tempBroughtPosts, tempPost)
	}
	return tempBroughtPosts, nil
}

//BringMeSomeMessages brings related messages from DB as slice of byte array. keyCandidate priors fieldCandidate lexicographically.
//This is the way to handle logging of chat as unique.
func BringMeSomeMessages(ctx context.Context, keyCandidate, fieldCandidate string) ([][]byte, error) {
	ctxn, cancel := context.WithTimeout(ctx, TIMEOUT)
	defer cancel()
	rows, err := conn.Query(ctxn, "SELECT message FROM messages WHERE key_candidate=$1 AND field_candidate=$2 ORDER BY message_id DESC LIMIT 40;", keyCandidate, fieldCandidate)
	defer rows.Close()
	if err != nil {
		log.Println("bring me some messages query failed:", err)
		return nil, err
	}

	someMsgBytes := [][]byte{}

	for i := 0; rows.Next(); i++ {

		tempMsgByte := []byte{}
		err := rows.Scan(&tempMsgByte)
		if err != nil {
			log.Println("scan row bring me some of my messages failed:", err)
			return nil, err
		}
		someMsgBytes = append(someMsgBytes, tempMsgByte)
	}

	return someMsgBytes, nil
}

//UpdateMyBio updates bio for given bioString and username
func UpdateMyBio(ctx context.Context, bioString, username string) error {
	ctxT, cancel := context.WithTimeout(ctx, TIMEOUT)
	_, err = conn.Exec(ctxT, "UPDATE user_creds SET bio = $1 WHERE username = $2;", bioString, username)
	cancel()
	if err != nil {
		log.Println("update bio failed:", err)
		return err
	}

	return nil
}

//BringMeSomePosts function returns posts from certain username and also from his friends
func BringMeSomePosts(ctx context.Context, username string) ([]PostThatBeTemplated, error) {
	rows, err := conn.Query(ctx, "SELECT * FROM posts WHERE postername IN (SELECT friendname FROM relations WHERE username = $1 UNION SELECT  $2::VARCHAR(50)) ORDER BY post_time DESC LIMIT 10;", username, username)
	defer rows.Close()
	if err != nil {
		log.Println("getFriends query failed", err)
		return nil, err
	}
	tempBroughtPosts := []PostThatBeTemplated{}
	for i := 0; rows.Next(); i++ {
		tempPost := PostThatBeTemplated{}

		err = rows.Scan(&tempPost.PostId, &tempPost.Postername, &tempPost.PostMessage, &tempPost.PostTime, &tempPost.PostYtEmbedLink, &tempPost.PostImageFilepath)
		if err != nil {
			log.Println("scanrow bringmesomeposts failed:", err)
			return nil, err
		}
		if tempPost.PostImageFilepath != "" {
			tempImagePath, err := BringMeImage(string(tempPost.PostImageFilepath), username)
			if err == nil {
				tempPost.PostImageFilepath = template.HTML(fmt.Sprintf("<img src=\"%s\" class=\"media-object\" alt=\"Failed to load post image :(\" style=\"max-width: 95%%; max-height: 95%%;\">", tempImagePath))

			}
		}
		tempBroughtPosts = append(tempBroughtPosts, tempPost)
	}
	return tempBroughtPosts, nil
}

//LoadMoreWithOffset function returns posts from certain username and also from his friends as paginated
func LoadMoreWithOffset(ctx context.Context, username string, pageNum int) ([]ToBeLoadedMore, error) {
	offset := 10 * (pageNum - 1)
	//todo sorun var mı loadMoreQueryStr := fmt.Sprintf("SELECT * FROM posts WHERE postername IN (SELECT friendname FROM relations WHERE username =$1 UNION SELECT $2::VARCHAR(50)) ORDER BY post_time DESC OFFSET $3 LIMIT 10;", username, username, offset)
	rows, err := conn.Query(ctx, "SELECT * FROM posts WHERE postername IN (SELECT friendname FROM relations WHERE username =$1 UNION SELECT $2::VARCHAR(50)) ORDER BY post_time DESC OFFSET $3 LIMIT 10;", username, username, offset)
	defer rows.Close()
	if err != nil {
		log.Println("getFriends query failed", err)
		return nil, err
	}
	tempListToBeLoadedMore := []ToBeLoadedMore{}
	for i := 0; rows.Next(); i++ {
		tempPost := PostThatBeSaved{}

		err = rows.Scan(&tempPost.PostId, &tempPost.Postername, &tempPost.PostMessage, &tempPost.PostTime, &tempPost.PostYtEmbedLink, &tempPost.PostImageFilepath)
		if err != nil {
			log.Println("scanrow bringmesomeposts failed:", err)
			return nil, err
		}
		if tempPost.PostImageFilepath != "" {
			tempImagePath, err := BringMeImage(string(tempPost.PostImageFilepath), username)
			if err == nil {
				tempPost.PostImageFilepath = fmt.Sprintf("<img src=\"%s\" class=\"media-object\" alt=\"Failed to load post image :(\" style=\"max-width: 95%%; max-height: 95%%;\">", tempImagePath)
			}
		}
		tempDispStr := ""
		if tempPost.Postername != username {
			tempDispStr = fmt.Sprintf(`style="display: none"`)
		} else {
			tempDispStr = fmt.Sprintf(`style="display: inline"`)
		}

		tempLoadMore := ToBeLoadedMore(fmt.Sprintf(`
<div class="panel panel-default" style="border-style:dot-dot-dash;border-width:thick;border-bottom-color:#618685">
          <div class="panel-body">
            %s 
            <br>
            <p>%s</p> 
            <br>
            %s  
          </div>
          <div class="panel-footer">
            <span>posted at <b>%s</b> by %s</span>
            <span class="pull-right" %s ><a class="text-danger" href="javascript:void(0)" onclick="deletePost('/delpost/%s')">[delete]</a><a id="dummy-delete" style="display: none" href="" ></a></span>
          </div>

        </div>
        <hr>
`, tempPost.PostImageFilepath, tempPost.PostMessage, tempPost.PostYtEmbedLink, tempPost.PostTime.Format("15:04:05 02/01/2006"), tempPost.Postername, tempDispStr, tempPost.PostId))

		tempListToBeLoadedMore = append(tempListToBeLoadedMore, tempLoadMore)
	}
	return tempListToBeLoadedMore, nil
}

//LoadMoreWithOffsetAllSameUsername function returns posts from certain username as paginated
func LoadMoreWithOffsetAllSameUsername(ctx context.Context, username string, pageNum int) ([]ToBeLoadedMore, error) {
	offset := 10 * (pageNum - 1)
	//todo bakılacak loadMoreQueryStr := fmt.Sprintf("SELECT * FROM posts WHERE postername=$1 ORDER BY post_time DESC OFFSET $2 LIMIT 10;", username, offset)
	rows, err := conn.Query(ctx, "SELECT * FROM posts WHERE postername=$1 ORDER BY post_time DESC OFFSET $2 LIMIT 10;", username, offset)
	defer rows.Close()
	if err != nil {
		log.Println("load friends post query failed", err)
		return nil, err
	}
	tempListToBeLoadedMore := []ToBeLoadedMore{}
	for i := 0; rows.Next(); i++ {
		tempPost := PostThatBeSaved{}

		err = rows.Scan(&tempPost.PostId, &tempPost.Postername, &tempPost.PostMessage, &tempPost.PostTime, &tempPost.PostYtEmbedLink, &tempPost.PostImageFilepath)
		if err != nil {
			log.Println("scanrow bring me some his posts failed:", err)
			return nil, err
		}
		if tempPost.PostImageFilepath != "" {
			tempImagePath, err := BringMeImage(tempPost.PostImageFilepath, username)
			if err == nil {
				tempPost.PostImageFilepath = fmt.Sprintf("<img src=\"\\%s\" class=\"media-object\" alt=\"Failed to load post image :(\" style=\"max-width: 95%%; max-height: 95%%;\">", tempImagePath)
			}
		}
		tempLoadMore := ToBeLoadedMore(fmt.Sprintf(`
<div class="panel panel-default" style="border-style:dot-dot-dash;border-width:thick;border-bottom-color:#618685">
          <div class="panel-body">
            %s 
            <br>
            <p>%s</p> 
            <br>
            %s  
          </div>
          <div class="panel-footer">
            <span>posted at <b>%s</b> by %s</span>
            <span class="pull-right"  ><a class="text-danger" href="javascript:void(0)" onclick="deletePost('/delpost/%s')">[delete]</a><a id="dummy-delete" style="display: none" href="" ></a></span>
          </div>
        </div>
        <hr>
`, tempPost.PostImageFilepath, tempPost.PostMessage, tempPost.PostYtEmbedLink, tempPost.PostTime.Format("15:04:05 02/01/2006"), tempPost.Postername, tempPost.PostId))

		tempListToBeLoadedMore = append(tempListToBeLoadedMore, tempLoadMore)
	}
	return tempListToBeLoadedMore, nil
}

//FilterUsersByLetters function makes query for search by filtering with 'letters' parameter and returns a slice o 3 random match
func FilterUsersByLetters(ctx context.Context, letters SearchLetters) ([]ToBeLoadedMore, error) {
	rows, err := conn.Query(ctx, "SELECT username,name,lastname from user_creds where (username || name) ~* $1 order by random() limit 3;", letters.Letters)
	defer rows.Close()
	if err != nil {
		log.Println("getFriends query failed", err)
		return nil, err
	}
	filteredList := []ToBeLoadedMore{}
	for i := 0; rows.Next(); i++ {
		tempUsername := ""
		tempFirstname := ""
		tempLastname := ""
		err = rows.Scan(&tempUsername, &tempFirstname, &tempLastname)
		if err != nil {
			log.Println("scan row filterusersbyletters failed:", err)
			return nil, err
		}
		dropdownHtmlStr := ToBeLoadedMore(fmt.Sprintf(`<div class="dropdivider"></div>
        <a class="dropcont" href="/user/%s">%s : %s %s</a>`, tempUsername, tempUsername, tempFirstname, tempLastname))
		filteredList = append(filteredList, dropdownHtmlStr)

	}
	return filteredList, nil
}

//BringMeFriends return friends list as slice of struct
func BringMeFriends(ctx context.Context, username string) ([]Relationship, error) {

	res, err := conn.Query(ctx, "SELECT friendname,since FROM relations WHERE username=$1;", username)
	defer res.Close()
	if err != nil {
		log.Println("getFriends query failed", err)
		return nil, err
	}
	tempRelationship := []Relationship{}
	for i := 0; res.Next(); i++ {
		tempRel := Relationship{}

		err = res.Scan(&tempRel.Friendname, &tempRel.Since.Time)
		if err != nil {
			log.Println("scanrow failed:", err)
			return nil, err
		}
		tempRelationship = append(tempRelationship, tempRel)
	}

	return tempRelationship, nil
}

//FindMeSuggestibleFriendsAndAlsoOneOfMine brings 3 random friends who is not friend of the username from DB and also one of his friends for "missed me?" div
func FindMeSuggestibleFriendsAndAlsoOneOfMine(ctx context.Context, username string) ([]string, string, error) {
	res, err := conn.Query(ctx, "SELECT username FROM user_creds AS uc WHERE uc.username NOT IN (SELECT friendname FROM relations WHERE username = $1) AND uc.username <> $2 ORDER BY random() LIMIT 3;", username, username)
	if err != nil {
		log.Println("query error1:", err)
		return nil, "", err
	}
	defer res.Close()
	toBeSuggested := []string{}
	for res.Next() {
		tempSuggestion := ""
		err = res.Scan(&tempSuggestion)
		if err != nil {
			log.Println("scan suggest friend query failed:", err)
			return nil, "", err
		}
		toBeSuggested = append(toBeSuggested, tempSuggestion)
	}
	if err = res.Err(); err != nil {
		log.Println("query error2:", err)
		return nil, "", err
	}
	randomFriend := ""
	row := conn.QueryRow(ctx, "SELECT friendname FROM relations WHERE username= $1 ORDER BY random() LIMIT 1;", username)
	err = row.Scan(&randomFriend)
	if err != nil && err != pgx.ErrNoRows {
		log.Println("getFriends query failed", err)
		return toBeSuggested, "", err
	} else if err == pgx.ErrNoRows {
		log.Println("no friend", err)
		return toBeSuggested, "", err
	}
	return toBeSuggested, randomFriend, nil

}

//QueryFriendship investigates whether username is friend with friendUsername or not
func QueryFriendship(ctx context.Context, username, friendUsername string) bool {

	result, err := conn.Query(ctx, "SELECT EXISTS(SELECT 1 FROM relations WHERE username = $1 AND friendname=$2);", username, friendUsername)
	defer result.Close()
	if err != nil {
		log.Println("Query username conn.query:", err)
		return false
	}
	ok := false
	for result.Next() {
		if err := result.Scan(&ok); err != nil || !ok {
			return false
		}
	}

	if !ok {
		return false
	}
	return true
}
