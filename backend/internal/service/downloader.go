package service

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/bladekkk/mcap/backend/internal/config"
)

type Downloader struct {
	cfg    *config.Config
	client *http.Client
}

func NewDownloader(cfg *config.Config) *Downloader {
	return &Downloader{
		cfg:    cfg,
		client: &http.Client{Timeout: 300 * time.Second},
	}
}

type DownloadResult struct {
	Path      string
	Size      int64
	Checksum  string
	ErrorCode string
}

func (d *Downloader) Download(ctx context.Context, url, filename string) (*DownloadResult, error) {
	cacheDir := config.GetJARCacheDir(d.cfg.DataDir)
	cachePath := filepath.Join(cacheDir, filename)

	// Проверка кэша
	if _, err := os.Stat(cachePath); err == nil {
		checksum, _ := d.calculateChecksum(cachePath)
		return &DownloadResult{Path: cachePath, Checksum: checksum}, nil
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &DownloadResult{ErrorCode: fmt.Sprintf("HTTP %d", resp.StatusCode)}, nil
	}

	tmpPath := filepath.Join(cacheDir, filename+".tmp")
	f, err := os.Create(tmpPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	total := resp.ContentLength
	downloaded := int64(0)
	buf := make([]byte, 32*1024)

	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			written, writeErr := f.Write(buf[:n])
			if writeErr != nil {
				return nil, writeErr
			}
			downloaded += int64(written)
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return nil, readErr
		}
	}

	if err := os.Rename(tmpPath, cachePath); err != nil {
		os.Remove(tmpPath)
		return nil, err
	}

	checksum, _ := d.calculateChecksum(cachePath)
	return &DownloadResult{Path: cachePath, Size: downloaded, Checksum: checksum}, nil
}

func (d *Downloader) calculateChecksum(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

type VersionInfo struct {
	ID      string
	Type    string
	URL     string
	Release string
	Builds  []int
	Source  string
}

type VanillaProvider struct{}

func (p *VanillaProvider) GetVersions() ([]VersionInfo, error) {
	resp, err := http.Get("https://piston-meta.mojang.com/mc/game/version_manifest_v2.json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var manifest struct {
		Versions []struct {
			ID      string `json:"id"`
			Type    string `json:"type"`
			URL     string `json:"url"`
			Release string `json:"release"`
		} `json:"versions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, err
	}

	var versions []VersionInfo
	for _, v := range manifest.Versions {
		if v.Type == "release" || v.Type == "snapshot" {
			versions = append(versions, VersionInfo{
				ID:      v.ID,
				Type:    v.Type,
				URL:     v.URL,
				Release: v.Release,
			})
		}
	}
	return versions, nil
}

type PaperProvider struct{}

func (p *PaperProvider) GetVersions() ([]VersionInfo, error) {
	resp, err := http.Get("https://api.papermc.io/v2/projects/paper")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		ProjectID string `json:"project_id"`
		Versions  []struct {
			Version string `json:"version"`
			Builds  []int  `json:"builds"`
		} `json:"versions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var versions []VersionInfo
	for _, v := range result.Versions {
		versions = append(versions, VersionInfo{
			ID:     v.Version,
			Builds: v.Builds,
			Source: "paper",
		})
	}
	return versions, nil
}

func (p *PaperProvider) GetBuilds(version string) ([]int, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.papermc.io/v2/projects/paper/versions/%s/builds", version))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Builds []int `json:"builds"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Builds, nil
}
