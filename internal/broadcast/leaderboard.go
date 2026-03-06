// Package broadcast manages WebSocket connections for real-time leaderboard updates.
// It is a dependency-free package that both handlers and services can import.
package broadcast

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	batchClients   = map[string]map[*websocket.Conn]bool{}
	batchClientsMu sync.Mutex
)

// Register adds a WebSocket connection to the broadcast group for a batch.
func Register(batchCode string, conn *websocket.Conn) {
	batchClientsMu.Lock()
	defer batchClientsMu.Unlock()
	if batchClients[batchCode] == nil {
		batchClients[batchCode] = map[*websocket.Conn]bool{}
	}
	batchClients[batchCode][conn] = true
}

// Unregister removes a WebSocket connection from the batch group.
func Unregister(batchCode string, conn *websocket.Conn) {
	batchClientsMu.Lock()
	defer batchClientsMu.Unlock()
	if conns, ok := batchClients[batchCode]; ok {
		delete(conns, conn)
	}
}

// LeaderboardPayload is what gets broadcast to all clients.
type LeaderboardPayload struct {
	Type      string        `json:"type"`
	BatchCode string        `json:"batchCode"`
	Entries   []interface{} `json:"entries"`
	UpdatedAt time.Time     `json:"updatedAt"`
}

// Broadcast sends a leaderboard message to all WebSocket clients in a batch.
func Broadcast(batchCode string, entries []interface{}) {
	batchClientsMu.Lock()
	conns, ok := batchClients[batchCode]
	batchClientsMu.Unlock()

	if !ok || len(conns) == 0 {
		return
	}

	msg, _ := json.Marshal(LeaderboardPayload{
		Type:      "leaderboard",
		BatchCode: batchCode,
		Entries:   entries,
		UpdatedAt: time.Now().UTC(),
	})

	batchClientsMu.Lock()
	defer batchClientsMu.Unlock()
	for conn := range conns {
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Printf("[broadcast] write error for %s: %v", batchCode, err)
			conn.Close()
			delete(conns, conn)
		}
	}
}
