package websocket

import (
	"context"
	"encoding/json"
	"github.com/HydroProtocol/hydro-sdk-backend/utils"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type ClientRequest struct {
	Type     string
	Channels []string
}

func handleClientRequest(client *Client) {
	utils.Info("New Client(%s) IP:(%s) Connect", client.ID, client.Conn.RemoteAddr())

	defer utils.Info("Client(%s) IP:(%s) Disconnect", client.ID, client.Conn.RemoteAddr())

	for {
		var req ClientRequest

		err := client.Conn.ReadJSON(&req)

		switch err.(type) {
		case *json.SyntaxError:
			utils.Error("request must be json")
			continue
		case *websocket.CloseError:
			return
		}

		utils.Debug("Recv c(%s): %+v", client.ID, req)

		switch req.Type {
		case "subscribe":
			for _, id := range req.Channels {
				channel := FindChannel(id)

				if channel == nil {
					// There is a risk to let user create channel freely.
					channel = CreateChannelByID(id)
				}

				if channel != nil {
					channel.AddSubscriber(client)
				}
			}
		case "unsubscribe":
			for _, id := range req.Channels {
				channel := FindChannel(id)

				if channel == nil {
					continue
				}

				channel.RemoveSubscriber(client.ID)
			}
		}
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func connectHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Print("upgrade error:", err)
		return
	}

	defer c.Close()

	client := NewClient()
	client.Conn = c

	handleClientRequest(client)
}

func StartSocketServer(ctx context.Context) {
	srv := &http.Server{Addr: ":3002"}

	http.HandleFunc("/", connectHandler)

	go func() {
		// returns ErrServerClosed on graceful close
		utils.Info("Websocket Server is listening on 0.0.0.0:3002")
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Serve Exit Error: %s", err)
		}
	}()

	<-ctx.Done()

	// now close the server gracefully ("shutdown")
	if err := srv.Shutdown(context.Background()); err != nil {
		panic(err)
	}
}
