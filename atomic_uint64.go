package supervisor

import "encoding/binary"

// AtomicUint64 atomic uint64
type AtomicUint64 struct {
	atomicValue *atomicValue
}

// Increment increments current saved value
func (ai64 *AtomicUint64) Increment() error {
	return ai64.atomicValue.trySet(func(preValue []byte) []byte {
		var pre uint64

		if preValue != nil {
			pre = ai64.fromBytes(preValue)
		}

		post := pre + 1

		return ai64.toBytes(post)
	})
}

// Decrement decrements current saved value
func (ai64 *AtomicUint64) Decrement() error {
	return ai64.atomicValue.trySet(func(preValue []byte) []byte {
		var pre uint64

		if preValue != nil {
			pre = ai64.fromBytes(preValue)
		}

		post := pre

		if pre > 0 {
			post--
		} else {
			post = 0
		}

		return ai64.toBytes(post)
	})
}

// TrySet tries to set or override with new value
func (ai64 *AtomicUint64) TrySet(v uint64) error {
	return ai64.atomicValue.trySet(func(preValue []byte) []byte {
		return ai64.toBytes(v)
	})
}

// Get retries current value
func (ai64 *AtomicUint64) Get() (uint64, error) {
	av, err := ai64.atomicValue.get()
	if err != nil {
		return 0, err
	}
	return ai64.fromBytes(av.postValue), nil
}

func (ai64 *AtomicUint64) toBytes(v uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, v)
	return b
}

func (ai64 *AtomicUint64) fromBytes(b []byte) uint64 {
	return binary.LittleEndian.Uint64(b)
}

// NewAtomicUint64 returns new AtomicUint64
func NewAtomicUint64(client *Client, path string) *AtomicUint64 {
	al := AtomicUint64{
		atomicValue: newAtomicValue(client, path),
	}
	return &al
}
