package ui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

var playlistsPerPageOptions = []int{8, 12, 16, 24}

var perPageLabels = []string{"8", "12", "16", "24"}

type Config struct {
	OutputDir        string `json:"output_dir"`
	PlaylistsPerPage int    `json:"playlists_per_page"`
}

func defaultOutputDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "./AppleMusicExports"
	}
	return filepath.Join(home, "AppleMusicExports")
}

func defaultConfig() Config {
	return Config{
		OutputDir:        defaultOutputDir(),
		PlaylistsPerPage: 12,
	}
}

func configPath() (string, error) {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "playlist-md", "config.json"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "playlist-md", "config.json"), nil
}

func loadConfig() Config {
	cfg := defaultConfig()
	path, err := configPath()
	if err != nil {
		return cfg
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}
	_ = json.Unmarshal(data, &cfg)
	if cfg.OutputDir == "" {
		cfg.OutputDir = defaultOutputDir()
	}
	cfg.PlaylistsPerPage = normalizePerPage(cfg.PlaylistsPerPage)
	return cfg
}

func saveConfig(cfg Config) {
	cfg.PlaylistsPerPage = normalizePerPage(cfg.PlaylistsPerPage)
	path, err := configPath()
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(path, append(data, '\n'), 0o600)
}

func expandPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if path == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			return home
		}
		return path
	}
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

func normalizePerPage(n int) int {
	best := playlistsPerPageOptions[1]
	bestDist := abs(n - best)
	for _, option := range playlistsPerPageOptions {
		if d := abs(n - option); d < bestDist {
			best = option
			bestDist = d
		}
	}
	return best
}

func cyclePerPage(current, delta int) int {
	current = normalizePerPage(current)
	idx := 0
	for i, option := range playlistsPerPageOptions {
		if option == current {
			idx = i
			break
		}
	}
	idx = (idx + delta) % len(playlistsPerPageOptions)
	if idx < 0 {
		idx += len(playlistsPerPageOptions)
	}
	return playlistsPerPageOptions[idx]
}

func indexOfInt(steps []int, cur int) int {
	for i, n := range steps {
		if n == cur {
			return i
		}
	}
	return 0
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
