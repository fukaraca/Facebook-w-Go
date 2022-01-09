package lib

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
	"time"
)

var upgrader = websocket.Upgrader{
	HandshakeTimeout: TIMEOUT,
	CheckOrigin: func(r *http.Request) bool {
		//host'u (origin) kontrol eder.
		/*		origin := r.Header.Get("Origin")
				return origin == "http://localhost:8080" */

		return true
	},
	Subprotocols:    nil, //burada belirtilen subprotocolleri client ile negotiate eder
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type sessionInfo struct {
	sessionId      string
	keyCandidate   string
	fieldCandidate string
}

var m sync.Mutex

//WsChat is function that handles chat websocket API process
func WsChat(c *gin.Context) {
	//response header cookie ve subprotocol için kullanılır
	done := make(chan bool)
	push := make(chan bool)
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("upgrade failed:", err)
	}
	defer ws.Close()
	//defer ws.Close()
	//this ctx will be used for main chain
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()
	//context to be used timed queries
	ctxT, cancelT := context.WithTimeout(ctx, TIMEOUT)
	defer cancelT()

	username, err := c.Cookie("uid")
	if err == http.ErrNoCookie {
		log.Println("No cookie error: ", err)
		return
	}

	//get receiverID on open
	_, rcv_B, err := ws.ReadMessage()
	if err != nil {
		log.Println("receiver id couldn't be get", err)
		return
	}
	rcv_str := string(rcv_B)
	sessionIDstr := ""

	//to prevent possible race conditioning, assign lexicographical noble as key
	keyCandidate := ""
	fieldCandidate := ""
	if username < rcv_str {
		keyCandidate = username
		fieldCandidate = rcv_str
	} else {
		keyCandidate = rcv_str
		fieldCandidate = username
	}

	//check existance of active session in redis with username as keyCandidate and msg receiver as fieldCandidate. uuid(sess_id) will be value of that field
	m.Lock()
	ok, err := cache.HExists(ctxT, keyCandidate, fieldCandidate).Result()
	if err != nil {
		log.Println("check exist of active session failed:", err)
		return
	}
	if !ok { //there is no active session
		sessionID, err := uuid.NewV4()
		if err != nil {
			log.Println("session id creation failed:", err)
			return
		}

		res := cache.HSet(ctxT, keyCandidate, fieldCandidate, sessionID.String())
		sessionIDstr = sessionID.String()
		if num, err := res.Result(); num != 1 {
			log.Println("creating a session to redis failed:", err)
			return
		}

		lenSomeMsgBytes := 0
		if someMsgByte, err := BringMeSomeMessages(ctx, keyCandidate, fieldCandidate); err != fmt.Errorf("no row") {
			if err != nil {
				log.Println("bring me some messages query failed:", err)
				return
			}
			lenSomeMsgBytes = len(someMsgByte)
			for i := lenSomeMsgBytes - 1; i >= 0; i-- {
				err = ws.WriteMessage(1, someMsgByte[i])
				if err != nil {
					log.Println("writing initial messsages from DB failed:", err)
					return
				}
				_, err = cache.LPush(ctxT, sessionIDstr, someMsgByte[i]).Result()
				if err != nil {
					log.Println("push to redis failed", err)
					return
				}
			}

		}
		tempSessionInfo := sessionInfo{
			sessionId:      sessionIDstr,
			keyCandidate:   keyCandidate,
			fieldCandidate: fieldCandidate,
		}

		go reCollector(tempSessionInfo, int64(lenSomeMsgBytes))
		m.Unlock()
	} else {

		//this means there is active session in redis
		//get sessionID for noble keyname
		sessionIDstr, err = cache.HGet(ctxT, keyCandidate, fieldCandidate).Result()
		if err != nil {
			log.Println("get sess_id from redis failed:", err)
			return
		}

		//get all messages from active session in redis db
		listMsgbyte := [][]byte{}
		err := cache.LRange(ctxT, sessionIDstr, 0, -1).ScanSlice(&listMsgbyte)
		if err != nil {
			log.Println("messages couldn't be get from redis", err)
			return
		}
		for i := len(listMsgbyte) - 1; i >= 0; i-- {
			err = ws.WriteMessage(1, listMsgbyte[i])
			if err != nil {
				log.Println("writing initial messsages from DB failed:", err)
				return
			}
		}
		m.Unlock()
	}

	go controller(ws, ctx, sessionIDstr, push)
	go reader(ctx, ws, sessionIDstr, done)
	go writer(ctx, ws, sessionIDstr, done, push)

	//blocks until done or endit signal
	select {
	case <-ctx.Done():
		switch ctx.Err() {
		case context.DeadlineExceeded:
			log.Println("context timeout exceeded context done final")
		case context.Canceled:
			log.Println("context cancelled by force. whole process is complete ctxdone final")
		}
	}

}

//reader function reads messages from client and after error or close send done signal to writer
func reader(ctx context.Context, ws *websocket.Conn, sessionIDstr string, done chan<- bool) {
	defer close(done)
	//en son en üstteki defer calısır
	for {
		msgtype, msgByte, err := ws.ReadMessage()
		if err != nil {
			log.Println("read message failed maybe due to closing websocket:", err, msgtype)
			done <- true
			return
		}
		ctxR, cancelTempR := context.WithTimeout(ctx, 2*time.Second)
		_, err = cache.RPush(ctxR, sessionIDstr, msgByte).Result()
		cancelTempR()
		if err != nil {
			log.Println("push to redis failed", err)
			return
		}

	}
}

//writer function is goroutine that loops and writes messages that transfered to redis from reader function
func writer(ctx context.Context, ws *websocket.Conn, sessionIDstr string, done, push <-chan bool) {
	defer ws.Close()

	for {

		select {
		case <-done:
			// the reader is done, so return
			return
		case <-push:
			// get data from channel
			ctxW, cancel := context.WithTimeout(ctx, 2*time.Second)
			latestMsg, err := cache.LIndex(ctxW, sessionIDstr, -1).Bytes()
			cancel()
			if err != nil {
				log.Println("latest message couldn't be read from redis:", err)
				return
			}

			err = ws.WriteMessage(1, latestMsg)
			if err != nil {
				log.Println("write mes err:", err)
				return
			}
		}
	}
}

//controller function checks by certain timeouts whether new message is sent or not. If yes, sends true to push channel
func controller(ws *websocket.Conn, ctx context.Context, sessionIDstr string, push chan<- bool) {
	defer ws.Close()
	//redis notification subscription
	pSubChanStr := fmt.Sprintf("__key*__:%s", sessionIDstr)
	pubSub := cache.PSubscribe(ctx, pSubChanStr)
	eventsPubsub := pubSub.Channel()
	for notif := range eventsPubsub {

		select {
		case <-ctx.Done():
			close(push)
			pubSub.Close()
			return
		default:
			switch notif.Payload {
			case "del":
				err = pubSub.PUnsubscribe(context.Background(), pSubChanStr)
				if err != nil {
					log.Println("resubscription failed:", err)
				}
				err = pubSub.Close()
				close(push)
				return
			case "rpush":
				push <- true

			}
		}
	}

}

//reCollector function saves messages in the redis to DB within certain intervals.
func reCollector(info sessionInfo, lenInitialMsg int64) {
	//recollection interval
	time.Sleep(600 * time.Second)
	//new context is used
	ctxR, cancel := context.WithTimeout(Ctx, TIMEOUT)
	defer cancel()
	_, err := cache.Del(ctxR, info.keyCandidate).Result()
	if err != nil {
		log.Println("deletion of session pointer failed:", err)
	}

	listMsgByte := [][]byte{}
	err = cache.LRange(ctxR, info.sessionId, lenInitialMsg, -1).ScanSlice(&listMsgByte)
	if err != nil {
		log.Println("messages couldn't be get from redis", err)
		return
	}
	_, err = cache.Del(ctxR, info.sessionId).Result()
	if err != nil {
		log.Println("deletion of session failed:", err)
	}

	for _, msgSingle := range listMsgByte {
		_, err = conn.Exec(ctxR, "INSERT INTO messages (key_candidate, field_candidate, message) VALUES ($1,$2,$3);", info.keyCandidate, info.fieldCandidate, msgSingle)
		if err != nil {
			log.Println("re-writing to DB from redis failed:", err)
			return
		}
	}

}
