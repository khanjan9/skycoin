# skycoin

Total timetaken:
6 days (6may - 12may) coding time 2 days. 
-this includes 
-learning about Go syntax and features,
-initial design of chat applications 
-learning goroutin and channels in details

Client process:
-on coming up, finds server , if not online tries reconeceting every 2 seconds 
-once connected, server sends  "ServerHello" and "ClietList" message
-ServerHello is first message and contains PublicKey for client
-ClientList contains all clients connected to server, client can chose 1 to chat


Server Process:
-default listening on port localhost: 8082, clients connect to same
-on receiving client connection, server sends pubkey with 'ServerHello'
-if it is only one client in list, no 'ServerList' is sent untill second client joins (client waits for 'ClientList' from server)
-Server receving message has dest client id in starting of msg and it is forwarded accordingly

Message format:
String: "<msgId>:<msg content>"
eg:
  0:6-7-9-34   ->msgid: 0 (ClientList)  and 6,7,9,34 are 4 clients with pubkey
  1:5          -> msgid : 1 (newClient) and 5 is pubkey of that client
  2:3-hello!! -> msgid: 2 (regularMsg), sourceClient: 3 and msg content "Hello"
  3:23        -> msgid: 3 (ServerHello), 23 is client's pubKey
  
