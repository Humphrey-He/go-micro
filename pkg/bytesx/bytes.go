package bytesx

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"strconv"
)

const (
	KB = 1024
	MB = 1024 * KB
	GB = 1024 * MB
	TB = 1024 * GB
)

// Size represents a byte size with formatting methods.
type Size int64

// Bytes returns the size in bytes.
func (s Size) Bytes() int64 {
	return int64(s)
}

// Kilobytes returns the size in kilobytes.
func (s Size) Kilobytes() float64 {
	return float64(s) / KB
}

// Megabytes returns the size in megabytes.
func (s Size) Megabytes() float64 {
	return float64(s) / MB
}

// Gigabytes returns the size in gigabytes.
func (s Size) Gigabytes() float64 {
	return float64(s) / GB
}

// Terabytes returns the size in terabytes.
func (s Size) Terabytes() float64 {
	return float64(s) / TB
}

// String returns the size as a human-readable string.
func (s Size) String() string {
	switch {
	case s >= TB:
		return strconv.FormatFloat(float64(s)/TB, 'f', 2, 64) + " TB"
	case s >= GB:
		return strconv.FormatFloat(float64(s)/GB, 'f', 2, 64) + " GB"
	case s >= MB:
		return strconv.FormatFloat(float64(s)/MB, 'f', 2, 64) + " MB"
	case s >= KB:
		return strconv.FormatFloat(float64(s)/KB, 'f', 2, 64) + " KB"
	default:
		return strconv.FormatInt(int64(s), 10) + " B"
	}
}

// ToHex converts bytes to hex string.
func ToHex(b []byte) string {
	return hex.EncodeToString(b)
}

// FromHex converts hex string to bytes.
func FromHex(s string) ([]byte, error) {
	return hex.DecodeString(s)
}

// Concat concatenates multiple byte slices.
func Concat(b ...[]byte) []byte {
	return bytes.Join(b, nil)
}

// Contains checks if the slice contains the byte slice.
func Contains(s, sub []byte) bool {
	return bytes.Contains(s, sub)
}

// Count counts the number of occurrences of sep in s.
func Count(s, sep []byte) int {
	return bytes.Count(s, sep)
}

// Equal checks if two byte slices are equal.
func Equal(a, b []byte) bool {
	return bytes.Equal(a, b)
}

// HasPrefix checks if the slice has the prefix.
func HasPrefix(s, prefix []byte) bool {
	return bytes.HasPrefix(s, prefix)
}

// HasSuffix checks if the slice has the suffix.
func HasSuffix(s, suffix []byte) bool {
	return bytes.HasSuffix(s, suffix)
}

// Index returns the index of the first occurrence of sep in s.
func Index(s, sep []byte) int {
	return bytes.Index(s, sep)
}

// LastIndex returns the index of the last occurrence of sep in s.
func LastIndex(s, sep []byte) int {
	return bytes.LastIndex(s, sep)
}

// Split splits the slice around the separator.
func Split(s, sep []byte) [][]byte {
	return bytes.Split(s, sep)
}

// SplitN splits the slice around the separator with a limit.
func SplitN(s, sep []byte, n int) [][]byte {
	return bytes.SplitN(s, sep, n)
}

// Trim removes leading and trailing bytes.
func Trim(s, cutset []byte) []byte {
	return bytes.Trim(s, string(cutset))
}

// TrimLeft removes leading bytes.
func TrimLeft(s, cutset []byte) []byte {
	return bytes.TrimLeft(s, string(cutset))
}

// TrimRight removes trailing bytes.
func TrimRight(s, cutset []byte) []byte {
	return bytes.TrimRight(s, string(cutset))
}

// TrimSpace removes leading and trailing white space.
func TrimSpace(s []byte) []byte {
	return bytes.TrimSpace(s)
}

// ToString converts bytes to string.
func ToString(b []byte) string {
	return string(b)
}

// FromString converts string to bytes.
func FromString(s string) []byte {
	return []byte(s)
}

// ReadUint16 reads a uint16 in big-endian order.
func ReadUint16(b []byte) (uint16, error) {
	r := bytes.NewReader(b)
	var v uint16
	err := binary.Read(r, binary.BigEndian, &v)
	return v, err
}

// ReadUint32 reads a uint32 in big-endian order.
func ReadUint32(b []byte) (uint32, error) {
	r := bytes.NewReader(b)
	var v uint32
	err := binary.Read(r, binary.BigEndian, &v)
	return v, err
}

// ReadUint64 reads a uint64 in big-endian order.
func ReadUint64(b []byte) (uint64, error) {
	r := bytes.NewReader(b)
	var v uint64
	err := binary.Read(r, binary.BigEndian, &v)
	return v, err
}

// WriteUint16 writes a uint16 in big-endian order.
func WriteUint16(v uint16) []byte {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, v)
	return buf
}

// WriteUint32 writes a uint32 in big-endian order.
func WriteUint32(v uint32) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, v)
	return buf
}

// WriteUint64 writes a uint64 in big-endian order.
func WriteUint64(v uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, v)
	return buf
}

// Grow grows the slice capacity.
func Grow(b []byte, n int) []byte {
	return append(b, make([]byte, n)...)[:len(b)]
}

// Clone creates a copy of the slice.
func Clone(b []byte) []byte {
	return bytes.Clone(b)
}

// Compare compares two slices lexicographically.
func Compare(a, b []byte) int {
	return bytes.Compare(a, b)
}
