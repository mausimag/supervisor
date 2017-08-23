package supervisor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMutexTimeout(t *testing.T) {
	assert := assert.New(t)
	clients := makeClientSlice(5)
	lc := len(clients)
	locks := make([]*Mutex, lc)
	lockPath := "/supervisor/test/mutex/key01"

	locks[0] = NewMutex(clients[0], lockPath)
	assert.Equal(locks[0].Acquire(1, time.Second), nil)

	for idx := 1; idx < lc; idx++ {
		locks[idx] = NewMutex(clients[idx], lockPath)
		assert.Equal(locks[idx].Acquire(1, time.Second).Error(), "Timeout")
	}

	assert.Equal(locks[0].Release(), nil)
	closeClients(clients)
}

func TestMutexMultipleLockAndReleasePrev(t *testing.T) {
	assert := assert.New(t)
	clients := makeClientSlice(6)
	locks := make([]*Mutex, len(clients))
	lockPath := "/supervisor/test/mutex/key02"

	for idx, client := range clients {
		locks[idx] = NewMutex(client, lockPath)
	}

	for idx, lock := range locks {
		if (idx+1)%2 == 0 {
			assert.Equal(locks[idx-1].Release(), nil)
		} else {
			assert.Equal(lock.Acquire(1, time.Second), nil)
		}
	}

	closeClients(clients)
}
