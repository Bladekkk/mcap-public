package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/bladekkk/mcap/backend/internal/config"
)

type PluginAggregator struct {
	cfg      *config.Config
	client   *http.Client
	cache    map[string]cacheEntry
	cacheTTL time.Duration
}

type cacheEntry struct {
	data      []byte
	createdAt time.Time
}

func NewPluginAggregator(cfg *config.Config) *PluginAggregator {
	return &PluginAggregator{
		cfg:      cfg,
		client:   &http.Client{Timeout: 30 * time.Second},
		cache:    make(map[string]cacheEntry),
		cacheTTL: 10 * time.Minute,
	}
}

type PluginInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Description string `json:"description"`
	Author    string `json:"author"`
	Downloads int    `json:"downloads"`
	Version   string `json:"version"`
	License   string `json:"license"`
	Website   string `json:"website"`
	DownloadURL string `json:"download_url"`
	Source    string `json:"source"`
}

func (pa *PluginAggregator) Search(query string, filters map[string][]string) ([]PluginInfo, error) {
	cacheKey := pa.buildCacheKey(query, filters)
	if entry, ok := pa.getCached(cacheKey); ok {
		var result []PluginInfo
		if err := json.Unmarshal(entry.data, &result); err == nil {
			return result, nil
		}
	}

	var allPlugins []PluginInfo

	// Modrinth
	if plugins, err := pa.searchModrinth(query, filters); err == nil {
		allPlugins = append(allPlugins, plugins...)
	}

	// Hangar
	if plugins, err := pa.searchHangar(query, filters); err == nil {
		allPlugins = append(allPlugins, plugins...)
	}

	// CurseForge (если есть ключ)
	if pa.cfg.CurseForgeKey != "" {
		if plugins, err := pa.searchCurseForge(query, filters); err == nil {
			allPlugins = append(allPlugins, plugins...)
		}
	}

	pa.setCache(cacheKey, allPlugins)
	return allPlugins, nil
}

func (pa *PluginAggregator) searchModrinth(query string, filters map[string][]string) ([]PluginInfo, error) {
	baseURL := "https://api.modrinth.com/v2/search"
	params := url.Values{}
	params.Set("query", query)
	params.Set("facets", pa.buildModrinthFacets(filters))

	resp, err := pa.client.Get(baseURL + "?" + params.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Hits []struct {
			ProjectID   string `json:"project_id"`
			Title       string `json:"title"`
			Description string `json:"description"`
			Author      string `json:"author"`
			Downloads   int64  `json:"downloads"`
			Files       []struct {
				Url string `json:"url"`
			} `json:"files"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var plugins []PluginInfo
	for _, h := range result.Hits {
		if len(h.Files) > 0 {
			plugins = append(plugins, PluginInfo{
				ID:          h.ProjectID,
				Name:        h.Title,
				Description: h.Description,
				Author:      h.Author,
				Downloads:   int(h.Downloads),
				DownloadURL: h.Files[0].Url,
				Source:      "modrinth",
			})
		}
	}
	return plugins, nil
}

func (pa *PluginAggregator) buildModrinthFacets(filters map[string][]string) string {
	var facets []string
	if versions, ok := filters["versions"]; ok && len(versions) > 0 {
		for _, v := range versions {
			facets = append(facets, fmt.Sprintf(`"versions:%s"`, v))
		}
	}
	if cores, ok := filters["loaders"]; ok && len(cores) > 0 {
		for _, c := range cores {
			facets = append(facets, fmt.Sprintf(`"categories:%s"`, c))
		}
	}
	if len(facets) == 0 {
		return "[]"
	}
	return "[" + strings.Join(facets, ",") + "]"
}

func (pa *PluginAggregator) searchHangar(query string, filters map[string][]string) ([]PluginInfo, error) {
	baseURL := "https://hangar.papermc.io/api/v1/projects"
	if query != "" {
		baseURL += "?q=" + url.QueryEscape(query)
	}

	resp, err := pa.client.Get(baseURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result []struct {
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
		Author      string `json:"author"`
		Downloads   int64  `json:"downloads"`
		Links       struct {
			Website string `json:"website"`
		} `json:"links"`
		versions []struct {
			ProjectVersion string `json:"projectVersion"`
			IsStable       bool   `json:"isStable"`
		} `json:"versions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var plugins []PluginInfo
	for _, p := range result {
		plugins = append(plugins, PluginInfo{
			ID:          p.Slug,
			Name:        p.Name,
			Description: p.Description,
			Author:      p.Author,
			Downloads:   int(p.Downloads),
			Website:     p.Links.Website,
			Source:      "hangar",
		})
	}
	return plugins, nil
}

func (pa *PluginAggregator) searchCurseForge(query string, filters map[string][]string) ([]PluginInfo, error) {
	// CurseForge API требует API ключ, реализация упрощена
	// В продакшене использовать официальный API: https://docs.curseforge.com/
	return nil, nil
}

func (pa *PluginAggregator) buildCacheKey(query string, filters map[string][]string) string {
	var parts []string
	parts = append(parts, "query="+query)
	for k, v := range filters {
		parts = append(parts, k+"="+strings.Join(v, ","))
	}
	return strings.Join(parts, "&")
}

func (pa *PluginAggregator) getCached(key string) ([]byte, bool) {
	entry, ok := pa.cache[key]
	if !ok {
		return nil, false
	}
	if time.Since(entry.createdAt) > pa.cacheTTL {
		delete(pa.cache, key)
		return nil, false
	}
	return entry.data, true
}

func (pa *PluginAggregator) setCache(key string, data []PluginInfo) {
	jsonData, _ := json.Marshal(data)
	pa.cache[key] = cacheEntry{
		data:      jsonData,
		createdAt: time.Now(),
	}
}
