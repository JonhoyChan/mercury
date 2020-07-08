package ksuid

import (
	"bytes"
	"crypto/rand"
	"database/sql/driver"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"sync"
	"time"
)

const (
	// KSUID's epoch starts more recently so that the 32-bit number space gives a
	// significantly higher useful lifetime of around 136 years from March 2017.
	// This number (14e8) was picked to be easy to remember.
	epochStamp int64 = 1400000000

	// Timestamp is a uint32
	timestampLengthInBytes = 4

	// Payload is 16-bytes
	payloadLengthInBytes = 16

	// KSUIDs are 20 bytes when binary encoded
	byteLength = timestampLengthInBytes + payloadLengthInBytes

	// The length of a KSUID when string (base62) encoded
	stringEncodedLength = 27

	// A string-encoded minimum value for a KSUID
	minStringEncoded = "000000000000000000000000000"

	// A string-encoded maximum value for a KSUID
	maxStringEncoded = "aWgEPTl1tmebfsQzFP4bxwgy80V"
)

// KSUIDs are 20 bytes:
//  00-03 byte: uint32 BE UTC timestamp with custom epoch
//  04-19 byte: random "payload"
type KSUID [byteLength]byte

var (
	rander     = rand.Reader
	randMutex  = sync.Mutex{}
	randBuffer = [payloadLengthInBytes]byte{}

	errSize        = fmt.Errorf("Valid KSUIDs are %v bytes", byteLength)
	errStrSize     = fmt.Errorf("Valid encoded KSUIDs are %v characters", stringEncodedLength)
	errStrValue    = fmt.Errorf("Valid encoded KSUIDs are bounded by %s and %s", minStringEncoded, maxStringEncoded)
	errPayloadSize = fmt.Errorf("Valid KSUID payloads are %v bytes", payloadLengthInBytes)

	// Represents a completely empty (invalid) KSUID
	Nil KSUID
	// Represents the highest value a KSUID can have
	Max = KSUID{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255}
)

// Append appends the string representation of i to b, returning a slice to a
// potentially larger memory area.
func (id KSUID) Append(b []byte) []byte {
	return fastAppendEncodeBase62(b, id[:])
}

// The timestamp portion of the ID as a Time object
func (id KSUID) Time() time.Time {
	return correctedUTCTimestampToTime(id.Timestamp())
}

// The timestamp portion of the ID as a bare integer which is uncorrected
// for KSUID's special epoch.
func (id KSUID) Timestamp() uint32 {
	return binary.BigEndian.Uint32(id[:timestampLengthInBytes])
}

// The 16-byte random payload without the timestamp
func (id KSUID) Payload() []byte {
	return id[timestampLengthInBytes:]
}

// String-encoded representation that can be passed through Parse()
func (id KSUID) String() string {
	return string(id.Append(make([]byte, 0, stringEncodedLength)))
}

// Raw byte representation of KSUID
func (id KSUID) Bytes() []byte {
	// Safe because this is by-value
	return id[:]
}

// IsNil returns true if this is a "nil" KSUID
func (id KSUID) IsNil() bool {
	return id == Nil
}

// Get satisfies the flag.Getter interface, making it possible to use KSUIDs as
// part of of the command line options of a program.
func (id KSUID) Get() interface{} {
	return id
}

// Set satisfies the flag.Value interface, making it possible to use KSUIDs as
// part of of the command line options of a program.
func (id *KSUID) Set(s string) error {
	return id.UnmarshalText([]byte(s))
}

func (id KSUID) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

func (id KSUID) MarshalBinary() ([]byte, error) {
	return id.Bytes(), nil
}

func (id *KSUID) UnmarshalText(b []byte) error {
	ksuid, err := Parse(string(b))
	if err != nil {
		return err
	}
	*id = ksuid
	return nil
}

func (id *KSUID) UnmarshalBinary(b []byte) error {
	ksuid, err := FromBytes(b)
	if err != nil {
		return err
	}
	*id = ksuid
	return nil
}

// Value converts the KSUID into a SQL driver value which can be used to
// directly use the KSUID as parameter to a SQL query.
func (id KSUID) Value() (driver.Value, error) {
	if id.IsNil() {
		return nil, nil
	}
	return id.String(), nil
}

// Scan implements the sql.Scanner interface. It supports converting from
// string, []byte, or nil into a KSUID value. Attempting to convert from
// another type will return an error.
func (id *KSUID) Scan(src interface{}) error {
	switch v := src.(type) {
	case nil:
		return id.scan(nil)
	case []byte:
		return id.scan(v)
	case string:
		return id.scan([]byte(v))
	default:
		return fmt.Errorf("Scan: unable to scan type %T into KSUID", v)
	}
}

func (id *KSUID) scan(b []byte) error {
	switch len(b) {
	case 0:
		*id = Nil
		return nil
	case byteLength:
		return id.UnmarshalBinary(b)
	case stringEncodedLength:
		return id.UnmarshalText(b)
	default:
		return errSize
	}
}

// Parse decodes a string-encoded representation of a KSUID object
func Parse(s string) (KSUID, error) {
	if len(s) != stringEncodedLength {
		return Nil, errStrSize
	}

	src := [stringEncodedLength]byte{}
	dst := [byteLength]byte{}

	copy(src[:], s[:])

	if err := fastDecodeBase62(dst[:], src[:]); err != nil {
		return Nil, errStrValue
	}

	return FromBytes(dst[:])
}

func timeToCorrectedUTCTimestamp(t time.Time) uint32 {
	return uint32(t.Unix() - epochStamp)
}

func correctedUTCTimestampToTime(ts uint32) time.Time {
	return time.Unix(int64(ts)+epochStamp, 0)
}

// Generates a new KSUID. In the strange case that random bytes
// can't be read, it will panic.
func New() KSUID {
	ksuid, err := NewRandom()
	if err != nil {
		panic(fmt.Sprintf("Couldn't generate KSUID, inconceivable! error: %v", err))
	}
	return ksuid
}

// Generates a new KSUID
func NewRandom() (ksuid KSUID, err error) {
	return NewRandomWithTime(time.Now())
}

func NewRandomWithTime(t time.Time) (ksuid KSUID, err error) {
	// Go's default random number generators are not safe for concurrent use by
	// multiple goroutines, the use of the rander and randBuffer are explicitly
	// synchronized here.
	randMutex.Lock()

	_, err = io.ReadAtLeast(rander, randBuffer[:], len(randBuffer))
	copy(ksuid[timestampLengthInBytes:], randBuffer[:])

	randMutex.Unlock()

	if err != nil {
		ksuid = Nil // don't leak random bytes on error
		return
	}

	ts := timeToCorrectedUTCTimestamp(t)
	binary.BigEndian.PutUint32(ksuid[:timestampLengthInBytes], ts)
	return
}

// Constructs a KSUID from constituent parts
func FromParts(t time.Time, payload []byte) (KSUID, error) {
	if len(payload) != payloadLengthInBytes {
		return Nil, errPayloadSize
	}

	var ksuid KSUID

	ts := timeToCorrectedUTCTimestamp(t)
	binary.BigEndian.PutUint32(ksuid[:timestampLengthInBytes], ts)

	copy(ksuid[timestampLengthInBytes:], payload)

	return ksuid, nil
}

// Constructs a KSUID from a 20-byte binary representation
func FromBytes(b []byte) (KSUID, error) {
	var ksuid KSUID

	if len(b) != byteLength {
		return Nil, errSize
	}

	copy(ksuid[:], b)
	return ksuid, nil
}

// Sets the global source of random bytes for KSUID generation. This
// should probably only be set once globally. While this is technically
// thread-safe as in it won't cause corruption, there's no guarantee
// on ordering.
func SetRand(r io.Reader) {
	if r == nil {
		rander = rand.Reader
		return
	}
	rander = r
}

// Implements comparison for KSUID type
func Compare(a, b KSUID) int {
	return bytes.Compare(a[:], b[:])
}

// Sorts the given slice of KSUIDs
func Sort(ids []KSUID) {
	quickSort(ids, 0, len(ids)-1)
}

// IsSorted checks whether a slice of KSUIDs is sorted
func IsSorted(ids []KSUID) bool {
	if len(ids) != 0 {
		min := ids[0]
		for _, id := range ids[1:] {
			if bytes.Compare(min[:], id[:]) > 0 {
				return false
			}
			min = id
		}
	}
	return true
}

func quickSort(a []KSUID, lo int, hi int) {
	if lo < hi {
		pivot := a[hi]
		i := lo - 1

		for j, n := lo, hi; j != n; j++ {
			if bytes.Compare(a[j][:], pivot[:]) < 0 {
				i++
				a[i], a[j] = a[j], a[i]
			}
		}

		i++
		if bytes.Compare(a[hi][:], a[i][:]) < 0 {
			a[i], a[hi] = a[hi], a[i]
		}

		quickSort(a, lo, i-1)
		quickSort(a, i+1, hi)
	}
}

// Next returns the next KSUID after id.
func (id KSUID) Next() KSUID {
	zero := makeUint128(0, 0)

	t := id.Timestamp()
	u := uint128Payload(id)
	v := add128(u, makeUint128(0, 1))

	if v == zero { // overflow
		t++
	}

	return v.ksuid(t)
}

// Prev returns the previoud KSUID before id.
func (id KSUID) Prev() KSUID {
	max := makeUint128(math.MaxUint64, math.MaxUint64)

	t := id.Timestamp()
	u := uint128Payload(id)
	v := sub128(u, makeUint128(0, 1))

	if v == max { // overflow
		t--
	}

	return v.ksuid(t)
}
