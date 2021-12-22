package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
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
func QueryLog(c *gin.Context, strQuery string) (string, error) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	result, err := conn.Query(ctx, strQuery)
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

//QueryUsername func checks if given profileID is exist in DB
func QueryUsername(ctx context.Context, str string) bool {
	result, err := conn.Query(ctx, str)
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
	queryProfileString := fmt.Sprintf("SELECT name,lastname,mobilenumber,country,birthday,gender,avatarpath FROM user_creds WHERE username='%s';", username)
	bringy := conn.QueryRow(ctx, queryProfileString)
	err = bringy.Scan(&tempQurryed.name, &tempQurryed.lastname, &tempQurryed.mobileNumber, &tempQurryed.country, &tempBirthday, &tempQurryed.gender, &tempQurryed.avatarPath)
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

	//filePathString := fmt.Sprintf("./web/asset/avatars/%s/", username)
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

//BringMeThatProfile function brings the given profile from DB.
func BringMeThatProfile(ctx context.Context, profileID string) (EuserCred, error) {
	tempQurryed := EuserCred{}
	tempBirthday := pgtype.Date{}
	queryProfileString := fmt.Sprintf("SELECT name, lastname, relationship,school,location,workplace,birthday,bio,avatarpath FROM user_creds WHERE username='%s';", profileID)
	bringy := conn.QueryRow(ctx, queryProfileString)
	err = bringy.Scan(&tempQurryed.Name, &tempQurryed.Lastname, &tempQurryed.Relationship, &tempQurryed.School, &tempQurryed.Location, &tempQurryed.Workplace, &tempBirthday, &tempQurryed.Bio, &tempQurryed.AvatarPath)
	if err != nil {
		return EuserCred{}, fmt.Errorf("scanning the reloaded user infos from DB is failed:%s", err.Error())
	}
	tempQurryed.Birthday.Time, err = time.Parse("2006-01-02", tempBirthday.Time.Format("2006-01-02"))

	return tempQurryed, err
}

//UnfriendQuery unfriends given thatProfile from username
func UnfriendQuery(ctx context.Context, username, thatProfile string) error {
	tempStructFriends, err := BringMeFriends(ctx, username)
	if err != nil {
		log.Println("BringmeFriends failed", err)
		return err
	}
	//unfriend as delete from jsonb
	for i, structFriend := range tempStructFriends {
		if structFriend.FriendID == thatProfile {
			removefunc := func() []FriendWhoToBeAdded {
				newArr := []FriendWhoToBeAdded{}
				for in, el := range tempStructFriends {
					if in != i {
						newArr = append(newArr, el)
					}
				}
				return newArr
			}
			tempStructFriends = removefunc()
			break
		}
	}
	jsonNewFriends, err := json.Marshal(tempStructFriends)
	if err != nil {
		log.Println("marshal to new friend struct failed:", err)
		return err
	}
	updateQueryStr := fmt.Sprintf("UPDATE friends SET friendsince='%s'::JSONB WHERE username='%s';", jsonNewFriends, username)
	_, err = conn.Exec(ctx, updateQueryStr)
	if err != nil {
		log.Println("update new friend list failed", err)
		return err
	}
	return nil
}

//InsertNewFriendRowForUser prepares DB row for username. This is necessary for preventing null error
func InsertNewFriendRowForUser(ctx context.Context, username string) error {
	addUsernameIfNotExist := fmt.Sprintf("INSERT INTO friends (username) VALUES ('%s') ON CONFLICT (username) DO NOTHING;", username)
	_, err = conn.Exec(ctx, addUsernameIfNotExist)
	if err != nil {
		log.Println("insert username if not exist query was failed:", err)
		return err
	}
	return nil
}

//BringMeFriends return friends list as slice of struct
func BringMeFriends(ctx context.Context, username string) ([]FriendWhoToBeAdded, error) {

	err := InsertNewFriendRowForUser(ctx, username)
	if err != nil {
		log.Println("insert username if not exist query was failed:", err)
		return nil, err
	}

	getFriends := fmt.Sprintf("SELECT friendsince FROM friends WHERE username='%s';", username)
	res, err := conn.Query(ctx, getFriends)
	defer res.Close()
	if err != nil {
		log.Println("getFriends query failed", err)
		return nil, err
	}
	tempJSONB := pgtype.JSONB{}
	for res.Next() {
		err = res.Scan(&tempJSONB)
		if err != nil {
			log.Println("scanrow failed:", err)
			return nil, err
		}
	}
	tempStructFriends := []FriendWhoToBeAdded{}
	err = json.Unmarshal(tempJSONB.Bytes, &tempStructFriends)
	if err != nil {
		log.Println("unmarshal addfriendsfunc failed:", err)
		return nil, err
	}
	return tempStructFriends, nil
}

//QueryFriendship investigates whether username is friend with friendUsername or not
func QueryFriendship(ctx context.Context, username, friendUsername string) bool {
	queryStr := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM friends,jsonb_to_recordset(friends.friendsince) AS items(friendid text) WHERE items.friendid = '%s' AND friends.username='%s');", friendUsername, username)
	ok := QueryUsername(ctx, queryStr) //check that if that profile is your friend or not
	if !ok {
		return false
	}
	return true
}
