package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

	"google.golang.org/api/idtoken"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

var (
	serverAddr = flag.String("server", "ws://localhost:8080/ws", "Server address (ws://host:port/ws)")
	keyPath = flag.String("key", "", "The path to the service account key.json file")
)

func main() {

	flag.Parse()
	ctx := context.Background()

	opts := &websocket.DialOptions{}

	if *keyPath != "" {
		audience := "webapp"
		tokenSource, err := idtoken.NewTokenSource(ctx, audience, idtoken.WithCredentialsFile(*keyPath))
	
		if err != nil {
			fmt.Printf("idtoken.NewTokenSource: %v\n", err)
			os.Exit(1)
		}
	
		token, err := tokenSource.Token()
		if err != nil {
			fmt.Printf("TokenSource.Token: %v\n", err)
			os.Exit(1)
		}

		opts.HTTPHeader = http.Header{"authorization": {fmt.Sprintf("Bearer %s", token.AccessToken)}}
	}

	c, _, err := websocket.Dial(ctx, *serverAddr, opts)

	if err != nil {
		fmt.Printf("error %s\n", err)
		os.Exit(1)
	}
	defer c.Close(websocket.StatusNormalClosure, "")

	request := request{Message: "ping"}
	fmt.Printf("request = %v\n", request)
	err = wsjson.Write(ctx, c, request)
	if err != nil {
		fmt.Printf("error %s\n", err)
		os.Exit(1)
	}

	result := map[string]interface{}{}
	err = wsjson.Read(ctx, c, &result)
	if err != nil {
		fmt.Printf("error %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("response = %v\n", result)
}

type request struct {
	Message string `json:"message"`
}
