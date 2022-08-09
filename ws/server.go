package ws

import (
	"context"
	"errors"
	"io"
	"net/http"

	"go.uber.org/zap"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// ServeHTTP is the websocket implementation
func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	zap.L().Sugar().Debug("ws/server: received connect request")
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{})
	if err != nil {
		zap.L().Error("ws/server: error accept request", zap.Error(err))
		return
	}
	defer c.Close(websocket.StatusNormalClosure, "websocket server exit")
	handleRequest(r.Context(), c)
}

func handleRequest(ctx context.Context, c *websocket.Conn) {
	zap.L().Sugar().Debug("ws/server: wating for requests")
	for {
		var request Request
		err := wsjson.Read(ctx, c, &request)
		if err != nil {
			if errors.Is(err, io.EOF) {
				zap.L().Sugar().Debug("ws/server: end request received")
				return
			}

			if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
				zap.L().Info("ws/server: connection closed")
			} else {
				zap.L().Error("ws/server: failed to read json message", zap.Error(err))
			}

			return
		}

		zap.L().Sugar().Debugf("ws/server: received command %s", request.Message)

		response := &Response{
			Request: request,
			Message: "pong",
		}
		err = wsjson.Write(ctx, c, response)
		if err != nil {
			zap.L().Error("ws/server: failed to send the message", zap.Any("response", response), zap.Error(err))
			return
		}
	}
}

// Request represent a websocket request
type Request struct {
	Message string `json:"message"`
}

// Response represents a websocket response
type Response struct {
	Request Request `json:"request"`
	Message string  `json:"message"`
}
