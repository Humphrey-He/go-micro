package idgen

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

// UUID generates a standard UUID v4.
func UUID() string {
	return uuid.New().String()
}

// UUIDNoDash generates a UUID without dashes.
func UUIDNoDash() string {
	return uuid.New().String()
}

// ShortID generates a short ID based on timestamp and random bytes.
func ShortID() string {
	now := time.Now().UnixNano()
	return FormatNanoID(now)
}

// FormatNanoID formats a nano timestamp as a short ID.
func FormatNanoID(timestamp int64) string {
	const chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	b := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		b[i] = chars[timestamp%62]
		timestamp /= 62
	}
	return string(b)
}

// Snowflake generates a snowflake ID (Twitter's distributed ID algorithm).
type Snowflake struct {
	mu        sync.Mutex
	timestamp int64
	nodeID    int64
	sequence  int64
}

const (
	epoch          = int64(1609459200000) // 2021-01-01 00:00:00 UTC
	nodeIDBits     = uint(10)
	sequenceBits   = uint(12)
	nodeIDShift    = sequenceBits
	timestampShift = nodeIDBits + sequenceBits
	sequenceMask   = int64(-1) ^ (int64(-1) << sequenceBits)
	maxNodeID      = int64(-1) ^ (int64(-1) << nodeIDBits)
)

// NewSnowflake creates a new Snowflake ID generator.
func NewSnowflake(nodeID int64) (*Snowflake, error) {
	if nodeID < 0 || nodeID > maxNodeID {
		return nil, errors.New("node ID must be between 0 and 1023")
	}
	return &Snowflake{
		nodeID: nodeID,
	}, nil
}

// Generate generates a new snowflake ID.
func (s *Snowflake) Generate() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixNano()/1000000 - epoch
	if now == s.timestamp {
		s.sequence = (s.sequence + 1) & sequenceMask
		if s.sequence == 0 {
			for now == s.timestamp {
				now = time.Now().UnixNano()/1000000 - epoch
			}
		}
	} else {
		s.sequence = 0
	}
	s.timestamp = now
	return (now << timestampShift) | (s.nodeID << nodeIDShift) | s.sequence
}
