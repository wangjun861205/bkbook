package main

import (
	"flag"
	"log"
	"net"

	"google.golang.org/grpc"

	"../api"
)

func main() {
	var addr string
	flag.StringVar(&addr, "a", "127.0.0.1:9999", "listen address")
	flag.Parse()
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	s, err := api.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}
	grpcServer := grpc.NewServer()
	api.RegisterBookServer(grpcServer, s)
	log.Fatal(grpcServer.Serve(l))
}
