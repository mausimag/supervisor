package supervisor

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAtomicUint64SetGet(t *testing.T) {
	assert := assert.New(t)

	client := NewClient(
		SetZookeeperNodes("127.0.0.1"),
	)

	if err := client.Connect(); err != nil {
		fmt.Println(err.Error())
	}

	vint64 := NewAtomicUint64(client, "/vars/var01")
	assert.Equal(vint64.TrySet(10), nil)

	val, _ := vint64.Get()
	assert.Equal(val, uint64(10))
}

func TestAtomicUint64SetGetIncDec(t *testing.T) {
	assert := assert.New(t)

	client := NewClient(
		SetZookeeperNodes("127.0.0.1"),
	)

	if err := client.Connect(); err != nil {
		fmt.Println(err.Error())
	}

	vint64 := NewAtomicUint64(client, "/vars/var01")
	assert.Equal(vint64.TrySet(10), nil)

	val, _ := vint64.Get()
	assert.Equal(val, uint64(10))

	assert.Equal(vint64.Increment(), nil)
	val, _ = vint64.Get()
	assert.Equal(val, uint64(11))

	assert.Equal(vint64.Decrement(), nil)
	val, _ = vint64.Get()
	assert.Equal(val, uint64(10))
}
