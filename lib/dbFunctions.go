package lib

import (
	con "context"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"log"
	"strings"
)
//QueryErr Function for converting query errors into much human readible strings
func QueryErr(err error) string{
	var returnee string
	pgerr,_:=err.(*pgconn.PgError)

	switch {
	case strings.Contains(pgerr.Message, "null value") && strings.Contains(pgerr.Message, "password"):
		returnee ="Password can't be left empty"

	case strings.Contains(pgerr.Message, "null value") && strings.Contains(pgerr.Message, "email"):
		returnee ="E-mail can't be left empty"

	case strings.Contains(pgerr.Message, "null value") && strings.Contains(pgerr.Message, "username"):
		returnee ="Username can't be left empty"

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
func QueryLog(strQuery string)(string,error){
	result,err:= conn.Query(con.Background(),strQuery)
	defer result.Close()
	if err != nil {
		log.Println("Query Log conn.query:",err)
	}
	password:=""
	for result.Next(){
		if err:=result.Scan(&password);err==pgx.ErrNoRows{
			return "",fmt.Errorf("username not found")
		} else if err==nil{
			return password,nil
		}else {
			return "",err
		}

	}
	return "",err
}
//
