package persistence

import (
	"context"
	"outgoing/app/service/account/model"
	"outgoing/x/types"
	"time"
)

type Persister interface {
	Ping() error
	Close() error
	User() UserPersister
}

type Cacher interface {
	// Check cache
	Ping() error
	// Close cache
	Close() error
	// Determine whether the user can continue to login
	ContinueLogin(uid string) (bool, time.Duration, error)
	// Increase the number of login failures
	IncreaseFailedNumber(uid string) (int64, error)
	// Clean the number of login failures
	CleanFailedNumber(uid string) error
}

type UserPersister interface {
	// User registration
	Register(ctx context.Context, id int64, uid types.Uid, mobile, avatar, ip string) (string, error)
	// User login via mobile or vid + password
	LoginViaPassword(_ context.Context, uc *model.UserCredential, pwd, version, deviceID, ip string) error
	// Returns record for a given account mobile
	GetCredentialViaMobile(mobile string) (*model.UserCredential, error)
	// GetAll returns account records for a given list of account IDs
	// GetAll(ids ...uint64) ([]*User, error)
	// Update updates account record
	// Update(id uint64, update map[string]interface{}) error
	// Delete deletes account record
	// Delete(id uint64) error
}
