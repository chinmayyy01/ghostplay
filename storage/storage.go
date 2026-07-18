package storage

import (
	"bufio"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Record struct {
	ID          string      `json:"id"`
	Timestamp   time.Time   `json:"timestamp"`
	Method      string      `json:"method"`
	Path        string      `json:"path"`
	Query       string      `json:"query"`
	ReqHeaders  http.Header `json:"req_headers"`
	ReqBody     string      `json:"req_body"`
	RespStatus  int         `json:"resp_status"`
	RespHeaders http.Header `json:"resp_headers"`
	RespBody    string      `json:"resp_body"`
	DurationMs  int64       `json:"duration_ms"`
}

type Store struct {
	mu   sync.Mutex
	file *os.File
	path string
}

func NewStore(path string) (*Store, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
 
	return &Store{file: f}, nil
}

func (s *Store) Save(r Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.Marshal(r)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	_, err = s.file.Write(data)
	return err
}

func genID() string {
	return formatInt36(time.Now().UnixNano())
}

func formatInt36(n int64) string {
	const digits = "0123456789abcdefghijklmnopqrstuvwxyz"
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 16)
	for n > 0 {
		buf = append([]byte{digits[n%36]}, buf...)
		n /= 36
	}
	return string(buf)
}

func (s *Store) All() ([]Record, error) {
	f, err := os.Open(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Record{}, nil
		}
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
 
	var records []Record
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var r Record
		if err := json.Unmarshal(line, &r); err != nil {
			continue
		}
		records = append(records, r)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	for i, j := 0, len(records)-1; i < j; i, j = i+1, j-1 {
		records[i], records[j] = records[j], records[i]
	}
 
	return records, nil
}

func (s *Store) Get(id string) (*Record, error) {
	records, err := s.All()
	if err != nil {
		return nil, err
	}
	for _, r := range records {
		if r.ID == id {
			return &r, nil
		}
	}
	return nil, nil
}

func NewRecord(method, path, query string, reqHeaders http.Header, reqBody string, respStatus int, respHeaders http.Header, respBody string, durationMs int64) Record {
	return Record{
		ID:          genID(),
		Timestamp:   time.Now(),
		Method:      method,
		Path:        path,
		Query:       query,
		ReqHeaders:  reqHeaders,
		ReqBody:     reqBody,
		RespStatus:  respStatus,
		RespHeaders: respHeaders,
		RespBody:    respBody,
		DurationMs:  durationMs,
	}
}