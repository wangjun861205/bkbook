package client

import (
	"../api"
	"google.golang.org/grpc"
)

func NewClient(addr string) (*api.BookClient, error) {
	conn, err := grpc.Dial(addr)
	if err != nil {
		return nil, err
	}
	c := api.NewBookClient(conn)
	return &c, nil
}
