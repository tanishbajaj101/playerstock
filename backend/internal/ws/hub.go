package ws

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/stakestock/backend/internal/pubsub"
)

type subscribeMsg struct {
	Subscribe []string `json:"subscribe"`
}

type Hub struct {
	bus     *pubsub.Bus
	mu      sync.Mutex
	clients map[*client]struct{}
	// channel ref counts: how many clients are subscribed
	channelRefs map[string]int
	// per-channel redis subscription (one per unique channel across all clients)
	redisSubs map[string]context.CancelFunc
}

type client struct {
	conn     *websocket.Conn
	channels map[string]struct{}
	send     chan []byte
}

func NewHub(bus *pubsub.Bus) *Hub {
	return &Hub{
		bus:         bus,
		clients:     make(map[*client]struct{}),
		channelRefs: make(map[string]int),
		redisSubs:   make(map[string]context.CancelFunc),
	}
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // CORS handled by chi middleware
	})
	if err != nil {
		log.Printf("ws accept: %v", err)
		return
	}

	c := &client{
		conn:     conn,
		channels: make(map[string]struct{}),
		send:     make(chan []byte, 256),
	}

	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()

	ctx, cancel := context.WithCancel(r.Context())
	defer func() {
		cancel()
		h.removeClient(c)
		conn.CloseNow()
	}()

	// Writer goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-c.send:
				if !ok {
					return
				}
				if err := conn.Write(ctx, websocket.MessageText, msg); err != nil {
					return
				}
			}
		}
	}()

	// Reader loop: handle subscribe messages
	for {
		var msg subscribeMsg
		if err := wsjson.Read(ctx, conn, &msg); err != nil {
			return
		}
		h.handleSubscribe(c, msg.Subscribe)
	}
}

func (h *Hub) handleSubscribe(c *client, channels []string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, ch := range channels {
		if _, already := c.channels[ch]; already {
			continue
		}
		c.channels[ch] = struct{}{}
		h.channelRefs[ch]++
		if h.channelRefs[ch] == 1 {
			// First subscriber — start Redis listener
			h.startRedisListener(ch)
		}
	}
}

func (h *Hub) startRedisListener(channel string) {
	ctx, cancel := context.WithCancel(context.Background())
	h.redisSubs[channel] = cancel

	go func() {
		sub := h.bus.Subscribe(ctx, channel)
		defer sub.Close()
		ch := sub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				h.broadcast(channel, []byte(msg.Payload))
			}
		}
	}()
}

func (h *Hub) broadcast(channel string, data []byte) {
	// Wrap in envelope: {"channel":"...", "data":{...}}
	envelope, err := json.Marshal(map[string]json.RawMessage{
		"channel": mustMarshalString(channel),
		"data":    data,
	})
	if err != nil {
		return
	}

	h.mu.Lock()
	var targets []*client
	for c := range h.clients {
		if _, ok := c.channels[channel]; ok {
			targets = append(targets, c)
		}
	}
	h.mu.Unlock()

	for _, c := range targets {
		select {
		case c.send <- envelope:
		default:
			// Slow client — drop
		}
	}
}

func (h *Hub) removeClient(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.clients, c)
	for ch := range c.channels {
		h.channelRefs[ch]--
		if h.channelRefs[ch] <= 0 {
			delete(h.channelRefs, ch)
			if cancel, ok := h.redisSubs[ch]; ok {
				cancel()
				delete(h.redisSubs, ch)
			}
		}
	}
}

func mustMarshalString(s string) json.RawMessage {
	b, _ := json.Marshal(s)
	return b
}
