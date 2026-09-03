package engine

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/imflawlezz/playlist-md/assets"
)

type Client struct {
	corePath string
}

func NewClient() (*Client, error) {
	path, err := coreBinaryPath()
	if err != nil {
		return nil, err
	}
	return &Client{corePath: path}, nil
}

func coreBinaryPath() (string, error) {
	if override := os.Getenv("PLAYLIST_MD_CORE"); override != "" {
		if _, err := os.Stat(override); err == nil {
			return override, nil
		}
		return "", fmt.Errorf("PLAYLIST_MD_CORE points to missing file: %s", override)
	}

	if len(assets.Core) == 0 {
		return "", fmt.Errorf("embedded %s is missing; run make", CoreName)
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256(assets.Core)
	cached := filepath.Join(cacheDir, "playlist-md", hex.EncodeToString(sum[:16]), CoreName)
	if info, err := os.Stat(cached); err == nil && info.Size() == int64(len(assets.Core)) {
		return cached, nil
	}

	if err := os.MkdirAll(filepath.Dir(cached), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(cached, assets.Core, 0o755); err != nil {
		return "", err
	}
	return cached, nil
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type StatusResponse struct {
	Status string `json:"status"`
}

type VersionResponse struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Playlist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Track struct {
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	Album    string `json:"album"`
	Year     *int   `json:"year"`
	Position int    `json:"position"`
}

type PlaylistDetail struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Tracks []Track `json:"tracks"`
}

type PlaylistsResponse struct {
	Playlists []Playlist `json:"playlists"`
}

type ExportResult struct {
	ExportedPlaylists     int      `json:"exported_playlists"`
	ExportedTracks        int      `json:"exported_tracks"`
	ExportedLibraryTracks int      `json:"exported_library_tracks"`
	Output                string   `json:"output"`
	RemovedStaleFiles     []string `json:"removed_stale_files"`
	LogPath               string   `json:"log_path,omitempty"`
}

type ProgressEvent struct {
	Type  string `json:"type"`
	Phase string `json:"phase"`
	Name  string `json:"name,omitempty"`
	Index int    `json:"index,omitempty"`
	Total int    `json:"total,omitempty"`
}

func (c *Client) run(args ...string) (stdout []byte, stderr []byte, err error) {
	cmd := exec.Command(c.corePath, args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err = cmd.Run()
	return outBuf.Bytes(), errBuf.Bytes(), err
}

func decodeError(stderr []byte, err error) error {
	text := strings.TrimSpace(string(stderr))
	if text != "" {
		lines := strings.Split(text, "\n")
		// Export may emit progress JSON on stderr before {"error":"..."}; prefer the last error object.
		for i := len(lines) - 1; i >= 0; i-- {
			line := strings.TrimSpace(lines[i])
			if line == "" {
				continue
			}
			var resp ErrorResponse
			if json.Unmarshal([]byte(line), &resp) == nil && resp.Error != "" {
				return fmt.Errorf("%s", resp.Error)
			}
		}
		for i := len(lines) - 1; i >= 0; i-- {
			line := strings.TrimSpace(lines[i])
			if line != "" {
				return fmt.Errorf("%s", line)
			}
		}
	}
	if err != nil {
		return err
	}
	return fmt.Errorf("unknown error from %s", CoreName)
}

func (c *Client) Version() (VersionResponse, error) {
	out, stderr, err := c.run("version")
	if err != nil {
		return VersionResponse{}, decodeError(stderr, err)
	}
	var resp VersionResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		return VersionResponse{}, err
	}
	return resp, nil
}

func (c *Client) Status() (string, error) {
	out, stderr, err := c.run("status")
	if err != nil {
		return "", decodeError(stderr, err)
	}
	var resp StatusResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		return "", err
	}
	return resp.Status, nil
}

func (c *Client) Authorize() (string, error) {
	out, stderr, err := c.run("authorize")
	if err != nil {
		return "", decodeError(stderr, err)
	}
	var resp StatusResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		return "", err
	}
	return resp.Status, nil
}

func (c *Client) ListPlaylists() ([]Playlist, error) {
	out, stderr, err := c.run("list-playlists")
	if err != nil {
		return nil, decodeError(stderr, err)
	}
	var resp PlaylistsResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		return nil, err
	}
	return resp.Playlists, nil
}

func (c *Client) GetPlaylist(id string) (PlaylistDetail, error) {
	out, stderr, err := c.run("get-playlist", "--id", id)
	if err != nil {
		return PlaylistDetail{}, decodeError(stderr, err)
	}
	var detail PlaylistDetail
	if err := json.Unmarshal(out, &detail); err != nil {
		return PlaylistDetail{}, err
	}
	return detail, nil
}

func (c *Client) IndexTracks(onPlaylist func(PlaylistDetail)) error {
	cmd := exec.Command(c.corePath, "index-tracks")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf
	if err := cmd.Start(); err != nil {
		return err
	}

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var detail PlaylistDetail
		if json.Unmarshal([]byte(line), &detail) != nil || detail.ID == "" {
			continue
		}
		if onPlaylist != nil {
			onPlaylist(detail)
		}
	}
	waitErr := cmd.Wait()
	if scanErr := scanner.Err(); scanErr != nil {
		return scanErr
	}
	if waitErr != nil {
		return decodeError(errBuf.Bytes(), waitErr)
	}
	return nil
}

func (c *Client) Export(output string, ids []string, exportAll bool, writeLogs bool, onProgress func(ProgressEvent)) (ExportResult, error) {
	args := []string{"export", "--output", output}
	if exportAll {
		args = append(args, "--all")
	} else if len(ids) > 0 {
		args = append(args, "--ids", strings.Join(ids, ","))
	}
	return c.runExport(args, writeLogs, onProgress)
}

func (c *Client) ExportLibrary(output string, writeLogs bool, onProgress func(ProgressEvent)) (ExportResult, error) {
	args := []string{"export", "--output", output, "--library"}
	return c.runExport(args, writeLogs, onProgress)
}

func (c *Client) runExport(args []string, writeLogs bool, onProgress func(ProgressEvent)) (ExportResult, error) {
	if writeLogs {
		args = append(args, "--write-logs")
	} else {
		args = append(args, "--no-write-logs")
	}

	cmd := exec.Command(c.corePath, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return ExportResult{}, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return ExportResult{}, err
	}

	if err := cmd.Start(); err != nil {
		return ExportResult{}, err
	}

	var errBuf bytes.Buffer
	progressErr := readProgress(io.TeeReader(stderr, &errBuf), onProgress)
	stdoutBytes, readErr := io.ReadAll(stdout)
	waitErr := cmd.Wait()

	if progressErr != nil {
		return ExportResult{}, progressErr
	}
	if readErr != nil {
		return ExportResult{}, readErr
	}
	if waitErr != nil {
		return ExportResult{}, decodeError(errBuf.Bytes(), waitErr)
	}

	var result ExportResult
	if err := json.Unmarshal(stdoutBytes, &result); err != nil {
		return ExportResult{}, err
	}
	return result, nil
}

func readProgress(r io.Reader, onProgress func(ProgressEvent)) error {
	buf := make([]byte, 4096)
	remaining := ""
	for {
		n, err := r.Read(buf)
		if n > 0 {
			remaining += string(buf[:n])
			for {
				idx := strings.IndexByte(remaining, '\n')
				if idx < 0 {
					break
				}
				line := strings.TrimSpace(remaining[:idx])
				remaining = remaining[idx+1:]
				if line == "" {
					continue
				}
				if onProgress == nil {
					continue
				}
				var event ProgressEvent
				if json.Unmarshal([]byte(line), &event) == nil && event.Type == "progress" {
					onProgress(event)
				}
			}
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func Forward(args []string) int {
	client, err := NewClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	stdout, stderr, err := client.run(args...)
	if err != nil {
		fmt.Fprintln(os.Stderr, decodeError(stderr, err))
		return 1
	}
	if len(stdout) > 0 {
		os.Stdout.Write(stdout)
	}
	if len(stderr) > 0 {
		os.Stderr.Write(stderr)
	}
	return 0
}
