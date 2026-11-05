package service

import (
	"context"
	"fmt"
	"time"

	"github.com/bladekkk/mcap/backend/internal/database"
)

type MetricsCollector struct {
	db       *database.DB
	serverID string
	interval time.Duration
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewMetricsCollector(db *database.DB, serverID string, interval time.Duration) *MetricsCollector {
	ctx, cancel := context.WithCancel(context.Background())
	return &MetricsCollector{
		db:       db,
		serverID: serverID,
		interval: interval,
		ctx:      ctx,
		cancel:   cancel,
	}
}

func (mc *MetricsCollector) Start(sm *ServerManager) {
	ticker := time.NewTicker(mc.interval)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-mc.ctx.Done():
				return
			case <-ticker.C:
				metrics, err := sm.GetMetrics(mc.serverID)
				if err != nil {
					continue
				}
				mc.saveMetrics(metrics)
			}
		}
	}()
}

func (mc *MetricsCollector) Stop() {
	mc.cancel()
}

func (mc *MetricsCollector) saveMetrics(m *ServerMetrics) error {
	_, err := mc.db.Exec(`
		INSERT INTO metrics_history (server_id, timestamp, tps, mspt, online_players, ram_mb, cpu_percent)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		mc.serverID, time.Now(), m.TPS, m.MSPT, m.OnlinePlayers, m.RSSMB, m.CPUPercent)
	return err
}
