package ws

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"testing"
	"time"

	hclient "github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/network/standard"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/hertz-contrib/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aeroxe/sign-flow/backend/internal/events"
	"github.com/aeroxe/sign-flow/backend/internal/registry"
)

// startTestServer boots a real Hertz server with the /ws route on an
// ephemeral port and returns its address.
func startTestServer(t *testing.T, hub *Hub) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := ln.Addr().(*net.TCPAddr).Port
	require.NoError(t, ln.Close())

	h := server.New(server.WithHostPorts(fmt.Sprintf("127.0.0.1:%d", port)))
	h.NoHijackConnPool = true
	Register(registry.New(), h.Group("/"), hub)
	go h.Spin()
	t.Cleanup(func() { _ = h.Shutdown(context.Background()) })
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	// Spin is async: wait until the listener actually accepts connections
	// before handing the address to the client (avoids flaky 'connection refused').
	require.Eventually(t, func() bool {
		conn, err := net.DialTimeout("tcp", addr, 200*time.Millisecond)
		if err != nil {
			return false
		}
		_ = conn.Close()
		return true
	}, 5*time.Second, 20*time.Millisecond)
	return addr
}

// dialWS connects a websocket client via the hertz client + ClientUpgrader.
func dialWS(t *testing.T, addr, origin string) *websocket.Conn {
	t.Helper()
	c, err := hclient.NewClient(hclient.WithDialer(standard.NewDialer()))
	require.NoError(t, err)
	req, resp := protocol.AcquireRequest(), protocol.AcquireResponse()
	req.SetRequestURI("http://" + addr + "/ws")
	req.SetMethod("GET")
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	u := &websocket.ClientUpgrader{}
	u.PrepareRequest(req)
	err = c.Do(context.Background(), req, resp)
	require.NoError(t, err)
	conn, err := u.UpgradeResponse(req, resp)
	require.NoError(t, err)
	return conn
}

func TestHubBroadcastsToAllClients(t *testing.T) {
	bus := events.NewBus()
	hub := NewHub(bus, 10, OriginPolicy{})
	addr := startTestServer(t, hub)

	conns := []*websocket.Conn{dialWS(t, addr, ""), dialWS(t, addr, "")}
	defer func() {
		for _, c := range conns {
			_ = c.Close()
		}
	}()
	require.Eventually(t, func() bool { return hub.Count() == 2 }, 2*time.Second, 20*time.Millisecond)

	bus.Publish(events.Event{EventType: "contract_created", Data: []byte(`{"contract_id":"c1"}`)})

	for i, c := range conns {
		_ = c.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, msg, err := c.ReadMessage()
		require.NoErrorf(t, err, "client %d did not receive broadcast", i)
		assert.Contains(t, string(msg), "contract_created")
	}
}

func TestHubRejectsUnknownOriginInProduction(t *testing.T) {
	bus := events.NewBus()
	hub := NewHub(bus, 10, OriginPolicy{Allowed: []string{"https://app.example.com"}, Enforce: true})
	addr := startTestServer(t, hub)

	c, err := hclient.NewClient(hclient.WithDialer(standard.NewDialer()))
	require.NoError(t, err)
	req, resp := protocol.AcquireRequest(), protocol.AcquireResponse()
	req.SetRequestURI("http://" + addr + "/ws")
	req.SetMethod("GET")
	req.Header.Set("Origin", "https://evil.example.com")
	u := &websocket.ClientUpgrader{}
	u.PrepareRequest(req)
	require.NoError(t, c.Do(context.Background(), req, resp))
	_, err = u.UpgradeResponse(req, resp)
	require.Error(t, err, "dial with unknown origin must be rejected")
}

func TestHubAllowsListedOriginInProduction(t *testing.T) {
	bus := events.NewBus()
	hub := NewHub(bus, 10, OriginPolicy{Allowed: []string{"https://app.example.com"}, Enforce: true})
	addr := startTestServer(t, hub)

	conn := dialWS(t, addr, "https://app.example.com")
	defer conn.Close()
	require.Eventually(t, func() bool { return hub.Count() == 1 }, 2*time.Second, 20*time.Millisecond)
}

func TestHubDropsSlowConsumer(t *testing.T) {
	bus := events.NewBus()
	hub := NewHub(bus, 10, OriginPolicy{})
	addr := startTestServer(t, hub)

	// Connect but never read: its 32-message buffer fills and it gets dropped.
	conn := dialWS(t, addr, "")
	defer conn.Close()
	require.Eventually(t, func() bool { return hub.Count() == 1 }, 2*time.Second, 20*time.Millisecond)

	for i := 0; i < 200; i++ {
		bus.Publish(events.Event{EventType: "burst", Data: []byte(`{"i":` + fmt.Sprint(i) + `}`)})
	}

	require.Eventually(t, func() bool { return hub.Count() == 0 }, 5*time.Second, 20*time.Millisecond)
}

func TestHubCloseDropsConnections(t *testing.T) {
	bus := events.NewBus()
	hub := NewHub(bus, 10, OriginPolicy{})
	addr := startTestServer(t, hub)

	conn := dialWS(t, addr, "")
	defer conn.Close()
	require.Eventually(t, func() bool { return hub.Count() == 1 }, 2*time.Second, 20*time.Millisecond)

	hub.Close()
	require.Eventually(t, func() bool { return hub.Count() == 0 }, 2*time.Second, 20*time.Millisecond)

	// After Close, new connections are rejected (429) rather than accepted.
	u, _ := url.Parse("http://" + addr + "/ws")
	c, err := hclient.NewClient(hclient.WithDialer(standard.NewDialer()))
	require.NoError(t, err)
	req, resp := protocol.AcquireRequest(), protocol.AcquireResponse()
	req.SetRequestURI(u.String())
	req.SetMethod("GET")
	up := &websocket.ClientUpgrader{}
	up.PrepareRequest(req)
	require.NoError(t, c.Do(context.Background(), req, resp))
	_, err = up.UpgradeResponse(req, resp)
	require.Error(t, err, "upgrade after Close must fail")
}

func TestOriginPolicyCheck(t *testing.T) {
	prod := OriginPolicy{Allowed: []string{"https://a.example.com"}, Enforce: true}
	assert.True(t, prod.check("https://a.example.com"))
	assert.False(t, prod.check("https://b.example.com"))
	assert.True(t, prod.check(""), "non-browser clients (no Origin) pass")

	dev := OriginPolicy{}
	assert.True(t, dev.check("https://anything.example.com"), "dev allows any origin")
}
