package password

// Hasher provides methods for generating and comparing password hashes.
type Hasher interface {
	// Generate returns a password derived from the password or an error if the password method failed.
	Generate(p string) (string, error)
	// Compare a password to a password and return nil if they match or an error otherwise.
	Compare(p1, p2 string) error
}
