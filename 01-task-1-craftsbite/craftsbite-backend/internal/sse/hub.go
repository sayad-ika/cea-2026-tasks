package sse

import "sync"

// Client represents one connected browser tab (admin/logistics user)
type Client struct {
	Channel chan []byte
}

// Hub manages all connected SSE clients and fans out broadcasts
type Hub struct {
	mu         sync.RWMutex
	clients    map[*Client]struct{}
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan []byte
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]struct{}),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan []byte, 256), // buffered so senders never block
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.clients[client] = struct{}{}
			h.mu.Unlock()

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Channel)
			}
			h.mu.Unlock()

		case msg := <-h.Broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.Channel <- msg:
				default:
					// Slow client — skip rather than block the hub
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastHeadcount serialises the payload and sends it to all connected clients
func (h *Hub) BroadcastHeadcount(payload any) {
	// Marshalling is done in the handler layer — see headcount_sse.go
	// This just fans out already-serialised bytes
	select {
	case h.Broadcast <- payload.([]byte):
	default:
		// Broadcast channel full — skip cycle, next update will catch up
	}
}
