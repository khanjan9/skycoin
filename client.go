/*
*  @file : client
*  @brief: File contains client code, 
*
*/

package main

import (
	"fmt"
	"bufio"
	"os"
	"time"
	"net"
	"strings"
	"io"
	//"reflect"
	"strconv"
	"errors"
) 


//Globals
const reconnectWaitTime = ( 2  * time.Second)

var ClientListReceived bool = false
var MyName string = ""

const debugOn bool  = true 

//return codes (int rv)
const (
	SkySuccess = 0
	SkyFail = -1
)

const (
	ErrorNoVisible = 0
	ErrorVisible = 1
)

const (
	connHost = "localhost"
	connPort = "8082"
	connType = "tcp"
)

//Msg IDs, to be kept in sync between server , client
//each received sent msg has "<msgId>:<msgContent>" format
const (
	clientList = iota // format eg: "<clientList>:pubkey1-pubkey2-pubkey3-" eg: 0:2-4-5-
	newClient  //format: <newClient>:pubkey
	regularMsg  //format: <regularMsg>:<content>
	serverHello //format: <serverHello>:<MyPubKey>
)

var conn net.Conn  //global for this client 

var PubKey string  //my public key to be stored here

var TalkingWith int //current dest. client's pubkey

func main() {
	Init()

	go ParseIncomingMsg()

	//....
	WaitForEver()
}

//
// Common functinos
//


//this will accept user input and send to dest client
//only once server has sent client list and user has chosen dest client
//else it will send msg to client with pubkey stored in TakingWith
//if user types nothing presses enter, it will do nothing 
func StartTyping() {
		for true {
		fmt.Print("Type: ")
		input_data := AcceptInput()

		//Send message only if it is not NULL
		if (input_data != "") {
			//fmt.Printf("\n%s\n", input_data)
			//send message here
			fmt.Fprintf(conn, strconv.Itoa(TalkingWith) + ":" + input_data + "\n")
		}
	}
}



//function that parse incoming server msg based on msgId

func ParseIncomingMsg() {
	//Waiting for server response here
	for true {
		if (!ClientListReceived) {
			fmt.Printf("You are the only one, Waiting for others !!")
		}
		incomingMsg, err := ReadMsg(conn)

		if(err != nil) {
			ErrorLog(ErrorVisible, "disonnected from server , err: %s", err.Error())
			os.Exit(1)		
		}

		msgId, err := getMsgId(incomingMsg)
		if(err != nil) {
			ErrorLog(ErrorVisible, "Failed getting message ID,msg ignored, err: %s", err.Error())
			continue
		}

		msgContent, err := getMsgContent(incomingMsg)
		if(err != nil) {
			ErrorLog(ErrorNoVisible, "err: %s", err.Error())
			os.Exit(1)
		}

		if ((msgId != clientList && msgId != serverHello) && TalkingWith == 0) {
			ErrorLog(ErrorVisible, "Ignoring message yet ")
			continue;
		} 

		switch msgId {
			case clientList:
				if(!ClientListReceived) {

					DisplayClientList(msgContent)
					ClientListReceived = true

					for TalkingWith == 0 {
						fmt.Printf("Enter pubkey of client to talk with : ")
						TmpClientId := AcceptInput();
						TalkingWith, err = strconv.Atoi(TmpClientId)
						if(err != nil)  {
							ErrorLog(ErrorVisible ,"Invalid Client Chosen\n")
						}
						//Start typing messages
						//fmt.Printf("talking with %d\n\n", TalkingWith)
					}
					//keep typing to chosen dest 
					go StartTyping()

				} else {
					//Ignore if already got it .. not needed now 
					ErrorLog(ErrorNoVisible, "Received clientList again, ignored. \n")
				}
				break
			case newClient:
				//FYI
				fmt.Printf("\n\n<<< New Client %s joined !!! >>>\n\n", msgContent)
				break
			case regularMsg: 
				//Received reply, parse and print to user.
				From, FinalMsg, err := GetSenderPubKey(msgContent)
				if (err == nil) {
					fmt.Println("Msg From", From, " >>", FinalMsg)
				} else {
					ErrorLog(ErrorVisible, "Invalid msg from server")
				}
				break
			case serverHello: //Sever hello msg contains public key for this client
				fmt.Printf("INFO: my pubKey is %s", msgContent)
				PubKey = msgContent
				break
			default:
				ErrorLog(ErrorVisible, "Invalid msgId : %d", msgId)
				break
		}
	} //End of for true loop of receive msg
}

//msg in format  of X:msg, parse it 
//return msg value
func getMsgContent(msg string) (string, error) {
	msgData := strings.Split(msg, ":")
	if (len(msgData) >= 1) {
		return msgData[1], nil
	}
	return "", errors.New("Null Msg from server")
}

//display client list from server msg 
//let user chose one from list

func DisplayClientList(msg string) {
	msgList := strings.Split(msg,  "-")
	for i :=0; i< len(msgList); i++ {
		if(msgList[i] != "\n" ) {
			fmt.Printf("'%s' \n", msgList[i])
			//fmt.Println(strconv.Itoa(i), msgList[i])
		}
	}
}

func getMsgId(msg string) (int, error ){

	stringId := strings.Split(msg, ":")

	intVar, err := strconv.Atoi(stringId[0])

	return intVar, err
}

func GetSenderPubKey(msg string) (string, string, error) {
	SenderPublicKey := strings.Split(msg, "-")
	if (len(SenderPublicKey) > 1) {
		return SenderPublicKey[0], SenderPublicKey[1], nil
	}
	return "", "", errors.New("Invalid msg from server")
}

func ReadMsg(conn net.Conn) (string, error ){
	for {
		message, err := bufio.NewReader(conn).ReadString('\n')
		if (err != io.EOF) {
			if (debugOn) { 
				fmt.Print("Raw message Received:", string(message))
			}
			//Add message to global queue with clientId
			return message, nil
		} else {
			fmt.Println("disonnected " + conn.RemoteAddr().String() + "\n" )
			conn.Close()
			return "", errors.New("EOF from server")
		}
	}
}

func Init() {
	var rv int = SkySuccess

	fmt.Println("Welcome to Skycoin end-to-end chat !! \n")

	rv = ConnectToServer()
	for rv != SkySuccess {
		ErrorLog(ErrorVisible, "Retrying");
		time.Sleep(reconnectWaitTime)
		rv = ConnectToServer()
	}
}

func ConnectToServer() int {
	var err error = nil
	conn, err = net.Dial(connType, connHost+":"+connPort)
	if err != nil {
		ErrorLog(ErrorNoVisible,"Error connecting: %s", err.Error())
		return SkyFail
	}
	return SkySuccess;
}

func AcceptInput() string {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text()
}

func ErrorLog(visible int, str string,a ...interface{}) {
	if (true) {
		fmt.Printf(str + "\n", a ...)
	} else {
		//logging to file / etc as needed
	}
	//other error logging if needed
}

func WaitForEver() {
	for true {
		time.Sleep(time.Minute)
	}
}