package main

import (
	"fmt"
	"net"
	"os"

	"github.com/go-faster/errors"
	v1 "github.com/kaynelza/NTF/internal/presentation/grpc/v1"
	pb "github.com/kaynelza/NTF/pkg/grpc/notifyd/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	if err := start(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func start() error {
	var server *v1.NotificationServiceServer

	if err := run(server); err != nil {
		return errors.Wrap(err, "run")
	}

	return nil
}

func run(server *v1.NotificationServiceServer) error {
	list, err := net.Listen("tcp", ":8080")
	if err != nil {
		return err
	}

	srv := grpc.NewServer()
	pb.RegisterNotificationServiceServer(srv, server)
	reflection.Register(srv)

	return srv.Serve(list)
}
