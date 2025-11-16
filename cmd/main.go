package main

import serv "pull-request-reviewers-service/server"

func main() {
	server := serv.Server{}
	server.Init()
	server.Start()
}
