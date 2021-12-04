package lib

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)

//NoRoute404 is handler function for 404 Page
func NoRoute404(c *gin.Context) {
	c.HTML(http.StatusNotFound,"404.html",nil)
}

//GetIndex is handler function for homepage or startpage
func GetIndex(c *gin.Context) {
	if !CheckSession(c){
		c.HTML(http.StatusUnauthorized,"index.html",nil)
		return
	}
	GetHome(c)
}

//PostCheckAuth is the handler function for login Authentication
func PostCheckAuth(c *gin.Context){
	if CheckSession(c){
		GetHome(c)
		return
	}
	logUserName:=*Striper(c.PostForm("usernameL"))
	logPassword:=*Striper(c.PostForm("passwordL"))
	//check login credential
	checkLoginQuery:=fmt.Sprintf("SELECT password FROM user_creds WHERE username LIKE '%s' ;",logUserName)
	hashedToBeChecked,err:=QueryLog(checkLoginQuery)
	if err != nil {
		loginMessage:=err.Error()+" Something went wrong with DB/Query!"
		c.HTML(http.StatusBadGateway,"index.html",gin.H{
			"messageL":loginMessage,
		})
		return
	}
	if!CheckPasswordHash(logPassword,hashedToBeChecked){
		loginMessage:="Password or Username is incorrect"
		c.HTML(http.StatusForbidden,"index.html",gin.H{
			"messageL":loginMessage,
		})
		return
	}
	lastLogTime:=time.Now().Format(time.RFC3339)
	insertLastLog:=fmt.Sprintf("UPDATE user_creds SET lastlogin = '%s' WHERE username = '%s';",lastLogTime,logUserName)
	_,err=conn.Exec(Ctx,insertLastLog)
	if err != nil {
		log.Println("last login time insertion has error:",err)
	}
	SetCookie(logUserName,c)
	c.HTML(http.StatusOK,"home.html",gin.H{
		"message":"Welcome "+logUserName,
	})
}
//PostCheckReg is handler function to Register new user
func PostCheckReg(c *gin.Context){
	if CheckSession(c){
		GetHome(c)
		return
	}
	regUserName:=*Striper(c.PostForm("usernameReg"))
	tempPass:=*Striper(c.PostForm("passwordReg"))
	regPassword,err:= HashPassword(tempPass)
	regEmail:= *Striper(c.PostForm("emailReg"))
	regCreatedOn:=time.Now().Format(time.RFC3339)
	//insert new users if not exist
	checkRegQuery:=fmt.Sprintf("INSERT INTO user_creds (user_id,username,password,email,createdon) VALUES (nextval('user_id_seq'),'%s','%s','%s','%s');",regUserName,regPassword,regEmail,regCreatedOn)
	res,err:= conn.Exec(Ctx,checkRegQuery)
	messageR:="New account created successfully! You can login anytime."
	//if any error occurs
	if err != nil {
		messageR=QueryErr(err)
		log.Println("Registration error :",messageR)
		c.HTML(http.StatusUnauthorized,"index.html",gin.H{
			"messageR":messageR,
		})
		return
	}
	//if it's OK
	log.Println("register Query:",res.String())
	c.HTML(http.StatusOK,"index.html",gin.H{
		"messageR":messageR,
	})

}


func GetHome(c *gin.Context) {
	if !CheckSession(c){
		GetIndex(c)
		return
	}

	c.HTML(200,"home.html",nil)

}


