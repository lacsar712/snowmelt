package store

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type JournalEntry struct {
	UnitID    string    `json:"unit_id"`
	Event     string    `json:"event"`
	Payload   string    `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
}

type Journal struct {
	mu       sync.Mutex
	path     string
	capacity int
	entries  []JournalEntry
}

func NewJournal(path string, capacity int) *Journal {
	if capacity <= 0 {
		capacity = 256
	}
	return &Journal{path: path, capacity: capacity, entries: make([]JournalEntry, 0, capacity)}
}

func (j *Journal) Append(unitID, event, payload string) (JournalEntry, error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	entry := JournalEntry{
		UnitID: unitID, Event: event, Payload: payload, Timestamp: time.Now(),
	}
	j.entries = append(j.entries, entry)
	if len(j.entries) > j.capacity {
		j.entries = j.entries[len(j.entries)-j.capacity:]
	}
	if j.path != "" {
		if err := j.persistEntry(entry); err != nil {
			return entry, err
		}
	}
	return entry, nil
}

func (j *Journal) persistEntry(entry JournalEntry) error {
	if err := os.MkdirAll(filepath.Dir(j.path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(j.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	_, err = f.Write(append(data, '\n'))
	return err
}

func (j *Journal) Recent(n int) []JournalEntry {
	j.mu.Lock()
	defer j.mu.Unlock()
	if n <= 0 || n > len(j.entries) {
		n = len(j.entries)
	}
	start := len(j.entries) - n
	out := make([]JournalEntry, n)
	copy(out, j.entries[start:])
	return out
}

func (j *Journal) Load() error {
	if j.path == "" {
		return nil
	}
	f, err := os.Open(j.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()
	j.mu.Lock()
	defer j.mu.Unlock()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var entry JournalEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}
		j.entries = append(j.entries, entry)
	}
	if len(j.entries) > j.capacity {
		j.entries = j.entries[len(j.entries)-j.capacity:]
	}
	return scanner.Err()
}

func (j *Journal) Export(path string) error {
	j.mu.Lock()
	defer j.mu.Unlock()
	data, err := json.MarshalIndent(j.entries, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func (j *Journal) Count() int {
	j.mu.Lock()
	defer j.mu.Unlock()
	return len(j.entries)
}

func (j *Journal) Path() string { return j.path }

func FormatJournalPayload(fields map[string]string) string {
	data, _ := json.Marshal(fields)
	return string(data)
}

func ParseJournalPayload(payload string) (map[string]string, error) {
	var fields map[string]string
	if err := json.Unmarshal([]byte(payload), &fields); err != nil {
		return nil, fmt.Errorf("parse journal payload: %w", err)
	}
	return fields, nil
}
