package lib

import (
	"context"
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

	return tempQurryed, nil
}

//BringMeAvatar function brings avatat from private filesystem
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
