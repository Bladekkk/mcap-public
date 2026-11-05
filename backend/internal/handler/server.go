package handler

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bladekkk/mcap/backend/internal/config"
	"github.com/bladekkk/mcap/backend/internal/database"
	"github.com/gofiber/fiber/v2"
)

type CreateServerRequest struct {
	Name       string `json:"name" validate:"required,min=1,max=50,alphanum"`
	JarPath    string `json:"jar_path" validate:"required"`
	JavaVersion string `json:"java_version"`
	XmsMB      int    `json:"xms_mb" validate:"min=256"`
	XmxMB      int    `json:"xmx_mb" validate:"min=256"`
	Port       int    `json:"port" validate:"min=1,max=65535"`
	AutoStart  bool   `json:"autostart"`
	EULA       bool   `json:"eula_accepted"`
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

		serverID := generateID()
		serverDir := config.GetServerDataDir(cfg.DataDir, serverID)
		if err := os.MkdirAll(serverDir, 0755); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create server dir"})
		}

		// Валидация имени
		if !isValidServerName(req.Name) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid server name"})
		}

		// Запись в БД
		_, err := db.Exec(`INSERT INTO servers (id, name, jar_path, status, port, java_version, xms_mb, xmx_mb, autostart, eula_accepted)
			VALUES (?, ?, ?, 'stopped', ?, ?, ?, ?, ?, ?)`,
			serverID, req.Name, req.JarPath, req.Port, req.JavaVersion, req.XmsMB, req.XmxMB, req.AutoStart, req.EULA)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to save server"})
		}

		// Создание eula.txt
		eulaPath := filepath.Join(serverDir, "eula.txt")
		eulaContent := fmt.Sprintf("eula=%v\n", req.EULA)
		if err := os.WriteFile(eulaPath, []byte(eulaContent), 0644); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create eula.txt"})
		}

		// Создание server.properties
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

		var servers []fiber.Map
		for rows.Next() {
			var s struct {
				ID           string
				Name         string
				JarPath      string
				Status       string
				Port         int
				JavaVersion  sql.NullString
				XmsMB        int
				XmxMB        int
				AutoStart    bool
				EULAAccepted bool
				CreatedAt    string
			}
			if err := rows.Scan(&s.ID, &s.Name, &s.JarPath, &s.Status, &s.Port,
				&s.JavaVersion, &s.XmsMB, &s.XmxMB, &s.AutoStart, &s.EULAAccepted, &s.CreatedAt); err != nil {
				continue
			}
			servers = append(servers, fiber.Map{
				"id":            s.ID,
				"name":          s.Name,
				"jar_path":      s.JarPath,
				"status":        s.Status,
				"port":          s.Port,
				"java_version":  s.JavaVersion.String,
				"xms_mb":        s.XmsMB,
				"xmx_mb":        s.XmxMB,
				"autostart":     s.AutoStart,
				"eula_accepted": s.EULAAccepted,
				"created_at":    s.CreatedAt,
			})
		}
		return c.JSON(fiber.Map{"servers": servers})
	}
}

func GetServer(db *database.DB, cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		serverID := c.Params("id")
		var s struct {
			ID           string
			Name         string
			JarPath      string
			Status       string
			Port         int
			JavaVersion  sql.NullString
			XmsMB        int
			XmxMB        int
			AutoStart    bool
			EULAAccepted bool
		}
		err := db.QueryRow("SELECT id, name, jar_path, status, port, java_version, xms_mb, xmx_mb, autostart, eula_accepted FROM servers WHERE id = ?", serverID).
			Scan(&s.ID, &s.Name, &s.JarPath, &s.Status, &s.Port, &s.JavaVersion, &s.XmsMB, &s.XmxMB, &s.AutoStart, &s.EULAAccepted)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "server not found"})
		}
		return c.JSON(fiber.Map{
			"id":            s.ID,
			"name":          s.Name,
			"jar_path":      s.JarPath,
			"status":        s.Status,
			"port":          s.Port,
			"java_version":  s.JavaVersion.String,
			"xms_mb":        s.XmsMB,
			"xmx_mb":        s.XmxMB,
			"autostart":     s.AutoStart,
			"eula_accepted": s.EULAAccepted,
		})
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
	// Use sonypflake or simple UUID-based ID
	return fmt.Sprintf("srv_%d", fiber.Now().UnixNano())
}
