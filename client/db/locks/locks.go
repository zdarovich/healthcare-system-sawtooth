package locks

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"healthcare-system-sawtooth/client/lib"
	"sync"
	"sync/atomic"
	"time"

	"github.com/globalsign/mgo"
	lock "github.com/square/mongo-lock"
)

type LockType string

const (
	TypeS LockType = "S"
	TypeX LockType = "X"

	lockCollection = "Locks"

	UserKeyRotationLock = "UserKeyRotation"
	OrgKeyRotationLock  = "OrgKeyRotation"

	GeneralTTL = 10
)

var (
	ErrAlreadyLocked = lock.ErrAlreadyLocked
	ErrLockNotFound  = lock.ErrLockNotFound

	mongoSession *mgo.Session
	mongoInit    uint32
	mongoMu      sync.Mutex
	lockClient   *lock.Client
	lockPurger   lock.Purger

	thisMachineId = "id"
)

func createIndexes(session *mgo.Session) error {
	lockCollection := session.DB(lib.MongoDbName).C(lockCollection)
	indexes := []mgo.Index{
		// Required.
		{
			Key:        []string{"resource"},
			Unique:     true,
			Background: false,
			Sparse:     true,
		},
		// Optional.
		{Key: []string{"exclusive.LockId"}},
		{Key: []string{"exclusive.ExpiresAt"}},
		{Key: []string{"shared.locks.LockId"}},
		{Key: []string{"shared.locks.ExpiresAt"}},
	}
	for _, idx := range indexes {
		if err := lockCollection.EnsureIndex(idx); err != nil {
			return err
		}
	}
	return nil
}

func getLockClient() (*lock.Client, error) {
	if atomic.LoadUint32(&mongoInit) == 1 {
		return lockClient, nil
	}

	mongoMu.Lock()
	defer mongoMu.Unlock()

	if mongoInit == 0 {
		session, err := mgo.Dial(lib.MongoDbUrl)
		if err != nil {
			return nil, err
		}

		mongoSession = session

		err = createIndexes(mongoSession)
		if err != nil {
			return nil, err
		}

		lockClient = lock.NewClient(session, lib.MongoDbName, lockCollection)
		lockPurger = lock.NewPurger(lockClient)

		atomic.StoreUint32(&mongoInit, 1)
	}

	return lockClient, nil
}

func getLockId(resource string, id string) string {
	return resource + id
}

// getLock gets a lock on the given resource for the given machine.
func getLock(resource string, id string, ttl uint, lockType LockType) error {
	client, err := getLockClient()
	if err != nil {
		return err
	}

	lockId := getLockId(resource, id)

	_, err = lockPurger.Purge()
	if err != nil {
		return err
	}
	if lockType == TypeX {
		err = client.XLock(resource, lockId, lock.LockDetails{TTL: ttl, Owner: id})
	} else {
		err = client.SLock(resource, lockId, lock.LockDetails{TTL: ttl, Owner: id}, -1)
	}
	if err != nil {
		return err
	}

	return nil
}

// GetLock gets a lock on the given resource for the current machine.
func GetLock(resource string, ttl uint) error {
	return getLock(resource, thisMachineId, ttl, TypeX)
}

func getLockWaitWithId(resource, id string, ttl uint, lockType LockType) error {
	var err error
	for err = getLock(resource, id, ttl, lockType); err == ErrAlreadyLocked; err = getLock(resource, id, ttl, lockType) {
		// TODO: handle all this better
		time.Sleep(1000 * time.Millisecond)
	}
	if err != nil {
		return err
	}
	return nil
}

func getLockWaitGenId(resource string, ttl uint, lockType LockType) (string, error) {
	b := make([]byte, 256)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	id := hex.Dump(b)

	err = getLockWaitWithId(resource, id, ttl, lockType)
	if err != nil {
		return "", err
	}

	return id, nil
}

// renewLock renews a lock on the given resource for the given machine.
func renewLock(resource string, machineId string, ttl uint) error {
	client, err := getLockClient()
	if err != nil {
		return err
	}

	lockId := getLockId(resource, machineId)

	lockStatuses, err := client.Renew(lockId, ttl)
	if err != nil {
		return err
	}
	if len(lockStatuses) < 1 || lockStatuses[0].LockId != lockId {
		return fmt.Errorf("Lock on resource %v by machine %v was not renewed. No error reported.", resource, machineId)
	}

	return nil
}

// RenewLock gets a lock on the given resource for the current machine.
func RenewLock(resource string, ttl uint) error {
	return renewLock(resource, thisMachineId, ttl)
}

// getOrRenewLock gets a lock if available or renews the lock if already held by the given machine.
func getOrRenewLock(resource string, machineId string, ttl uint) error {
	err := getLock(resource, machineId, ttl, TypeX)
	if err == ErrAlreadyLocked {
		client, err := getLockClient()
		if err != nil {
			return err
		}
		lockStatuses, statusErr := client.Status(lock.Filter{Resource: resource, TTLgte: 1})
		if statusErr != nil {
			return statusErr
		}
		if len(lockStatuses) < 1 {
			err = getLock(resource, machineId, ttl, TypeX)
			if err != nil {
				return err
			}
			return nil
		}

		if len(lockStatuses) > 1 {
			return fmt.Errorf("Expected to find 1 lock. Found: %v.", len(lockStatuses))
		}

		if lockStatuses[0].Owner == machineId {
			err = renewLock(resource, machineId, ttl)
			if err == ErrLockNotFound {
				err = getLock(resource, machineId, ttl, TypeX)
				if err != nil {
					return err
				}
				return nil
			} else if err != nil {
				return err
			}
			return nil
		}
	}
	return err
}

// GetOrRenewLock gets a lock if available or renews the lock if already held by the current machine.
func GetOrRenewLock(resource string, ttl uint) error {
	return getOrRenewLock(resource, thisMachineId, ttl)
}

// unlock unlocks a lock on the given resource by the given machine.
func UnlockWithId(resource string, id string) error {
	client, err := getLockClient()
	if err != nil {
		return err
	}

	lockId := getLockId(resource, id)

	for _, err := client.Unlock(lockId); err != nil; _, err = client.Unlock(lockId) {
	}
	return nil
}
