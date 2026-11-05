package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bladekkk/mcap/backend/internal/config"
	"github.com/bladekkk/mcap/backend/internal/database"
	"github.com/bladekkk/mcap/backend/internal/service"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return r.Header.Get("Origin") == "http://localhost:3000"
	},
}

type WebSocketHandler struct {
	db *database.DB
}

func NewWebSocketHandler(db *database.DB) *WebSocketHandler {
	return &WebSocketHandler{db: db}
}

func (h *WebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	_, msg, err := conn.ReadMessage()
	if err != nil {
		fmt.Println("Read error:", err)
		return
	}

	serverID := string(msg)
	if serverID == "" {
		conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"server_id required"}`))
		return
	}

	conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`{"type":"connected","server_id":"%s"}`, serverID)))

	go h.sendMetricsLoop(conn, serverID)
}

func (h *WebSocketHandler) sendMetricsLoop(conn *websocket.Conn, serverID string) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			metrics, err := getServerMetrics(serverID)
			if err != nil {
				conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`{"type":"error","message":"%s"}`, err.Error())))
				return
			}
			conn.WriteMessage(websocket.TextMessage, metrics)
		}
	}
}

func getServerMetrics(serverID string) (string, error) {
	cfg := config.LoadConfig()
	sm := service.NewServerManager(cfg)

	metrics, err := sm.GetMetrics(serverID)
	if err != nil {
		return "", err
	}

	data := map[string]interface{}{
		"type":          "metrics",
		"server_id":     serverID,
		"timestamp":     time.Now().Format(time.RFC3339),
		"tps":           metrics.TPS,
		"mspt":          metrics.MSPT,
		"ram_mb":        metrics.RSSMB,
		"cpu_percent":   metrics.CPUPercent,
		"online_players": metrics.OnlinePlayers,
	}

	b, _ := json.Marshal(data)
	return string(b), nil
}
