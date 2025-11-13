package tuiapp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

// TableSummary is the minimal data needed to list and search tables.
type TableSummary struct {
	TableIdentity string
	Identifier    string
	Name          string
	Provider      string
	Summary       string
	FilePath      string
	Keywords      []string
}

type convertedTable struct {
	Identifier     string `json:"identifier"`
	Classification struct {
		TableIdentity    string   `json:"tableIdentity"`
		TableName        string   `json:"tableName"`
		ProviderName     string   `json:"providerName"`
		TableDescription string   `json:"tableDescription"`
		Keywords         []string `json:"keywords"`
	} `json:"classification"`
}

// LoadTableSummaries reads all *.json files in dir and returns sorted summaries.
func LoadTableSummaries(dir string) ([]TableSummary, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read json dir: %w", err)
	}

	var summaries []TableSummary
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}
		path := filepath.Join(dir, file.Name())
		summary, err := LoadTableSummary(path)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, *summary)
	}

	sort.Slice(summaries, func(i, j int) bool {
		ai, _ := strconv.Atoi(summaries[i].TableIdentity)
		bi, _ := strconv.Atoi(summaries[j].TableIdentity)
		if ai != bi {
			return ai < bi
		}
		return summaries[i].Name < summaries[j].Name
	})

	return summaries, nil
}

// LoadTableSummary loads a single table summary from a JSON file.
func LoadTableSummary(path string) (*TableSummary, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var ct convertedTable
	if err := json.Unmarshal(raw, &ct); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	return &TableSummary{
		TableIdentity: ct.Classification.TableIdentity,
		Identifier:    ct.Identifier,
		Name:          ct.Classification.TableName,
		Provider:      ct.Classification.ProviderName,
		Summary:       ct.Classification.TableDescription,
		FilePath:      path,
		Keywords:      ct.Classification.Keywords,
	}, nil
}

// FilterSummaries performs a fuzzy search over names and identifiers.
func FilterSummaries(items []TableSummary, query string) []TableSummary {
	query = strings.TrimSpace(query)
	if query == "" {
		return append([]TableSummary(nil), items...)
	}

	lowerQuery := strings.ToLower(query)
	tokens := strings.Fields(lowerQuery)

	var (
		filteredIdx []int
		choices     []string
	)
	for i, item := range items {
		extra := strings.Join(item.Keywords, " ")
		body := strings.ToLower(strings.Join([]string{
			item.Name,
			item.Identifier,
			item.TableIdentity,
			item.Provider,
			item.Summary,
			extra,
		}, " "))

		matchesAll := true
		for _, token := range tokens {
			if !strings.Contains(body, token) {
				matchesAll = false
				break
			}
		}
		if !matchesAll {
			continue
		}

		filteredIdx = append(filteredIdx, i)
		choices = append(choices, body)
	}

	if len(filteredIdx) == 0 {
		return nil
	}

	matches := fuzzy.RankFindNormalizedFold(lowerQuery, choices)
	if len(matches) == 0 {
		// fallback to filtered order if fuzzy produced nothing
		result := make([]TableSummary, len(filteredIdx))
		for i, idx := range filteredIdx {
			result[i] = items[idx]
		}
		return result
	}
	sort.SliceStable(matches, func(i, j int) bool {
		return matches[i].Distance < matches[j].Distance
	})
	const maxResults = 100
	result := make([]TableSummary, 0, min(len(matches), maxResults))
	for i, match := range matches {
		if i >= maxResults {
			break
		}
		result = append(result, items[filteredIdx[match.OriginalIndex]])
	}
	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
