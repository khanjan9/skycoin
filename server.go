/*
*  @file : server.go
*  @brief: File contains server code, 
*
*/
package main

import (
	"fmt"
	"bufio"
	"os"
	"time"
	"net"
	"io"
	"strconv"
	"strings"
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

var msg chan [2]string //channel to pass incoming message to main goroutine

var clients map[net.Conn]int //map file contains net.Conn mapping with pubkey of client 

var PublicKey int = 1 //Start giving PibKey to clients 1 onwards


func main() {

	//inits
	clients = make(map[net.Conn]int)
	msg = make(chan [2]string)


	//start listening..
	l, err := net.Listen(connType, connHost+":"+connPort)
	if (err != nil) {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer l.Close()
	fmt.Println("Sever is online and listening\n")
	

	//keep accepting client connections
	go AcceptConnection(l)


	//when any client sends message, it is sent to main go routine using msg channel
	go ReadFromChan(msg)


	//waiting forever  ):
	WaitForEver()
}


//reads message from channel and parsing it and sending it to dest 
//msg formate: "x:msg", x is int, dest client pubkey
//msg is content , 
//eg: "3:hello" for net.conn. to be sent to client with pubKey 3.
func ReadFromChan(msg_chn chan [2]string) {
	for true {
		msg := <- msg_chn
		fmt.Println("Raw Received : " + msg[0] + "from client " + msg[1])
		id, MsgToSend, err := getMsgContent(msg[0])

		if(err != nil) {
			fmt.Printf("Invlid msg received%s", err.Error())
			//ignoring this msg to avoid crash 
			continue
		}
		conn := findByValue(clients, id)
		if (conn == nil) {
			fmt.Println("Bad client detected ... ")
			//closing for safe side
			delete(clients, conn)
			conn.Close()
			continue
		}
		fmt.Fprintf(conn, strconv.Itoa(regularMsg) + ":" + msg[1] + "-" + MsgToSend  + "\n")
	}
}


//find in map table by value and return int pubkey .
// eg find net.Conn for client pubkey X in argument pubkey
func findByValue(clients map[net.Conn]int, PubKey int) (net.Conn) {
	for conn, pubKeyLocal := range clients {
		if(pubKeyLocal == PubKey) {			
			return conn
		}
	}
	return nil
}

//converts client message "X:msg" to int (x), string(msg)

//eg string "1:Hello" , to be parsed as below and return:
//int - 1
//string- hello
//error - if invalid string 
func getMsgContent(msg string) (int, string,error ){

	stringId := strings.Split(msg, ":")

	intVar, err := strconv.Atoi(stringId[0])

	return intVar, stringId[1], err
}


/*
On accepting client connection:
1. give pubKey to client with msg id ServerHello (id:3)
2. notify all connected clients that there is new client
3. send client list to all clients, clients which has list, will ignore, other will chose
4. start goroutine to accept incoming message from this net.conn
*/
func AcceptConnection(l net.Listener) {
	for true {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error connecting:", err.Error())
			continue
		}
		fmt.Println("Connected " + conn.RemoteAddr().String() + "\n" )

		//notify other clients about new client joined
		for Con, _ := range clients {
				fmt.Fprintf(Con, strconv.Itoa(newClient) + ":" +  strconv.Itoa(PublicKey+1) + "\n")			
		}

		//When client connects, send reply with its pubKey
		//create pubkey for that client
		PublicKey++
		clients[conn] = PublicKey
		fmt.Fprintf(conn, strconv.Itoa(serverHello) + ":" + strconv.Itoa(PublicKey) + "\n")

		//send client list to all clients
		if( len(clients) > 1) {
			//create cliet list string
			//formate : <clientListMsgId>:client1-client2-client3-
			//eg: 0:2-3-4-6-  => msgId: 0 (list) and client entries are 2,3,4,6
			ClitListString := strconv.Itoa(0) + ":"
			for _, Pkey := range clients {
				ClitListString += strconv.Itoa(Pkey) + "-"
			}
			fmt.Println("Sending to all", ClitListString)

			//send client list to all connected clients 
			//client who has already received it , will igonre it
			for Con,_ := range clients {
				fmt.Fprintf(Con,ClitListString + "\n")
			}

		}

		//keep reading message from this client
		go ReadMsg(conn)
	}
}


// Keep reading messages from perticular net.Conn
// for each client one instance of client
// send received message to channel to parse and send to dest. 
// if msg is EOF, delete client from "clients" map

//channel is [2]strings
//string[0] is msg from client
//string[1] is client's net.conn value

func ReadMsg(conn net.Conn) {
	for {
		message, err := bufio.NewReader(conn).ReadString('\n')
		if (err != io.EOF) {
			fmt.Println("Raw Message:", string(message))
			//Add message to global queue with clientId
			var chan_msg [2]string
			chan_msg[0] = message
			chan_msg[1] = strconv.Itoa(clients[conn]) 
			msg <- chan_msg
		} else {
			fmt.Println("Warn: disonnected " + conn.RemoteAddr().String() + "\n" )
			//delete client from our global db and close conn
			delete(clients, conn)
			conn.Close()
			return
		}
	}
}


//not letting main thread exit
func WaitForEver() {
	for true {
		time.Sleep(time.Minute)
	}
}