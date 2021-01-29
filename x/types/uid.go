package types

import (
	"encoding/base64"
	"encoding/binary"
	"mercury/config"

	"github.com/bwmarrin/snowflake"
	"golang.org/x/crypto/xtea"
)

// IDGenerator holds snowflake and encryption paramenets.
type IDGenerator struct {
	seq    *snowflake.Node
	cipher *xtea.Cipher
}

// Init initialises the ID generator
func (g *IDGenerator) Init(c config.IDGenerator) error {
	var err error

	if g.seq == nil {
		g.seq, err = snowflake.NewNode(c.WorkID)
	}
	if g.cipher == nil {
		g.cipher, err = xtea.NewCipher(c.Key)
	}

	return err
}

// Get generates a unique weakly-encryped random-looking ID.
// The ID is a unit64 with the highest bit possibly set which makes it
// incompatible with go's pre-1.9 gorm package.
func (g *IDGenerator) Get() ID {
	buf := getIDBuffer(g)
	return ID(binary.LittleEndian.Uint64(buf))
}

// GetStr generates the same unique ID as Get then returns it as
// base64-encoded string. Slightly more efficient than calling Get()
// then base64-encoding the result.
func (g *IDGenerator) GetStr() string {
	buf := getIDBuffer(g)
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(buf)
}

// getIdBuffer returns a byte array holding the ID bytes
func getIDBuffer(ug *IDGenerator) []byte {
	id := uint64(ug.seq.Generate().Int64())

	var src = make([]byte, 8)
	var dst = make([]byte, 8)
	binary.LittleEndian.PutUint64(src, id)
	ug.cipher.Encrypt(dst, src)

	return dst
}

// DecodeID takes an encrypted ID and decrypts it into a non-negative int64.
// This is needed for go/gorm compatibility where uint64 with high bit
// set is unsupported and possibly for other uses such as MySQL's recommendation
// for sequential primary keys.
func (g *IDGenerator) DecodeID(id ID) int64 {
	var src = make([]byte, 8)
	var dst = make([]byte, 8)
	binary.LittleEndian.PutUint64(src, uint64(id))
	g.cipher.Decrypt(dst, src)
	return int64(binary.LittleEndian.Uint64(dst))
}

// EncodeInt64 takes a positive int64 and encrypts it into a ID.
// This is needed for go/gorm compatibility where uint64 with high bit
// set is unsupported  and possibly for other uses such as MySQL's recommendation
// for sequential primary keys.
func (g *IDGenerator) EncodeInt64(val int64) ID {
	var src = make([]byte, 8)
	var dst = make([]byte, 8)
	binary.LittleEndian.PutUint64(src, uint64(val))
	g.cipher.Encrypt(dst, src)
	return ID(binary.LittleEndian.Uint64(dst))
}
