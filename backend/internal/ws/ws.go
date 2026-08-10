// Package ws implements the real-time WebSocket hub broadcasting domain
// events (contract_created, signature_requested, signed_update,
// contract_executed) to connected clients.
package ws

import (
	"context"
	"encoding/json"
	"slices"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"
	"github.com/hertz-contrib/websocket"
	"go.uber.org/zap"

	"github.com/aeroxe/sign-flow/backend/internal/events"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/logger"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/safego"
	"github.com/aeroxe/sign-flow/backend/internal/registry"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 54 * time.Second // must be < pongWait
	maxMessageSize = 4096
	sendBuffer     = 32
)

// OriginPolicy controls which origins may open a WebSocket. Enforce=true
// (production) rejects anything not in Allowed; "" origin (non-browser) passes.
type OriginPolicy struct {
	Allowed []string
	Enforce bool
}

func (p OriginPolicy) check(origin string) bool {
	if !p.Enforce || origin == "" {
		return true
	}
	return slices.Contains(p.Allowed, origin)
}

// client is one live connection with a buffered, per-connection outbox so a
// slow consumer never blocks broadcasts to other clients.
type client struct {
	conn *websocket.Conn
	send chan []byte
}

// Hub tracks connections and fans events out.
type Hub struct {
	bus     *events.Bus
	max     int
	origin  OriginPolicy
	mu      sync.Mutex
	clients map[*client]bool
	closed  bool
}

// NewHub builds a hub subscribing to the event bus.
func NewHub(bus *events.Bus, max int, origin OriginPolicy) *Hub {
	if max <= 0 {
		max = 1000
	}
	h := &Hub{bus: bus, max: max, origin: origin, clients: map[*client]bool{}}
	bus.Subscribe(h)
	return h
}

// Count returns the number of live connections.
func (h *Hub) Count() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.clients)
}

// Close stops accepting new connections, closes all live ones and removes them
// immediately (graceful shutdown). The per-connection pumps exit on the next
// I/O error; closing the underlying socket interrupts a blocked ReadMessage.
func (h *Hub) Close() {
	h.mu.Lock()
	h.closed = true
	for cl := range h.clients {
		if uc := cl.conn.UnderlyingConn(); uc != nil {
			_ = uc.Close()
		}
		_ = cl.conn.Close()
		delete(h.clients, cl)
	}
	h.mu.Unlock()
}

// OnEvent implements events.Subscriber: enqueue to every client's outbox,
// dropping clients whose buffer is full (too slow to keep up).
func (h *Hub) OnEvent(ev events.Event) {
	payload, err := json.Marshal(ev)
	if err != nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	for cl := range h.clients {
		select {
		case cl.send <- payload:
		default: // slow consumer: drop and close
			_ = cl.conn.Close()
			delete(h.clients, cl)
			logger.L().Warn("websocket client dropped (slow consumer)")
		}
	}
}

// Serve upgrades the connection and registers it with the hub.
func (h *Hub) Serve(_ context.Context, c *app.RequestContext) {
	origin := string(c.Request.Header.Peek("Origin"))
	if !h.origin.check(origin) {
		c.AbortWithStatus(403)
		return
	}
	upgrader := websocket.HertzUpgrader{
		CheckOrigin: func(_ *app.RequestContext) bool { return h.origin.check(origin) },
	}
	h.mu.Lock()
	if h.closed || len(h.clients) >= h.max {
		h.mu.Unlock()
		c.AbortWithStatus(429)
		return
	}
	h.mu.Unlock()
	err := upgrader.Upgrade(c, func(conn *websocket.Conn) {
		cl := &client{conn: conn, send: make(chan []byte, sendBuffer)}
		h.mu.Lock()
		if h.closed {
			h.mu.Unlock()
			_ = conn.Close() // hub already shut down: drop this upgrade
			return
		}
		h.clients[cl] = true
		h.mu.Unlock()
		logger.L().Info("websocket client connected", zap.Int("connections", h.Count()))
		safego.Go(func() { h.writePump(cl) })
		h.readPump(cl) // blocks until disconnect; cleans up below
		h.mu.Lock()
		delete(h.clients, cl)
		h.mu.Unlock()
	})
	if err != nil {
		logger.L().Debug("websocket upgrade failed", zap.Error(err))
	}
}

// writePump sends queued messages and periodic pings, enforcing deadlines.
func (h *Hub) writePump(cl *client) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = cl.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-cl.send:
			_ = cl.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = cl.conn.WriteMessage(websocket.CloseMessage, nil)
				return
			}
			if err := cl.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = cl.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := cl.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// readPump drains reads, detects dead peers via pong timeout, and limits size.
func (h *Hub) readPump(cl *client) {
	defer func() {
		close(cl.send)
		_ = cl.conn.Close()
	}()
	cl.conn.SetReadLimit(maxMessageSize)
	_ = cl.conn.SetReadDeadline(time.Now().Add(pongWait))
	cl.conn.SetPongHandler(func(string) error {
		return cl.conn.SetReadDeadline(time.Now().Add(pongWait))
	})
	for {
		if _, _, err := cl.conn.ReadMessage(); err != nil {
			return
		}
	}
}

// Register adds the /ws route (public, excluded from permission seeding).
func Register(reg *registry.Registry, g *route.RouterGroup, hub *Hub) {
	reg.Register("GET", "/ws", "WebSocket Connection", "PUBLIC")
	g.GET("/ws", hub.Serve)
}
