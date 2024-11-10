# Hand-in-4, Group 404

The program supports one host node ("commander" in the consensus context) and at least 3 standard nodes.

Please initiate the host node first, followed by as many nodes as you wish.


To initialize the host:

  Navigate to the client directory, then run "go run client.go host". If a firewall prompt appears, press "Allow access"


To initialize subsequent nodes:

  Navigate to the client directory, then run "go run client.go". If a firewall prompt appears, press "Allow access"

An empty log file "mylog.log" is included in the project. The node(s) know to connect to it, 
so once you run the scripts, you can see occuring events registered in the log!
