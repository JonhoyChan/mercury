package persistence

type Cacher interface {
	// Check cache
	Ping() error
	// Close cache
	Close() error
	// Add mapping
	AddMapping(uid, sid, serverID string) error
	// Set the expiration time of the mapping
	ExpireMapping(uid, sid string) (bool, error)
	// Delete the mapping
	DeleteMapping(uid, sid string) error
}
