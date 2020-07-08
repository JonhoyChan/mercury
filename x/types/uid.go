package types

import (
	"encoding/base64"
	"encoding/binary"
	"outgoing/x/config"

	"github.com/bwmarrin/snowflake"
	"golang.org/x/crypto/xtea"
)

// UidGenerator holds snowflake and encryption paramenets.
type UidGenerator struct {
	seq    *snowflake.Node
	cipher *xtea.Cipher
}

type GeneratorUidProvider interface {
	GeneratorUid() *config.GeneratorUidConfig
}

// Init initialises the Uid generator
func (ug *UidGenerator) Init(c GeneratorUidProvider) error {
	var err error

	if ug.seq == nil {
		ug.seq, err = snowflake.NewNode(c.GeneratorUid().WorkID)
	}
	if ug.cipher == nil {
		ug.cipher, err = xtea.NewCipher(c.GeneratorUid().Key)
	}

	return err
}

// Get generates a unique weakly-encryped random-looking ID.
// The Uid is a unit64 with the highest bit possibly set which makes it
// incompatible with go's pre-1.9 gorm package.
func (ug *UidGenerator) Get() Uid {
	buf := getIDBuffer(ug)
	return Uid(binary.LittleEndian.Uint64(buf))
}

// GetStr generates the same unique ID as Get then returns it as
// base64-encoded string. Slightly more efficient than calling Get()
// then base64-encoding the result.
func (ug *UidGenerator) GetStr() string {
	buf := getIDBuffer(ug)
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(buf)
}

// getIdBuffer returns a byte array holding the Uid bytes
func getIDBuffer(ug *UidGenerator) []byte {
	id := uint64(ug.seq.Generate().Int64())

	var src = make([]byte, 8)
	var dst = make([]byte, 8)
	binary.LittleEndian.PutUint64(src, id)
	ug.cipher.Encrypt(dst, src)

	return dst
}

// DecodeUid takes an encrypted Uid and decrypts it into a non-negative int64.
// This is needed for go/gorm compatibility where uint64 with high bit
// set is unsupported and possibly for other uses such as MySQL's recommendation
// for sequential primary keys.
func (ug *UidGenerator) DecodeUid(uid Uid) int64 {
	var src = make([]byte, 8)
	var dst = make([]byte, 8)
	binary.LittleEndian.PutUint64(src, uint64(uid))
	ug.cipher.Decrypt(dst, src)
	return int64(binary.LittleEndian.Uint64(dst))
}

// EncodeInt64 takes a positive int64 and encrypts it into a Uid.
// This is needed for go/gorm compatibility where uint64 with high bit
// set is unsupported  and possibly for other uses such as MySQL's recommendation
// for sequential primary keys.
func (ug *UidGenerator) EncodeInt64(val int64) Uid {
	var src = make([]byte, 8)
	var dst = make([]byte, 8)
	binary.LittleEndian.PutUint64(src, uint64(val))
	ug.cipher.Encrypt(dst, src)
	return Uid(binary.LittleEndian.Uint64(dst))
}
