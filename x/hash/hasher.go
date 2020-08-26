package hash

// Hasher provides methods for generating and comparing password hashes.
type Hasher interface {
	// Hash creates a hash from data or returns an error.
	Hash(data []byte) ([]byte, error)
	// Compare compares data with a hash and returns an error if the two do not match.
	Compare(hash, data []byte) error
}
