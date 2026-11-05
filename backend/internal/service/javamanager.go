package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/bladekkk/mcap/backend/internal/config"
)

type JavaManager struct {
	cfg *config.Config
}

func NewJavaManager(cfg *config.Config) *JavaManager {
	return &JavaManager{cfg: cfg}
}

type JDKInfo struct {
	Version    string
	Path       string
	Size       int64
	Installed  bool
}

func (jm *JavaManager) DetectSystemOS() (string, error) {
	switch runtime.GOOS {
	case "linux":
		return "linux", nil
	case "windows":
		return "windows", nil
	case "darwin":
		return "macos", nil
	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

func (jm *JavaManager) DetectArch() (string, error) {
	switch runtime.GOARCH {
	case "amd64":
		return "x64", nil
	case "arm64":
		return "aarch64", nil
	case "386":
		return "x32", nil
	default:
		return "", fmt.Errorf("unsupported arch: %s", runtime.GOARCH)
	}
}

// MapGameVersionToJava определяет нужную версию Java по версии Minecraft
func MapGameVersionToJava(version string) int {
	// Парсинг версии (пример: "1.20.4", "1.21")
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return 8
	}

_MAJOR, err := strconv.Atoi(parts[0])
		if err != nil {
			return 8
		}
	MINOR, err := strconv.Atoi(parts[1])
		if err != nil {
			return 8
		}

	if MAJOR == 1 {
		if MINOR <= 12 {
			return 8
		} else if MINOR <= 15 {
			return 8
		} else if MINOR <= 16 {
			return 11
		} else if MINOR <= 17 {
			return 11
		} else if MINOR <= 19 {
			return 17
		} else if MINOR <= 20 {
			return 17
		} else if MINOR <= 21 {
			return 21
		}
	}
	return 17
}

func (jm *JavaManager) GetJDKURL(featureVersion int, osName, arch string) (string, error) {
	// Adoptium API: https://api.adoptium.net/v3/binary/latest/{feature_version}/ga/{os}/{arch}/jdk/hotspot/normal/eclipse
	return fmt.Sprintf("https://api.adoptium.net/v3/binary/latest/%d/ga/%s/%s/jdk/hotspot/normal/eclipse", featureVersion, osName, arch), nil
}

func (jm *JavaManager) DownloadJDK(ctx context.Context, version int, osName, arch string) (string, error) {
	url, err := jm.GetJDKURL(version, osName, arch)
	if err != nil {
		return "", err
	}

	// Получаем info о JDK
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var jdkInfo struct {
		Binary struct {
			Link string `json:"link"`
			Name string `json:"name"`
		} `json:"binary"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&jdkInfo); err != nil {
		return "", err
	}

	// Скачиваем JDK
	dl := NewDownloader(jm.cfg)
	result, err := dl.Download(ctx, jdkInfo.Binary.Link, jdkInfo.Binary.Name)
	if err != nil {
		return "", err
	}
	if result.ErrorCode != "" {
		return "", fmt.Errorf("download failed: %s", result.ErrorCode)
	}

	return result.Path, nil
}

func (jm *JavaManager) GetInstalledJDKs() ([]JDKInfo, error) {
	jdkDir := config.GetJDKCacheDir(jm.cfg.DataDir)
	entries, err := os.ReadDir(jdkDir)
	if err != nil {
		return nil, err
	}

	var jdks []JDKInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		jdks = append(jdks, JDKInfo{
			Version:    entry.Name(),
			Path:       filepath.Join(jdkDir, entry.Name()),
			Size:       info.Size(),
			Installed:  true,
		})
	}
	return jdks, nil
}

func (jm *JavaManager) ExtractJDK(archivePath string) (string, error) {
	// Определяем тип архива и распаковываем
	dir := filepath.Dir(archivePath)

	// Для упрощения — возвращаем путь к архиву
	// В продакшене добавить распаковку tar.gz/zip
	return archivePath, nil
}

func (jm *JavaManager) ValidateJavaVersion(javaPath string) (string, error) {
	cmd := exec.Command(javaPath, "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	// Парсим output: "openjdk version "17.0.9" 2023-10-17"
	outStr := string(output)
	parts := strings.Split(outStr, `"`)
	if len(parts) >= 2 {
		return parts[1], nil
	}
	return "unknown", nil
}
