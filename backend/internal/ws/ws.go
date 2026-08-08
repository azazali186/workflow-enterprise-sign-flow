// Package ws implements the real-time WebSocket hub broadcasting domain
// events (contract_created, signature_requested, signed_update,
// contract_executed) to connected clients.
package ws

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"
	"github.com/hertz-contrib/websocket"
	"go.uber.org/zap"

	"github.com/aeroxe/sign-flow/backend/internal/events"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/logger"
	"github.com/aeroxe/sign-flow/backend/internal/registry"
)

const writeWait = 10 * time.Second

// Hub tracks connections and fans events out.
type Hub struct {
	bus  *events.Bus
	max  int
	mu   sync.Mutex
	conns map[*websocket.Conn]bool
}

// NewHub builds a hub subscribing to the event bus.
func NewHub(bus *events.Bus, max int) *Hub {
	h := &Hub{bus: bus, max: max, conns: map[*websocket.Conn]bool{}}
	if max <= 0 {
		max = 1000
	}
	h.max = max
	bus.Subscribe(h)
	return h
}

// Count returns the number of live connections.
func (h *Hub) Count() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.conns)
}

// OnEvent implements events.Subscriber.
func (h *Hub) OnEvent(ev events.Event) {
	payload, err := json.Marshal(ev)
	if err != nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	for conn := range h.conns {
		_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			_ = conn.Close()
			delete(h.conns, conn)
		}
	}
}

var upgrader = websocket.HertzUpgrader{
	CheckOrigin: func(_ *app.RequestContext) bool { return true }, // CORS: allow any origin for realtime clients
}

// Serve upgrades the connection and registers it with the hub.
func (h *Hub) Serve(_ context.Context, c *app.RequestContext) {
	h.mu.Lock()
	if len(h.conns) >= h.max {
		h.mu.Unlock()
		c.AbortWithStatus(429)
		return
	}
	h.mu.Unlock()
	err := upgrader.Upgrade(c, func(conn *websocket.Conn) {
		h.mu.Lock()
		h.conns[conn] = true
		h.mu.Unlock()
		logger.L().Info("websocket client connected", zap.Int("connections", h.Count()))
		h.readPump(conn)
	})
	if err != nil {
		logger.L().Debug("websocket upgrade failed", zap.Error(err))
	}
}

// readPump drains reads to detect disconnects.
func (h *Hub) readPump(conn *websocket.Conn) {
	defer func() {
		h.mu.Lock()
		delete(h.conns, conn)
		h.mu.Unlock()
		_ = conn.Close()
	}()
	conn.SetReadLimit(4096)
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}

// Register adds the /ws route (public, excluded from permission seeding).
func Register(reg *registry.Registry, g *route.RouterGroup, hub *Hub) {
	reg.Register("GET", "/ws", "WebSocket Connection", "PUBLIC")
	g.GET("/ws", hub.Serve)
}
