package idgen

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"
)

// --- UUID ---

// UUIDVersion specifies the UUID version to generate.
type UUIDVersion int

const (
	// V4 generates a random UUID (RFC 4122 version 4).
	V4 UUIDVersion = 4
	// V7 generates a time-ordered UUID (RFC 9562 version 7).
	V7 UUIDVersion = 7
)

// GenerateUUID generates a single UUID string with the given options.
func GenerateUUID(opts ...UUIDOption) (string, error) {
	cfg := uuidConfig{version: V4}
	for _, o := range opts {
		o(&cfg)
	}

	var (
		raw string
		err error
	)

	switch cfg.version {
	case V4:
		raw, err = generateUUIDv4()
	case V7:
		raw, err = generateUUIDv7()
	default:
		return "", fmt.Errorf("idgen: unsupported UUID version %d (use 4 or 7)", cfg.version)
	}

	if err != nil {
		return "", fmt.Errorf("idgen: %w", err)
	}

	if cfg.noDashes {
		raw = strings.ReplaceAll(raw, "-", "")
	}

	if cfg.uppercase {
		raw = strings.ToUpper(raw)
	}

	return raw, nil
}

// GenerateUUIDs generates n UUID strings with the given options.
func GenerateUUIDs(n int, opts ...UUIDOption) ([]string, error) {
	if n <= 0 {
		n = 1
	}

	result := make([]string, 0, n)
	for range n {
		u, err := GenerateUUID(opts...)
		if err != nil {
			return nil, err
		}

		result = append(result, u)
	}

	return result, nil
}

// IsValidUUID checks if a string is a valid UUID format.
func IsValidUUID(s string) bool {
	s = strings.ReplaceAll(s, "-", "")
	if len(s) != 32 {
		return false
	}

	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
			return false
		}
	}

	return true
}

type uuidConfig struct {
	version   UUIDVersion
	uppercase bool
	noDashes  bool
}

// UUIDOption configures UUID generation.
type UUIDOption func(*uuidConfig)

// WithUUIDVersion sets the UUID version (4 or 7).
func WithUUIDVersion(v UUIDVersion) UUIDOption {
	return func(c *uuidConfig) { c.version = v }
}

// WithUppercase outputs the UUID in uppercase.
func WithUppercase() UUIDOption {
	return func(c *uuidConfig) { c.uppercase = true }
}

// WithNoDashes outputs the UUID without dashes.
func WithNoDashes() UUIDOption {
	return func(c *uuidConfig) { c.noDashes = true }
}

func generateUUIDv4() (string, error) {
	uuid := make([]byte, 16)

	_, err := rand.Read(uuid)
	if err != nil {
		return "", err
	}

	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant RFC 4122

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16]), nil
}

func generateUUIDv7() (string, error) {
	uuid := make([]byte, 16)
	now := time.Now().UnixMilli()

	uuid[0] = byte(now >> 40)
	uuid[1] = byte(now >> 32)
	uuid[2] = byte(now >> 24)
	uuid[3] = byte(now >> 16)
	uuid[4] = byte(now >> 8)
	uuid[5] = byte(now)

	_, err := rand.Read(uuid[6:])
	if err != nil {
		return "", err
	}

	uuid[6] = (uuid[6] & 0x0f) | 0x70 // Version 7
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant RFC 4122

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16]), nil
}

// --- ULID ---

const (
	ulidEncodedSize   = 26
	ulidTimestampSize = 6
)

var crockfordAlphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

// ULID represents a Universally Unique Lexicographically Sortable Identifier.
type ULID [16]byte

// GenerateULID generates a new ULID.
func GenerateULID() (ULID, error) {
	return GenerateULIDWithTime(time.Now())
}

// GenerateULIDWithTime generates a new ULID with the given timestamp.
func GenerateULIDWithTime(t time.Time) (ULID, error) {
	var u ULID

	ms := uint64(t.UnixMilli())

	u[0] = byte(ms >> 40)
	u[1] = byte(ms >> 32)
	u[2] = byte(ms >> 24)
	u[3] = byte(ms >> 16)
	u[4] = byte(ms >> 8)
	u[5] = byte(ms)

	_, err := rand.Read(u[ulidTimestampSize:])
	if err != nil {
		return u, err
	}

	return u, nil
}

// String returns the Crockford's Base32 encoded ULID.
func (u ULID) String() string {
	r := make([]byte, ulidEncodedSize)

	r[0] = crockfordAlphabet[(u[0]&224)>>5]
	r[1] = crockfordAlphabet[u[0]&31]
	r[2] = crockfordAlphabet[(u[1]&248)>>3]
	r[3] = crockfordAlphabet[((u[1]&7)<<2)|((u[2]&192)>>6)]
	r[4] = crockfordAlphabet[(u[2]&62)>>1]
	r[5] = crockfordAlphabet[((u[2]&1)<<4)|((u[3]&240)>>4)]
	r[6] = crockfordAlphabet[((u[3]&15)<<1)|((u[4]&128)>>7)]
	r[7] = crockfordAlphabet[(u[4]&124)>>2]
	r[8] = crockfordAlphabet[((u[4]&3)<<3)|((u[5]&224)>>5)]
	r[9] = crockfordAlphabet[u[5]&31]

	r[10] = crockfordAlphabet[(u[6]&248)>>3]
	r[11] = crockfordAlphabet[((u[6]&7)<<2)|((u[7]&192)>>6)]
	r[12] = crockfordAlphabet[(u[7]&62)>>1]
	r[13] = crockfordAlphabet[((u[7]&1)<<4)|((u[8]&240)>>4)]
	r[14] = crockfordAlphabet[((u[8]&15)<<1)|((u[9]&128)>>7)]
	r[15] = crockfordAlphabet[(u[9]&124)>>2]
	r[16] = crockfordAlphabet[((u[9]&3)<<3)|((u[10]&224)>>5)]
	r[17] = crockfordAlphabet[u[10]&31]
	r[18] = crockfordAlphabet[(u[11]&248)>>3]
	r[19] = crockfordAlphabet[((u[11]&7)<<2)|((u[12]&192)>>6)]
	r[20] = crockfordAlphabet[(u[12]&62)>>1]
	r[21] = crockfordAlphabet[((u[12]&1)<<4)|((u[13]&240)>>4)]
	r[22] = crockfordAlphabet[((u[13]&15)<<1)|((u[14]&128)>>7)]
	r[23] = crockfordAlphabet[(u[14]&124)>>2]
	r[24] = crockfordAlphabet[((u[14]&3)<<3)|((u[15]&224)>>5)]
	r[25] = crockfordAlphabet[u[15]&31]

	return string(r)
}

// Timestamp returns the time the ULID was created.
func (u ULID) Timestamp() time.Time {
	ms := uint64(u[0])<<40 | uint64(u[1])<<32 | uint64(u[2])<<24 |
		uint64(u[3])<<16 | uint64(u[4])<<8 | uint64(u[5])

	return time.UnixMilli(int64(ms))
}

// ULIDString generates a new ULID and returns it as a string.
func ULIDString() string {
	u, err := GenerateULID()
	if err != nil {
		return ""
	}

	return u.String()
}

// --- KSUID ---

const (
	ksuidEpoch        = 1400000000
	ksuidPayloadSize  = 16
	ksuidTimestampLen = 4
	ksuidTotalSize    = ksuidTimestampLen + ksuidPayloadSize
	ksuidEncodedSize  = 27
)

// KSUID represents a K-Sortable Unique IDentifier.
type KSUID [ksuidTotalSize]byte

// GenerateKSUID generates a new KSUID.
func GenerateKSUID() (KSUID, error) {
	var k KSUID

	timestamp := uint32(time.Now().Unix() - ksuidEpoch)

	k[0] = byte(timestamp >> 24)
	k[1] = byte(timestamp >> 16)
	k[2] = byte(timestamp >> 8)
	k[3] = byte(timestamp)

	_, err := rand.Read(k[ksuidTimestampLen:])
	if err != nil {
		return k, err
	}

	return k, nil
}

// String returns the base62 encoded KSUID.
func (k KSUID) String() string {
	return base62Encode(k[:])
}

// Timestamp returns the time the KSUID was created.
func (k KSUID) Timestamp() time.Time {
	ts := uint32(k[0])<<24 | uint32(k[1])<<16 | uint32(k[2])<<8 | uint32(k[3])
	return time.Unix(int64(ts)+ksuidEpoch, 0)
}

// KSUIDString generates a new KSUID and returns it as a string.
func KSUIDString() string {
	k, err := GenerateKSUID()
	if err != nil {
		return ""
	}

	return k.String()
}

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func base62Encode(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	result := make([]byte, ksuidEncodedSize)
	for i := range result {
		result[i] = 0
	}

	for _, b := range data {
		carry := int(b)
		for j := ksuidEncodedSize - 1; j >= 0; j-- {
			carry += 256 * int(result[j])
			result[j] = byte(carry % 62)
			carry /= 62
		}
	}

	for i := range result {
		result[i] = base62Chars[result[i]]
	}

	return string(result)
}

// --- NanoID ---

const (
	defaultNanoidAlphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz-"
	defaultNanoidLength   = 21
)

// NanoidOption configures NanoID generation.
type NanoidOption func(*nanoidConfig)

type nanoidConfig struct {
	length   int
	alphabet string
}

// WithNanoidLength sets the NanoID length (default 21).
func WithNanoidLength(n int) NanoidOption {
	return func(c *nanoidConfig) { c.length = n }
}

// WithNanoidAlphabet sets a custom alphabet for NanoID generation.
func WithNanoidAlphabet(a string) NanoidOption {
	return func(c *nanoidConfig) { c.alphabet = a }
}

// GenerateNanoid creates a NanoID with the given options.
func GenerateNanoid(opts ...NanoidOption) (string, error) {
	cfg := nanoidConfig{
		length:   defaultNanoidLength,
		alphabet: defaultNanoidAlphabet,
	}
	for _, o := range opts {
		o(&cfg)
	}

	if cfg.length <= 0 {
		cfg.length = defaultNanoidLength
	}

	if cfg.alphabet == "" {
		cfg.alphabet = defaultNanoidAlphabet
	}

	alphabetLen := big.NewInt(int64(len(cfg.alphabet)))
	result := make([]byte, cfg.length)

	for i := 0; i < cfg.length; i++ {
		idx, err := rand.Int(rand.Reader, alphabetLen)
		if err != nil {
			return "", err
		}

		result[i] = cfg.alphabet[idx.Int64()]
	}

	return string(result), nil
}

// NanoidString generates a new NanoID with default settings and returns it as a string.
func NanoidString() string {
	s, err := GenerateNanoid()
	if err != nil {
		return ""
	}

	return s
}

// --- Snowflake ---

const (
	snowflakeEpoch = 1577836800000 // Jan 01, 2020 UTC in ms

	snowflakeTimestampBits = 41
	snowflakeWorkerIDBits  = 10
	snowflakeSequenceBits  = 12

	snowflakeMaxWorkerID = (1 << snowflakeWorkerIDBits) - 1
	snowflakeMaxSequence = (1 << snowflakeSequenceBits) - 1

	snowflakeWorkerIDShift  = snowflakeSequenceBits
	snowflakeTimestampShift = snowflakeSequenceBits + snowflakeWorkerIDBits
)

// SnowflakeGenerator generates Snowflake IDs.
type SnowflakeGenerator struct {
	mu       sync.Mutex
	workerID int64
	sequence int64
	lastTime int64
}

// NewSnowflakeGenerator creates a new Snowflake generator with the given worker ID (0-1023).
func NewSnowflakeGenerator(workerID int64) *SnowflakeGenerator {
	return &SnowflakeGenerator{
		workerID: workerID & snowflakeMaxWorkerID,
	}
}

// Generate creates a new Snowflake ID.
func (g *SnowflakeGenerator) Generate() (int64, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now().UnixMilli() - snowflakeEpoch

	if now < g.lastTime {
		return 0, fmt.Errorf("clock moved backwards")
	}

	if now == g.lastTime {
		g.sequence = (g.sequence + 1) & snowflakeMaxSequence
		if g.sequence == 0 {
			for now <= g.lastTime {
				now = time.Now().UnixMilli() - snowflakeEpoch
			}
		}
	} else {
		g.sequence = 0
	}

	g.lastTime = now

	id := (now << snowflakeTimestampShift) |
		(g.workerID << snowflakeWorkerIDShift) |
		g.sequence

	return id, nil
}

var (
	defaultSnowflakeGen *SnowflakeGenerator
	snowflakeOnce       sync.Once
)

// GenerateSnowflake generates a new Snowflake ID using the default generator (worker 0).
func GenerateSnowflake() (int64, error) {
	snowflakeOnce.Do(func() {
		defaultSnowflakeGen = NewSnowflakeGenerator(0)
	})

	return defaultSnowflakeGen.Generate()
}

// SnowflakeString generates a new Snowflake ID and returns it as a string.
func SnowflakeString() string {
	id, err := GenerateSnowflake()
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%d", id)
}

// ParseSnowflake extracts components from a Snowflake ID.
func ParseSnowflake(id int64) (timestamp time.Time, workerID int64, sequence int64) {
	timestamp = time.UnixMilli((id >> snowflakeTimestampShift) + snowflakeEpoch)
	workerID = (id >> snowflakeWorkerIDShift) & snowflakeMaxWorkerID
	sequence = id & snowflakeMaxSequence

	return
}
