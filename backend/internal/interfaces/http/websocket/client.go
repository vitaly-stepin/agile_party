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

// Client represents a WebSocket client connection
type Client struct {
	// WebSocket connection
	conn *websocket.Conn

	// Hub managing this client
	hub *Hub

	// Room ID this client belongs to
	RoomID string

	// User ID of this client
	UserID string

	// Buffered channel of outbound messages
	send chan []byte

	// Message handler
	handler MessageHandler
}

// MessageHandler processes incoming WebSocket messages
type MessageHandler interface {
	HandleMessage(client *Client, message Message) error
}

// NewClient creates a new WebSocket client
func NewClient(conn *websocket.Conn, hub *Hub, roomID, userID string, handler MessageHandler) *Client {
	return &Client{
		conn:    conn,
		hub:     hub,
		RoomID:  roomID,
		UserID:  userID,
		send:    make(chan []byte, 256),
		handler: handler,
	}
}

// readPump pumps messages from the WebSocket connection to the hub
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
		var msg Message
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error for user %s in room %s: %v", c.UserID, c.RoomID, err)
			}
			break
		}

		// Handle the message
		if err := c.handler.HandleMessage(c, msg); err != nil {
			log.Printf("Error handling message from user %s in room %s: %v", c.UserID, c.RoomID, err)

			// Send error back to client
			c.SendError(err.Error(), "HANDLER_ERROR")
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection
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

			// Add queued messages to the current WebSocket message
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

// Start begins reading and writing for this client
// This blocks until the connection is closed, so it should be called
// as the last operation in the connection handler
func (c *Client) Start() {
	go c.writePump()
	c.readPump() // Block on read pump - when it returns, connection is closed
}

// SendError sends an error message to the client
func (c *Client) SendError(message, code string) {
	c.hub.BroadcastToRoom(c.RoomID, Message{
		Type: EventTypeError,
		Payload: ErrorPayload{
			Message: message,
			Code:    code,
		},
	}, nil)
}
