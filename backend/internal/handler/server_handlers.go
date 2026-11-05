package handler

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bladekkk/mcap/backend/internal/config"
	"github.com/bladekkk/mcap/backend/internal/database"
	"github.com/bladekkk/mcap/backend/internal/service"
	"github.com/gofiber/fiber/v2"
)

type CreateServerRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=50,alphanum"`
	JarPath     string `json:"jar_path" validate:"required"`
	JavaVersion string `json:"java_version"`
	XmsMB       int    `json:"xms_mb" validate:"min=256"`
	XmxMB       int    `json:"xmx_mb" validate:"min=256"`
	Port        int    `json:"port" validate:"min=1,max=65535"`
	AutoStart   bool   `json:"autostart"`
	EULA        bool   `json:"eula_accepted"`
}

type ServerInfo struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	JarPath       string `json:"jar_path"`
	Status        string `json:"status"`
	Port          int    `json:"port"`
	JavaVersion   string `json:"java_version"`
	XmsMB         int    `json:"xms_mb"`
	XmxMB         int    `json:"xmx_mb"`
	AutoStart     bool   `json:"autostart"`
	EULAAccepted  bool   `json:"eula_accepted"`
	CreatedAt     string `json:"created_at"`
	OnlinePlayers int    `json:"online_players,omitempty"`
	TPS           float64 `json:"tps,omitempty"`
	RAMMB         int    `json:"ram_mb,omitempty"`
	CPUPercent    float64 `json:"cpu_percent,omitempty"`
}

func CreateServer(db *database.DB, cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req CreateServerRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
		}

		if req.XmsMB > req.XmxMB {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Xms must be <= Xmx"})
		}

		if req.Name == "" || !isValidServerName(req.Name) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid server name"})
		}

		serverID := generateID()
		serverDir := config.GetServerDataDir(cfg.DataDir, serverID)
		if err := os.MkdirAll(serverDir, 0755); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create server dir"})
		}

		_, err := db.Exec(`INSERT INTO servers (id, name, jar_path, status, port, java_version, xms_mb, xmx_mb, autostart, eula_accepted)
			VALUES (?, ?, ?, 'stopped', ?, ?, ?, ?, ?, ?)`,
			serverID, req.Name, req.JarPath, req.Port, req.JavaVersion, req.XmsMB, req.XmxMB, req.AutoStart, req.EULA)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to save server"})
		}

		eulaPath := filepath.Join(serverDir, "eula.txt")
		eulaContent := fmt.Sprintf("eula=%v\n", req.EULA)
		if err := os.WriteFile(eulaPath, []byte(eulaContent), 0644); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create eula.txt"})
		}

		propsPath := filepath.Join(serverDir, "server.properties")
		props := fmt.Sprintf("server-port=%d\n", req.Port)
		if err := os.WriteFile(propsPath, []byte(props), 0644); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create server.properties"})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": serverID, "name": req.Name})
	}
}

func ListServers(db *database.DB, cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rows, err := db.Query("SELECT id, name, jar_path, status, port, java_version, xms_mb, xmx_mb, autostart, eula_accepted, created_at FROM servers")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to query servers"})
		}
		defer rows.Close()

		var servers []ServerInfo
		for rows.Next() {
			var s ServerInfo
			if err := rows.Scan(&s.ID, &s.Name, &s.JarPath, &s.Status, &s.Port,
				&s.JavaVersion, &s.XmsMB, &s.XmxMB, &s.AutoStart, &s.EULAAccepted, &s.CreatedAt); err != nil {
				continue
			}
			servers = append(servers, s)
		}
		return c.JSON(fiber.Map{"servers": servers})
	}
}

func GetServer(db *database.DB, cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		serverID := c.Params("id")
		var s ServerInfo
		err := db.QueryRow("SELECT id, name, jar_path, status, port, java_version, xms_mb, xmx_mb, autostart, eula_accepted, created_at FROM servers WHERE id = ?", serverID).
			Scan(&s.ID, &s.Name, &s.JarPath, &s.Status, &s.Port, &s.JavaVersion, &s.XmsMB, &s.XmxMB, &s.AutoStart, &s.EULAAccepted, &s.CreatedAt)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "server not found"})
		}
		return c.JSON(s)
	}
}

func StartServer(db *database.DB, cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		serverID := c.Params("id")

		var s struct {
			JarPath     string
			JavaVersion sql.NullString
			XmsMB       int
			XmxMB       int
		}
		err := db.QueryRow("SELECT jar_path, java_version, xms_mb, xmx_mb FROM servers WHERE id = ?", serverID).
			Scan(&s.JarPath, &s.JavaVersion, &s.XmsMB, &s.XmxMB)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "server not found"})
		}

		jarPath := filepath.Join(config.GetServerDataDir(cfg.DataDir, serverID), filepath.Base(s.JarPath))
		if _, err := os.Stat(jarPath); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "jar file not found"})
		}

		javaPath := s.JavaVersion.String
		if javaPath == "" {
			jm := service.NewJavaManager(cfg)
			osName, _ := jm.DetectSystemOS()
			arch, _ := jm.DetectArch()
			version := 17
			if s.JarPath != "" {
				version = service.MapGameVersionToJava(s.JarPath)
			}
			javaPath = filepath.Join(config.GetJDKCacheDir(cfg.DataDir), fmt.Sprintf("jdk-%d", version))
			if _, err := os.Stat(javaPath); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "java not installed"})
			}
		}

		sm := service.NewServerManager(cfg)
		sp := sm.CreateProcess(serverID, "unknown", jarPath, javaPath, s.XmsMB, s.XmxMB, "")
		if err := sm.Start(serverID); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		db.Exec("UPDATE servers SET status = 'running' WHERE id = ?", serverID)

		return c.JSON(fiber.Map{"status": "started", "id": serverID})
	}
}

func StopServer(db *database.DB, cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		serverID := c.Params("id")
		sm := service.NewServerManager(cfg)

		if err := sm.Stop(serverID, 30*time.Second); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		db.Exec("UPDATE servers SET status = 'stopped' WHERE id = ?", serverID)

		return c.JSON(fiber.Map{"status": "stopped", "id": serverID})
	}
}

func RestartServer(db *database.DB, cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		serverID := c.Params("id")
		sm := service.NewServerManager(cfg)

		if err := sm.Stop(serverID, 30*time.Second); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		if err := sm.Start(serverID); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		db.Exec("UPDATE servers SET status = 'running' WHERE id = ?", serverID)

		return c.JSON(fiber.Map{"status": "restarted", "id": serverID})
	}
}

func DeleteServer(db *database.DB, cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		serverID := c.Params("id")

		// Stop server if running
		sm := service.NewServerManager(cfg)
		if _, ok := sm.GetProcess(serverID); ok {
			sm.Stop(serverID, 10*time.Second)
		}

		// Remove from DB
		db.Exec("DELETE FROM servers WHERE id = ?", serverID)

		// Remove server directory
		serverDir := config.GetServerDataDir(cfg.DataDir, serverID)
		os.RemoveAll(serverDir)

		return c.JSON(fiber.Map{"status": "deleted", "id": serverID})
	}
}

func isValidServerName(name string) bool {
	if len(name) == 0 || len(name) > 50 {
		return false
	}
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
			return false
		}
	}
	return true
}

func generateID() string {
	return fmt.Sprintf("srv_%d", fiber.Now().UnixNano())
}

func ListJDKs(db *database.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rows, err := db.Query("SELECT id, version, arch, os, install_path, created_at FROM jdk_versions")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to query JDKs"})
		}
		defer rows.Close()

		var jdks []struct {
			ID          string
			Version     string
			Arch        string
			OS          string
			InstallPath string
			CreatedAt   string
		}
		for rows.Next() {
			var jdk struct {
				ID          string
				Version     string
				Arch        string
				OS          string
				InstallPath string
				CreatedAt   string
			}
			if err := rows.Scan(&jdk.ID, &jdk.Version, &jdk.Arch, &jdk.OS, &jdk.InstallPath, &jdk.CreatedAt); err != nil {
				continue
			}
			jdks = append(jdks, jdk)
		}
		return c.JSON(fiber.Map{"jdk_versions": jdks})
	}
}

func InstallJDK(db *database.DB, cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req struct {
			Version int `json:"version" validate:"required"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
		}

		jm := service.NewJavaManager(cfg)
		osName, err := jm.DetectSystemOS()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		arch, err := jm.DetectArch()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		// В реальном коде — асинхронное скачивание с WebSocket прогрессом
		javaPath, err := jm.DownloadJDK(c.Context(), req.Version, osName, arch)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		db.Exec(`INSERT INTO jdk_versions (id, version, arch, os, install_path) VALUES (?, ?, ?, ?, ?)`,
			fmt.Sprintf("jdk_%d", req.Version), strconv.Itoa(req.Version), arch, osName, javaPath)

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": "installed", "version": req.Version})
	}
}

func DownloadJAR(db *database.DB, cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req struct {
			URL      string `json:"url" validate:"required"`
			ServerID string `json:"server_id" validate:"required"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
		}

		filename := filepath.Base(req.URL)
		dl := service.NewDownloader(cfg)
		result, err := dl.Download(c.Context(), req.URL, filename)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		if result.ErrorCode != "" {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": result.ErrorCode})
		}

		db.Exec(`INSERT INTO downloaded_jars (server_id, filename, url, checksum, size_bytes) VALUES (?, ?, ?, ?, ?)`,
			req.ServerID, filename, req.URL, result.Checksum, result.Size)

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": "downloaded", "path": result.Path})
	}
}

func GetMetrics(db *database.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		serverID := c.Params("id")

		// Получаем последние метрики
		var m struct {
			TPS         sql.NullFloat64
			RAMMB       sql.NullInt64
			CPUPercent  sql.NullFloat64
			OnlinPlayers sql.NullInt64
		}
		err := db.QueryRow(`SELECT tps, ram_mb, cpu_percent, online_players FROM metrics_history WHERE server_id = ? ORDER BY timestamp DESC LIMIT 1`, serverID).
			Scan(&m.TPS, &m.RAMMB, &m.CPUPercent, &m.OnlinPlayers)

		if err != nil && err != sql.ErrNoRows {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{
			"tps":           m.TPS.Float64,
			"ram_mb":        m.RAMMB.Int64,
			"cpu_percent":   m.CPUPercent.Float64,
			"online_players": m.OnlinPlayers.Int64,
			"timestamp":     time.Now().Format(time.RFC3339),
		})
	}
}

func GetMetricsHistory(db *database.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		serverID := c.Params("id")

		rows, err := db.Query(`SELECT timestamp, tps, mspt, online_players, ram_mb, cpu_percent FROM metrics_history WHERE server_id = ? ORDER BY timestamp DESC LIMIT 100`, serverID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		defer rows.Close()

		var history []struct {
			Timestamp    string
			TPS          sql.NullFloat64
			MSPT         sql.NullFloat64
			OnlinPlayers sql.NullInt64
			RAMMB        sql.NullInt64
			CPUPercent   sql.NullFloat64
		}

		for rows.Next() {
			var h struct {
				Timestamp    string
				TPS          sql.NullFloat64
				MSPT         sql.NullFloat64
				OnlinPlayers sql.NullInt64
				RAMMB        sql.NullInt64
				CPUPercent   sql.NullFloat64
			}
			if err := rows.Scan(&h.Timestamp, &h.TPS, &h.MSPT, &h.OnlinPlayers, &h.RAMMB, &h.CPUPercent); err != nil {
				continue
			}
			history = append(history, h)
		}

		return c.JSON(fiber.Map{"history": history})
	}
}

func SearchPlugins(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		query := c.Query("q", "")
		filters := map[string][]string{
			"versions": strings.Split(c.Query("versions", ""), ","),
			"loaders":  strings.Split(c.Query("loaders", ""), ","),
		}

		pa := service.NewPluginAggregator(cfg)
		plugins, err := pa.Search(query, filters)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{"plugins": plugins})
	}
}

func DownloadPlugin(db *database.DB, cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req struct {
			PluginID string `json:"plugin_id" validate:"required"`
			ServerID string `json:"server_id" validate:"required"`
			Source   string `json:"source" validate:"required"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
		}

		pa := service.NewPluginAggregator(cfg)
		// В реальном коде — скачивание и распаковка в plugins/ сервера
		return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"error": "not implemented"})
	}
}
