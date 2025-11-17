package model

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

type Address struct {
	City    string `json:"city"`
	Country string `json:"country"`
	Street  string `json:"street"`
	ZipCode string `json:"zip_code"`
}

type Event struct {
	Type      string `json:"type"`
	Timestamp int64  `json:"timestamp"`
	Metadata  string `json:"metadata"` // JSON-encoded string for extra data
}

type SessionRecord struct {
	ID          string   `json:"id"`
	UserID      string   `json:"user_id"`
	UserName    string   `json:"user_name"`
	Email       string   `json:"email"`
	Device      string   `json:"device"`
	IP          string   `json:"ip"`
	IsMobile    bool     `json:"is_mobile"`
	IsActive    bool     `json:"is_active"`
	CreatedAt   int64    `json:"created_at"`
	LastSeenAt  int64    `json:"last_seen_at"`
	Tags        []string `json:"tags"`
	Address     Address  `json:"address"`
	Events      []Event  `json:"events"`
	SessionNote string   `json:"session_note"`
}

func GenerateSessionRecords(n int) []SessionRecord {
	records := make([]SessionRecord, n)
	now := time.Now().Unix()
	r := rand.New(rand.NewSource(42))

	for i := 0; i < n; i++ {
		id := fmt.Sprintf("session-%d", i)

		addr := Address{
			City:    fmt.Sprintf("City-%d", i%50),
			Country: "Wonderland",
			Street:  fmt.Sprintf("Street %d", i),
			ZipCode: fmt.Sprintf("%05d", i%10000),
		}

		events := make([]Event, 0, 10)
		for j := 0; j < 10; j++ {
			ev := Event{
				Type:      fmt.Sprintf("event_type_%d", j%3),
				Timestamp: now - int64(j*60),
				Metadata: fmt.Sprintf(`{"action":"click","index":%d,"score":%d}`,
					j, r.Intn(1000)),
			}
			events = append(events, ev)
		}

		record := SessionRecord{
			ID:         id,
			UserID:     fmt.Sprintf("user-%d", i%200),
			UserName:   fmt.Sprintf("User Name %d", i),
			Email:      fmt.Sprintf("user%d@example.com", i),
			Device:     []string{"android", "ios", "web", "desktop"}[i%4],
			IP:         fmt.Sprintf("10.0.%d.%d", (i/256)%256, i%256),
			IsMobile:   i%2 == 0,
			IsActive:   i%3 != 0,
			CreatedAt:  now - int64(r.Intn(3600*24)),
			LastSeenAt: now - int64(r.Intn(3600)),
			Tags:       []string{"beta", "test", fmt.Sprintf("cohort-%d", i%5)},
			Address:    addr,
			Events:     events,
			SessionNote: fmt.Sprintf(
				"Some longer note about this session #%d with random value %d",
				i, r.Intn(1_000_000),
			),
		}

		records[i] = record
	}
	return records
}

// ToHash: used for HSET
func (s SessionRecord) ToHash() (map[string]interface{}, error) {
	m := make(map[string]interface{})

	m["id"] = s.ID
	m["user_id"] = s.UserID
	m["user_name"] = s.UserName
	m["email"] = s.Email
	m["device"] = s.Device
	m["ip"] = s.IP
	m["is_mobile"] = s.IsMobile
	m["is_active"] = s.IsActive
	m["created_at"] = s.CreatedAt
	m["last_seen_at"] = s.LastSeenAt
	m["session_note"] = s.SessionNote

	tagsJSON, err := json.Marshal(s.Tags)
	if err != nil {
		return nil, fmt.Errorf("marshal tags: %w", err)
	}
	m["tags"] = string(tagsJSON)

	addrJSON, err := json.Marshal(s.Address)
	if err != nil {
		return nil, fmt.Errorf("marshal address: %w", err)
	}
	m["address"] = string(addrJSON)

	eventsJSON, err := json.Marshal(s.Events)
	if err != nil {
		return nil, fmt.Errorf("marshal events: %w", err)
	}
	m["events"] = string(eventsJSON)

	return m, nil
}

// FromHash: used after HGETALL
func FromHash(m map[string]string) (SessionRecord, error) {
	var rec SessionRecord
	var err error

	rec.ID = m["id"]
	rec.UserID = m["user_id"]
	rec.UserName = m["user_name"]
	rec.Email = m["email"]
	rec.Device = m["device"]
	rec.IP = m["ip"]
	rec.SessionNote = m["session_note"]

	if v, ok := m["is_mobile"]; ok {
		rec.IsMobile, err = strconv.ParseBool(v)
		if err != nil {
			return SessionRecord{}, fmt.Errorf("parse is_mobile: %w", err)
		}
	}

	if v, ok := m["is_active"]; ok {
		rec.IsActive, err = strconv.ParseBool(v)
		if err != nil {
			return SessionRecord{}, fmt.Errorf("parse is_active: %w", err)
		}
	}

	if v, ok := m["created_at"]; ok {
		rec.CreatedAt, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			return SessionRecord{}, fmt.Errorf("parse created_at: %w", err)
		}
	}

	if v, ok := m["last_seen_at"]; ok {
		rec.LastSeenAt, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			return SessionRecord{}, fmt.Errorf("parse last_seen_at: %w", err)
		}
	}

	if v, ok := m["tags"]; ok && v != "" {
		if err := json.Unmarshal([]byte(v), &rec.Tags); err != nil {
			return SessionRecord{}, fmt.Errorf("unmarshal tags: %w", err)
		}
	}

	if v, ok := m["address"]; ok && v != "" {
		if err := json.Unmarshal([]byte(v), &rec.Address); err != nil {
			return SessionRecord{}, fmt.Errorf("unmarshal address: %w", err)
		}
	}

	if v, ok := m["events"]; ok && v != "" {
		if err := json.Unmarshal([]byte(v), &rec.Events); err != nil {
			return SessionRecord{}, fmt.Errorf("unmarshal events: %w", err)
		}
	}

	return rec, nil
}
