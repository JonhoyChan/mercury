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

// Uid is a database-specific record id, suitable to be used as a primary key.
type Uid uint64

// ZeroUid is a constant representing uninitialized Uid.
const ZeroUid Uid = 0

// NullValue is a Unicode DEL character which indicated that the value is being deleted.
const NullValue = "\u2421"

// Lengths of various Uid representations
const (
	uidBase64Unpadded = 11
	p2pBase64Unpadded = 22
)

// IsZero checks if Uid is uninitialized.
func (uid Uid) IsZero() bool {
	return uid == ZeroUid
}

// Compare returns 0 if uid is equal to u2, 1 if u2 is greater than uid, -1 if u2 is smaller.
func (uid Uid) Compare(u2 Uid) int {
	if uid < u2 {
		return -1
	} else if uid > u2 {
		return 1
	}
	return 0
}

// MarshalBinary converts Uid to byte slice.
func (uid Uid) MarshalBinary() ([]byte, error) {
	dst := make([]byte, 8)
	binary.LittleEndian.PutUint64(dst, uint64(uid))
	return dst, nil
}

// UnmarshalBinary reads Uid from byte slice.
func (uid *Uid) UnmarshalBinary(b []byte) error {
	if len(b) < 8 {
		return errors.New("Uid.UnmarshalBinary: invalid length")
	}
	*uid = Uid(binary.LittleEndian.Uint64(b))
	return nil
}

// UnmarshalText reads Uid from string represented as byte slice.
func (uid *Uid) UnmarshalText(src []byte) error {
	if len(src) != uidBase64Unpadded {
		return errors.New("Uid.UnmarshalText: invalid length")
	}
	dec := make([]byte, base64.URLEncoding.WithPadding(base64.NoPadding).DecodedLen(uidBase64Unpadded))
	count, err := base64.URLEncoding.WithPadding(base64.NoPadding).Decode(dec, src)
	if count < 8 {
		if err != nil {
			return errors.New("Uid.UnmarshalText: failed to decode " + err.Error())
		}
		return errors.New("Uid.UnmarshalText: failed to decode")
	}
	*uid = Uid(binary.LittleEndian.Uint64(dec))
	return nil
}

// MarshalText converts Uid to string represented as byte slice.
func (uid *Uid) MarshalText() ([]byte, error) {
	if *uid == ZeroUid {
		return []byte{}, nil
	}
	src := make([]byte, 8)
	dst := make([]byte, base64.URLEncoding.WithPadding(base64.NoPadding).EncodedLen(8))
	binary.LittleEndian.PutUint64(src, uint64(*uid))
	base64.URLEncoding.WithPadding(base64.NoPadding).Encode(dst, src)
	return dst, nil
}

// MarshalJSON converts Uid to double quoted ("ajjj") string.
func (uid *Uid) MarshalJSON() ([]byte, error) {
	dst, _ := uid.MarshalText()
	return append(append([]byte{'"'}, dst...), '"'), nil
}

// UnmarshalJSON reads Uid from a double quoted string.
func (uid *Uid) UnmarshalJSON(b []byte) error {
	size := len(b)
	if size != (uidBase64Unpadded + 2) {
		return errors.New("Uid.UnmarshalJSON: invalid length")
	} else if b[0] != '"' || b[size-1] != '"' {
		return errors.New("Uid.UnmarshalJSON: unrecognized")
	}
	return uid.UnmarshalText(b[1 : size-1])
}

// String converts Uid to base64 string.
func (uid Uid) String() string {
	buf, _ := uid.MarshalText()
	return string(buf)
}

// String32 converts Uid to lowercase base32 string (suitable for file names on Windows).
func (uid Uid) String32() string {
	data, _ := uid.MarshalBinary()
	return strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(data))
}

// ParseUid parses string NOT prefixed with anything
func ParseUid(s string) Uid {
	var uid Uid
	_ = uid.UnmarshalText([]byte(s))
	return uid
}

// ParseUid32 parses base32-encoded string into Uid
func ParseUid32(s string) Uid {
	var uid Uid
	if data, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(s); err == nil {
		_ = uid.UnmarshalBinary(data)
	}
	return uid
}

// PrefixId converts Uid to string prefixed with the given prefix.
func (uid Uid) PrefixId(prefix string) string {
	if uid.IsZero() {
		return ""
	}
	return prefix + uid.String()
}

// UserId converts Uid to string prefixed with 'vid', like uidXXXXX
func (uid Uid) UID() string {
	return uid.PrefixId("uid")
}

// ParseUserUID parses account VID of the form "uidXXXXXX"
func ParseUserUID(s string) Uid {
	var uid Uid
	if strings.HasPrefix(s, "uid") {
		_ = (&uid).UnmarshalText([]byte(s)[3:])
	}
	return uid
}

// UidSlice is a slice of Uids sorted in ascending order.
type UidSlice []Uid

func (us UidSlice) find(uid Uid) (int, bool) {
	l := len(us)
	if l == 0 || us[0] > uid {
		return 0, false
	}
	if uid > us[l-1] {
		return l, false
	}
	idx := sort.Search(l, func(i int) bool {
		return uid <= us[i]
	})
	return idx, idx < l && us[idx] == uid
}

// Add uid to UidSlice keeping it sorted. Duplicates are ignored.
func (us *UidSlice) Add(uid Uid) bool {
	idx, found := us.find(uid)
	if found {
		return false
	}
	// Inserting without creating a temporary slice.
	*us = append(*us, ZeroUid)
	copy((*us)[idx+1:], (*us)[idx:])
	(*us)[idx] = uid
	return true
}

// Rem removes uid from UidSlice.
func (us *UidSlice) Rem(uid Uid) bool {
	idx, found := us.find(uid)
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

// Contains checks if the UidSlice contains the given uid
func (us UidSlice) Contains(uid Uid) bool {
	_, contains := us.find(uid)
	return contains
}

// P2PName takes two Uids and generates a P2P topic name
func (uid Uid) P2PName(u2 Uid) string {
	if !uid.IsZero() && !u2.IsZero() {
		b1, _ := uid.MarshalBinary()
		b2, _ := u2.MarshalBinary()

		if uid < u2 {
			b1 = append(b1, b2...)
		} else if uid > u2 {
			b1 = append(b2, b1...)
		} else {
			// Explicitly disable P2P with self
			return ""
		}

		return "p2p" + base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b1)
	}

	return ""
}

// ParseP2P extracts uids from the name of a p2p topic.
func ParseP2P(p2p string) (uid1, uid2 Uid, err error) {
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
		uid1 = Uid(binary.LittleEndian.Uint64(dec))
		uid2 = Uid(binary.LittleEndian.Uint64(dec[8:]))
	} else {
		err = errors.New("ParseP2P: missing or invalid prefix")
	}
	return
}

// TopicCategory is an enum of topic categories.
type TopicCategory int

const (
	// TopicCategoryMe is a value denoting 'me' topic.
	TopicCategoryMe TopicCategory = iota
	// TopicCategoryP2P is a a value denoting 'p2p topic.
	TopicCategoryP2P
	// TopicCategoryGrp is a a value denoting group topic.
	TopicCategoryGroup
	// TopicCategorySystem is a constant indicating a system topic.
	TopicCategorySystem
)

// GetTopicCategory given topic name returns topic category.
func GetTopicCategory(name string) TopicCategory {
	switch name[:3] {
	case "me":
		return TopicCategoryMe
	case "p2p":
		return TopicCategoryP2P
	case "group":
		return TopicCategoryGroup
	case "system":
		return TopicCategorySystem
	default:
		panic("invalid topic category for name '" + name + "'")
	}
}
