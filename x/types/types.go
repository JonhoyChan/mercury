// Package types provides data types for persisting objects in the databases.
package types

import (
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"sort"
	"strings"
)

// ID is a database-specific record id, suitable to be used as a primary key.
type ID uint64

// ZeroID is a constant representing uninitialized ID.
const ZeroID ID = 0

// NullValue is a Unicode DEL character which indicated that the value is being deleted.
const NullValue = "\u2421"

// Lengths of various ID representations
const (
	uidBase64Unpadded = 11
	p2pBase64Unpadded = 22
)

// Default ID prefix
const (
	PrefixUID = "uid"
	PrefixGID = "gid"
)

// IsZero checks if ID is uninitialized.
func (id ID) IsZero() bool {
	return id == ZeroID
}

// Compare returns 0 if uid is equal to u2, 1 if u2 is greater than uid, -1 if u2 is smaller.
func (id ID) Compare(u2 ID) int {
	if id < u2 {
		return -1
	} else if id > u2 {
		return 1
	}
	return 0
}

// MarshalBinary converts ID to byte slice.
func (id ID) MarshalBinary() ([]byte, error) {
	dst := make([]byte, 8)
	binary.LittleEndian.PutUint64(dst, uint64(id))
	return dst, nil
}

// UnmarshalBinary reads ID from byte slice.
func (id *ID) UnmarshalBinary(b []byte) error {
	if len(b) < 8 {
		return errors.New("ID.UnmarshalBinary: invalid length")
	}
	*id = ID(binary.LittleEndian.Uint64(b))
	return nil
}

// UnmarshalText reads ID from string represented as byte slice.
func (id *ID) UnmarshalText(src []byte) error {
	if len(src) != uidBase64Unpadded {
		return errors.New("ID.UnmarshalText: invalid length")
	}
	dec := make([]byte, base64.URLEncoding.WithPadding(base64.NoPadding).DecodedLen(uidBase64Unpadded))
	count, err := base64.URLEncoding.WithPadding(base64.NoPadding).Decode(dec, src)
	if count < 8 {
		if err != nil {
			return errors.New("ID.UnmarshalText: failed to decode " + err.Error())
		}
		return errors.New("ID.UnmarshalText: failed to decode")
	}
	*id = ID(binary.LittleEndian.Uint64(dec))
	return nil
}

// MarshalText converts ID to string represented as byte slice.
func (id *ID) MarshalText() ([]byte, error) {
	if *id == ZeroID {
		return []byte{}, nil
	}
	src := make([]byte, 8)
	dst := make([]byte, base64.URLEncoding.WithPadding(base64.NoPadding).EncodedLen(8))
	binary.LittleEndian.PutUint64(src, uint64(*id))
	base64.URLEncoding.WithPadding(base64.NoPadding).Encode(dst, src)
	return dst, nil
}

// MarshalJSON converts ID to double quoted ("ajjj") string.
func (id *ID) MarshalJSON() ([]byte, error) {
	dst, _ := id.MarshalText()
	return append(append([]byte{'"'}, dst...), '"'), nil
}

// UnmarshalJSON reads ID from a double quoted string.
func (id *ID) UnmarshalJSON(b []byte) error {
	size := len(b)
	if size != (uidBase64Unpadded + 2) {
		return errors.New("ID.UnmarshalJSON: invalid length")
	} else if b[0] != '"' || b[size-1] != '"' {
		return errors.New("ID.UnmarshalJSON: unrecognized")
	}
	return id.UnmarshalText(b[1 : size-1])
}

// String converts ID to base64 string.
func (id ID) String() string {
	buf, _ := id.MarshalText()
	return string(buf)
}

// String32 converts ID to lowercase base32 string (suitable for file names on Windows).
func (id ID) String32() string {
	data, _ := id.MarshalBinary()
	return strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(data))
}

// ParseID parses string NOT prefixed with anything
func ParseID(s string) ID {
	var id ID
	_ = id.UnmarshalText([]byte(s))
	return id
}

// ParseID32 parses base32-encoded string into ID
func ParseID32(s string) ID {
	var id ID
	if data, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(s); err == nil {
		_ = id.UnmarshalBinary(data)
	}
	return id
}

// PrefixId converts ID to string prefixed with the given prefix.
func (id ID) PrefixId(prefix string) string {
	if id.IsZero() {
		return ""
	}
	return prefix + id.String()
}

// UserId converts ID to string prefixed with 'uid'
func (id ID) UID() string {
	return id.PrefixId(PrefixUID)
}

// UserId converts ID to string prefixed with 'gid'
func (id ID) GID() string {
	return id.PrefixId(PrefixGID)
}

// ParseIDWithPrefix parses ID with the given prefix
func ParseIDWithPrefix(s, prefix string) ID {
	var id ID
	if strings.HasPrefix(s, prefix) {
		_ = (&id).UnmarshalText([]byte(s)[3:])
	}
	return id
}

// ParseUID parses ID with 'uid'
func ParseUID(s string) ID {
	return ParseIDWithPrefix(s, PrefixUID)
}

// ParseGID parses ID with 'gid'
func ParseGID(s string) ID {
	return ParseIDWithPrefix(s, PrefixGID)
}

// IDSlice is a slice of IDs sorted in ascending order.
type IDSlice []ID

func (us IDSlice) find(id ID) (int, bool) {
	l := len(us)
	if l == 0 || us[0] > id {
		return 0, false
	}
	if id > us[l-1] {
		return l, false
	}
	idx := sort.Search(l, func(i int) bool {
		return id <= us[i]
	})
	return idx, idx < l && us[idx] == id
}

// Add uid to IDSlice keeping it sorted. Duplicates are ignored.
func (us *IDSlice) Add(id ID) bool {
	idx, found := us.find(id)
	if found {
		return false
	}
	// Inserting without creating a temporary slice.
	*us = append(*us, ZeroID)
	copy((*us)[idx+1:], (*us)[idx:])
	(*us)[idx] = id
	return true
}

// Rem removes uid from IDSlice.
func (us *IDSlice) Rem(id ID) bool {
	idx, found := us.find(id)
	if !found {
		return false
	}
	if idx == len(*us)-1 {
		*us = (*us)[:idx]
	} else {
		*us = append((*us)[:idx], (*us)[idx+1:]...)
	}
	return true
}

// Contains checks if the IDSlice contains the given uid
func (us IDSlice) Contains(id ID) bool {
	_, contains := us.find(id)
	return contains
}

// P2PName takes two IDs and generates a P2P topic name
func (id ID) P2PName(u2 ID) string {
	if !id.IsZero() && !u2.IsZero() {
		b1, _ := id.MarshalBinary()
		b2, _ := u2.MarshalBinary()

		if id < u2 {
			b1 = append(b1, b2...)
		} else if id > u2 {
			b1 = append(b2, b1...)
		} else {
			return "self" + base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b1)
		}

		return "p2p" + base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b1)
	}

	return ""
}

// ParseP2P extracts uids from the name of a p2p topic.
func ParseP2P(p2p string) (uid1, uid2 ID, err error) {
	if strings.HasPrefix(p2p, "p2p") {
		src := []byte(p2p)[3:]
		if len(src) != p2pBase64Unpadded {
			err = errors.New("ParseP2P: invalid length")
			return
		}
		dec := make([]byte, base64.URLEncoding.WithPadding(base64.NoPadding).DecodedLen(p2pBase64Unpadded))
		var count int
		count, err = base64.URLEncoding.WithPadding(base64.NoPadding).Decode(dec, src)
		if count < 16 {
			if err != nil {
				err = errors.New("ParseP2P: failed to decode " + err.Error())
			} else {
				err = errors.New("ParseP2P: invalid decoded length")
			}
			return
		}
		uid1 = ID(binary.LittleEndian.Uint64(dec))
		uid2 = ID(binary.LittleEndian.Uint64(dec[8:]))
	} else {
		err = errors.New("ParseP2P: missing or invalid prefix")
	}
	return
}
