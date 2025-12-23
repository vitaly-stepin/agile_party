package websocket

import (
	"log"
	"time"

	"github.com/fasthttp/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 8192
)

type Client struct {
	conn    *websocket.Conn
	hub     *WsHub // Hub managing this client
	RoomID  string
	UserID  string
	send    chan []byte
	handler MessageHandler
}

type MessageHandler interface {
	HandleMessage(client *Client, message WsMessage) error
}

func NewClient(conn *websocket.Conn, hub *WsHub, roomID, userID string, handler MessageHandler) *Client {
	return &Client{
		conn:    conn,
		hub:     hub,
		RoomID:  roomID,
		UserID:  userID,
		send:    make(chan []byte, 256),
		handler: handler,
	}
}

func (c *Client) readPump() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in readPump for user %s in room %s: %v", c.UserID, c.RoomID, r)
		}
		c.hub.unregister <- c
		if c.conn != nil {
			c.conn.Close()
		}
	}()

	if c.conn == nil {
		log.Printf("readPump: connection is nil for user %s in room %s", c.UserID, c.RoomID)
		return
	}

	err := c.conn.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		log.Printf("readPump: failed to set read deadline for user %s in room %s: %v", c.UserID, c.RoomID, err)
		return
	}

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetPongHandler(func(string) error {
		if c.conn != nil {
			c.conn.SetReadDeadline(time.Now().Add(pongWait))
		}
		return nil
	})

	for {
		var msg WsMessage
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error for user %s in room %s: %v", c.UserID, c.RoomID, err)
			}
			break
		}

		if err := c.handler.HandleMessage(c, msg); err != nil {
			log.Printf("Error handling message from user %s in room %s: %v", c.UserID, c.RoomID, err)

			c.SendError(err.Error(), "HANDLER_ERROR")
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in writePump for user %s in room %s: %v", c.UserID, c.RoomID, r)
		}
		ticker.Stop()
		if c.conn != nil {
			c.conn.Close()
		}
	}()

	if c.conn == nil {
		log.Printf("writePump: connection is nil for user %s in room %s", c.UserID, c.RoomID)
		return
	}

	for {
		select {
		case message, ok := <-c.send:
			if c.conn == nil {
				return
			}
			// Set write timeout
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			if c.conn == nil {
				return
			}
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// This blocks until the connection is closed, so it should be called
// as the last operation in the connection handler
func (c *Client) Start() {
	go c.writePump()
	c.readPump()
}

func (c *Client) SendError(message, code string) {
	c.hub.BroadcastToRoom(c.RoomID, WsMessage{
		Type: EventTypeError,
		Payload: ErrorPayload{
			Message: message,
			Code:    code,
		},
	}, nil)
}
