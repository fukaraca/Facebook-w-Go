
let thatProfileId = document.getElementById('profile-id').innerText;
let clientID=document.getElementById('user-id').innerText
let textAreaChat = document.getElementById('chat-textarea');
let chatButton = document.getElementById('chat-post-button');
let chatBoxDiv=document.getElementById('chat-div')
let notificationCounter = 0
let toggleNotificationOn=false

textAreaChat.addEventListener("keyup", function(event) {
    // Number 13 is the "Enter" key on the keyboard
    if (event.keyCode === 13) {
        // Cancel the default action, if needed
        event.preventDefault();
        // Trigger the button element with a click
        chatButton.click();
    }
});
//ayrı js 'te
function openChat(){

    if (chatBoxDiv.style.display==='none'){
        chatBoxDiv.style.display='block'
        document.getElementById('messages-div').scrollTop = document.getElementById('messages-div').scrollHeight;
        notificationCounter=0
        toggleNotificationOn=false

    } else {
        chatBoxDiv.style.display='none'
        notificationCounter=0
        toggleNotificationOn=true

    }
}

//bu düzende olmak zorunda değil.
let ws=null
webSocketChat()

function wsOpen(event){
    ws.send(thatProfileId);
    document.getElementById('messages-div').innerHTML=""
    document.getElementById('ws-status-p').innerText='chat online!'
    document.getElementById('ws-con-close').innerText="Disconnect"
}
function wsOnMessage(event){
    taken=JSON.parse(event.data)
    if (taken.msgSender===clientID){//user is also sender
        let usernamePath=document.getElementById('avatar-username').innerText
        let elem=String.format('<div class="chat-left"><div class="chat-left-text"><img src={0} class="thumb-left-img">{1}</div><div class="time-text-left">{2}</div></div>',usernamePath,taken.msgBody,timeConverter(taken.msgTime))

        document.getElementById('messages-div').insertAdjacentHTML("beforeend", elem);

    }else {//querried user is sender
        let avtPath=document.getElementById('avatar-img').src
        let elem=String.format(`<div class="chat-right">
                            <div class="chat-right-text">
                                <img src="{0}" class="thumb-right-img" alt="thumbnail">
                                {1}
                            </div>
                            <div class="time-text-right">{2}</div>
                        </div>`,avtPath,taken.msgBody,timeConverter(taken.msgTime))
        document.getElementById('messages-div').insertAdjacentHTML("beforeend", elem);
    }
    document.getElementById('messages-div').scrollTop = document.getElementById('messages-div').scrollHeight;

    notificationChat()
}

function wsOnClose(event){
    /*if (event.wasClean){
        console.log(`[close] connection closed gracefully, code=${event.code} reason=${event.reason}`)
    } else {
        console.log(`[close] connection died code=${event.code} also=${event.reason}`)
    }*/


    document.getElementById('ws-status-p').innerText='chat offline!'
    document.getElementById('ws-con-close').innerText='Connect';
    ws=null
    wsOnConnect()

}
function wsOnError(error){
    console.log(`[error] this is error message: ${error}`)
    document.getElementById('ws-status-p').innerText='check console for err!'
}
//connect butonuna bağlı event
function wsOnConnect(){

    if (ws===null){
        document.getElementById('ws-status-p').innerText='Connecting!'
        webSocketChat()


    }else if (ws.readyState<2){ //if conn true disconn
        //closing
        document.getElementById('ws-status-p').innerText='Shutting down...'
        setTimeout(function (){
            if (ws.bufferedAmount==0){
                ws.close()
            }else {
                ws.close(1009,"waiting for unbuffering")
            }
        },1000)

    } else if (ws.readyState===2){ //if no conn try to reconn
        document.getElementById('ws-status-p').innerText='Still closing!'
    }
}
//Send buttonuna bağlı onclick eventi
function wsSendPost(){
    if (ws===null){
        return false
    }
    else if (ws.readyState>1){
        document.getElementById('ws-status-p').innerText='No connection!'
        return false
    }
    //create msg object
    let msg=new MsgObj()
    msg.msgBody=document.getElementById('chat-textarea').value
    msg.msgSender=clientID
    msg.msgReceiver=thatProfileId
    msg.msgTime=Date.now()
    ws.send(JSON.stringify(msg))
    document.getElementById('chat-textarea').value="";

}


function webSocketChat() {
    ws = new WebSocket("ws://localhost:8080/chat",[]);

    //event listenerlar
    //socket oluşturulduktan sonra 4 event dinlenir: open,message,error,close

    ws.onopen = function (){wsOpen(event)}

    //event message içerir. gelen mesajı ilgili div'in sonundan öncesine ekliyor. böylece yeni mesaj en altta çıkıyor.
    ws.onmessage = function (){ wsOnMessage(event)}

    //event close'a ilişkin verileri içerir. close codu veya reason gibi
    ws.onclose = function (){wsOnClose(event);}

    //error olunca
    ws.onerror =function (){wsOnError(error);}

}


function MsgObj(sender,receiver,time,message){
    this.msgSender=sender;
    this.msgReceiver=receiver;
    this.msgTime=time;
    this.msgBody=message
}
//formatting string with placeholders
String.format = function() {
    let s = arguments[0];
    for (let i = 0; i < arguments.length - 1; i++) {
        let reg = new RegExp("\\{" + i + "\\}", "gm");
        s = s.replace(reg, arguments[i + 1]);
    }
    return s;
}

function notificationChat(){
    if (document.getElementById('chat-div').style.display==='none'&& toggleNotificationOn===true){
        notificationCounter++
        document.getElementById('chat-open-button').innerText=`Pending messages: ${notificationCounter}`
    }else {
        document.getElementById('chat-open-button').innerText='Chat'
        notificationCounter=0
    }
}

function timeConverter(UNIX_timestamp){
    let a = new Date(UNIX_timestamp);
    let months = ['Jan','Feb','Mar','Apr','May','Jun','Jul','Aug','Sep','Oct','Nov','Dec'];
    let year = a.getFullYear();
    let month = months[a.getMonth()];
    let date = a.getDate();
    let hour = a.getHours();
    let min = a.getMinutes();
    let sec = a.getSeconds();
    let time = date + ' ' + month + ' ' + year + ' ' + hour + ':' + min + ':' + sec ;
    return time;
}

