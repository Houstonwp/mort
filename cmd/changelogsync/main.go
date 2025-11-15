package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	changeLogURL   = "https://mort.soa.org/WebService.asmx/GetListOfChangeLogs"
	xmlDownloadFmt = "https://mort.soa.org/data/t%d.xml"
)

type changeLogResponse struct {
	D struct {
		Total   int           `json:"total"`
		Page    int           `json:"page"`
		Records int           `json:"records"`
		Rows    []changeEntry `json:"rows"`
	} `json:"d"`
}

type changeEntry struct {
	TableIdentity int    `json:"TableIdentity"`
	TableID       int    `json:"FkXtbml"`
	Comment       string `json:"Comment"`
	SDate         string `json:"sDate"`
	LogDate       string `json:"LogDate"`
	UserID        string `json:"UserID"`
	Action        string `json:"Action"`
}

type stateFile struct {
	LastLogMillis int64 `json:"last_log_ms"`
}

func main() {
	var (
		statePath string
		xmlDir    string
		singleID  int
	)

	flag.StringVar(&statePath, "state", filepath.Join("json", "changelog_state.json"), "path to changelog state file")
	flag.StringVar(&xmlDir, "xml-dir", "xml", "directory where XML files are stored")
	flag.IntVar(&singleID, "table", 0, "fetch a single table id and skip changelog scan")
	flag.Parse()

	if singleID > 0 {
		if err := downloadTable(singleID, xmlDir); err != nil {
			fmt.Fprintf(os.Stderr, "failed to download table %d: %v\n", singleID, err)
			os.Exit(1)
		}
		fmt.Printf("downloaded t%d.xml\n", singleID)
		return
	}

	state, err := loadState(statePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load state: %v\n", err)
		os.Exit(1)
	}

	entries, err := fetchNewEntries(state.LastLogMillis)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to fetch change log: %v\n", err)
		os.Exit(1)
	}

	if len(entries) == 0 {
		fmt.Println("no new change log entries")
		return
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].logMillis() < entries[j].logMillis()
	})

	var maxLog int64 = state.LastLogMillis
	processed := 0
	for _, entry := range entries {
		logTs := entry.logMillis()
		action := strings.ToLower(entry.Action)
		switch {
		case strings.Contains(action, "delete"):
			if err := removeTable(entry.TableID, xmlDir); err != nil {
				fmt.Fprintf(os.Stderr, "failed removing t%d: %v\n", entry.TableID, err)
				os.Exit(1)
			}
			fmt.Printf("removed t%d.xml (log #%d on %s)\n", entry.TableID, entry.TableIdentity, entry.SDate)
		default:
			if err := downloadTable(entry.TableID, xmlDir); err != nil {
				fmt.Fprintf(os.Stderr, "failed downloading t%d: %v\n", entry.TableID, err)
				os.Exit(1)
			}
			fmt.Printf("updated t%d.xml (log #%d on %s: %s)\n", entry.TableID, entry.TableIdentity, entry.SDate, entry.Action)
		}
		if logTs > maxLog {
			maxLog = logTs
		}
		processed++
	}

	state.LastLogMillis = maxLog
	if err := saveState(statePath, state); err != nil {
		fmt.Fprintf(os.Stderr, "failed to save state: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("processed %d entries\n", processed)
}

func (c changeEntry) logMillis() int64 {
	s := strings.TrimSpace(c.LogDate)
	s = strings.TrimPrefix(s, "/Date(")
	s = strings.TrimSuffix(s, ")/")
	ms, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return ms
}

func loadState(path string) (*stateFile, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &stateFile{LastLogMillis: 0}, nil
	}
	if err != nil {
		return nil, err
	}
	var state stateFile
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func saveState(path string, state *stateFile) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func fetchNewEntries(lastLogMillis int64) ([]changeEntry, error) {
	page := 1
	rows := 1000
	client := &http.Client{Timeout: 30 * time.Second}
	var collected []changeEntry
	shouldStop := false

	for !shouldStop {
		body := fmt.Sprintf(`{"page":%d,"rows":%d,"sidx":"sDate","sord":"desc"}`, page, rows)
		req, err := http.NewRequest(http.MethodPost, changeLogURL, strings.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status %s: %s", resp.Status, respBody)
		}

		var parsed changeLogResponse
		if err := json.Unmarshal(respBody, &parsed); err != nil {
			return nil, err
		}

		for _, entry := range parsed.D.Rows {
			if entry.logMillis() <= lastLogMillis {
				shouldStop = true
				break
			}
			collected = append(collected, entry)
		}

		if shouldStop || page >= parsed.D.Total {
			break
		}
		page++
	}

	return collected, nil
}

func downloadTable(tableID int, xmlDir string) error {
	url := fmt.Sprintf(xmlDownloadFmt, tableID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %s: %s", resp.Status, string(body))
	}

	fileName := fmt.Sprintf("t%d.xml", tableID)
	outPath := filepath.Join(xmlDir, fileName)
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}

	tmpFile, err := os.CreateTemp("", "mort-xml-*.xml")
	if err != nil {
		return err
	}
	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return err
	}

	if err := tmpFile.Sync(); err != nil {
		return err
	}

	if err := os.Rename(tmpFile.Name(), outPath); err != nil {
		return err
	}

	return nil
}

func removeTable(tableID int, xmlDir string) error {
	path := filepath.Join(xmlDir, fmt.Sprintf("t%d.xml", tableID))
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	return os.Remove(path)
}
