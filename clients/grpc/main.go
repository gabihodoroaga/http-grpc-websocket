package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"os"

	"google.golang.org/api/idtoken"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	grpcinsecure "google.golang.org/grpc/credentials/insecure"
	grpcMetadata "google.golang.org/grpc/metadata"

	pb "github.com/gabihodoroaga/http-grpc-websocket/grpc/proto"
)

var (
	serverAddr = flag.String("server", "localhost:8080", "Server address (host:port)")
	keyPath = flag.String("key", "", "The path to the service account key.json file")
	insecure   = flag.Bool("insecure", true, "No SSL connection")
)

func main() {
	fmt.Printf("starting grpc client\n")

	flag.Parse()

	var opts []grpc.DialOption
	if *serverAddr == "" {
		fmt.Print("-server is empty\n")
		os.Exit(1)
	}
	
	if *insecure {
		opts = append(opts, grpc.WithTransportCredentials(grpcinsecure.NewCredentials()))
	} else {
		cred := credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: true,
		})
		opts = append(opts, grpc.WithTransportCredentials(cred))
	}

	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		fmt.Printf("failed to dial server %s: %v\n", *serverAddr, err)
		os.Exit(1)
	}
	defer conn.Close()

	client := pb.NewEchoClient(conn)

	ctx := context.Background()

	if *keyPath != "" {
		audience := "webapp"
		//tokenSource, err := idtoken.NewTokenSource(ctx, audience)
		tokenSource, err := idtoken.NewTokenSource(ctx, audience, idtoken.WithCredentialsFile(*keyPath))
	
		if err != nil {
			fmt.Printf("idtoken.NewTokenSource: %v", err)
			os.Exit(1)
		}
		token, err := tokenSource.Token()
		if err != nil {
			fmt.Printf("TokenSource.Token: %v", err)
			os.Exit(1)
		}
	
		// Add token to gRPC Request.
		ctx = grpcMetadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token.AccessToken)
	}

	result, err := client.Ping(ctx, &pb.PingRequest{})
	if err != nil {
		fmt.Printf("config rpc failed: %s\n", err.Error())
		os.Exit(1)
	}

	fmt.Printf("result %+v\n", result)
}
