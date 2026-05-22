package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type FileStore struct {
	mu   sync.Mutex
	root string
}

type indexEntry struct {
	Date    time.Time `json:"date"`
	Agent   string    `json:"agent"`
	Session string    `json:"session"`
	Tags    []string  `json:"tags"`
	File    string    `json:"file"`
}

type memoryIndex struct {
	Entries []indexEntry `json:"entries"`
}

func NewFileStore(rootDir string) (*FileStore, error) {
	if rootDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("obteniendo home dir: %w", err)
		}
		rootDir = filepath.Join(home, ".spider", "memory")
	}
	if err := os.MkdirAll(rootDir, 0755); err != nil {
		return nil, fmt.Errorf("creando directorio de memoria: %w", err)
	}
	return &FileStore{root: rootDir}, nil
}

func (s *FileStore) Save(entry *SummaryEntry, agentName string, tags []string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entry == nil {
		return "", nil
	}

	dateStr := entry.Date.Format("2006-01-02")
	sessionShort := entry.SessionID
	if len(sessionShort) > 8 {
		sessionShort = sessionShort[:8]
	}
	filename := fmt.Sprintf("%s_%s_%s.md", dateStr, agentName, sessionShort)
	filePath := filepath.Join(s.root, filename)

	entry.AgentName = agentName
	if tags != nil {
		entry.Tags = tags
	}

	var tagsYAML string
	if len(entry.Tags) > 0 {
		tagsYAML = "tags: [" + strings.Join(entry.Tags, ", ") + "]"
	}

	content := fmt.Sprintf(`---
date: %s
agent: %s
session: %s
%s
tokens_compressed: %d
---

## Resumen de sesión

%s
`,
		entry.Date.Format(time.RFC3339),
		entry.AgentName,
		entry.SessionID,
		tagsYAML,
		entry.TokensSaved,
		entry.Summary,
	)

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("guardando memoria: %w", err)
	}

	if err := s.updateIndex(indexEntry{
		Date:    entry.Date,
		Agent:   entry.AgentName,
		Session: entry.SessionID,
		Tags:    entry.Tags,
		File:    filename,
	}); err != nil {
		return "", err
	}

	return filePath, nil
}

func (s *FileStore) updateIndex(e indexEntry) error {
	idx := s.loadIndex()
	idx.Entries = append(idx.Entries, e)
	return s.writeIndex(idx)
}

func (s *FileStore) loadIndex() memoryIndex {
	idxPath := filepath.Join(s.root, "index.json")
	data, err := os.ReadFile(idxPath)
	if err != nil {
		return memoryIndex{}
	}
	var idx memoryIndex
	json.Unmarshal(data, &idx)
	return idx
}

func (s *FileStore) writeIndex(idx memoryIndex) error {
	idxPath := filepath.Join(s.root, "index.json")
	sort.Slice(idx.Entries, func(i, j int) bool {
		return idx.Entries[i].Date.After(idx.Entries[j].Date)
	})
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal index: %w", err)
	}
	return os.WriteFile(idxPath, data, 0644)
}

func (s *FileStore) Search(query string, tags []string, maxResults int) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	idx := s.loadIndex()
	if maxResults <= 0 {
		maxResults = 5
	}

	var candidates []indexEntry
	for _, e := range idx.Entries {
		if len(tags) > 0 {
			match := false
			for _, t := range tags {
				for _, et := range e.Tags {
					if strings.EqualFold(t, et) {
						match = true
						break
					}
				}
				if match {
					break
				}
			}
			if !match {
				continue
			}
		}
		candidates = append(candidates, e)
	}

	if len(candidates) > maxResults {
		candidates = candidates[:maxResults]
	}

	if len(candidates) == 0 {
		return "No se encontraron memorias relevantes.", nil
	}

	var results strings.Builder
	if query != "" {
		fmt.Fprintf(&results, "Búsqueda: %s\n\n", query)
	}

	for _, c := range candidates {
		filePath := filepath.Join(s.root, c.File)
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		body := string(data)
		if query != "" {
			bodyLC := strings.ToLower(body)
			queryLC := strings.ToLower(query)
			words := strings.Fields(queryLC)
			match := true
			for _, w := range words {
				if !strings.Contains(bodyLC, w) {
					match = false
					break
				}
			}
			if !match {
				continue
			}
		}

		summary := extractSummary(body, 1000)

		fmt.Fprintf(&results, "### %s (%s)\n", c.File, c.Date.Format("2006-01-02 15:04"))
		if len(c.Tags) > 0 {
			fmt.Fprintf(&results, "Tags: %s\n\n", strings.Join(c.Tags, ", "))
		}
		fmt.Fprintf(&results, "%s\n\n---\n\n", summary)
	}

	return results.String(), nil
}

func (s *FileStore) Root() string {
	return s.root
}

func (s *FileStore) List() ([]indexEntry, error) {
	idx := s.loadIndex()
	return idx.Entries, nil
}

func extractSummary(content string, maxLen int) string {
	parts := strings.SplitN(content, "## Resumen de sesión", 2)
	if len(parts) < 2 {
		parts = strings.SplitN(content, "\n\n", 2)
		if len(parts) < 2 {
			if len(content) > maxLen {
				return content[:maxLen] + "..."
			}
			return content
		}
	}

	summary := strings.TrimSpace(parts[1])
	if len([]rune(summary)) > maxLen {
		return string([]rune(summary)[:maxLen]) + "..."
	}
	return summary
}
