package supervisor

import (
	"bytes"
	"errors"
	"time"

	"github.com/samuel/go-zookeeper/zk"
)

// MakeValue the function that tries to save data will call this to transform the value
type MakeValue func(preValue []byte) []byte

// MutableAtomicValue holds value before and after save
type MutableAtomicValue struct {
	preValue  []byte
	postValue []byte
}

type atomicValue struct {
	client         *Client
	path           string
	MaxRetries     int
	RetryDelay     int
	RetryDelayUnit time.Duration
}

func (av *atomicValue) getCurrentValue(result *MutableAtomicValue, _stat *zk.Stat) (bool, error) {
	data, stat, err := av.client.checkAndGetNode(av.path)

	if err != nil {
		return false, err
	}

	if stat == nil {
		return false, nil
	}

	*_stat = *stat
	result.preValue = data
	return true, nil
}

func (av *atomicValue) get() (*MutableAtomicValue, error) {
	stat := new(zk.Stat)
	result := &MutableAtomicValue{}

	if _, err := av.getCurrentValue(result, stat); err != nil {
		return nil, err
	}

	result.postValue = result.preValue
	return result, nil
}

func (av *atomicValue) compareAndSet(expected, newValue []byte) error {
	stat := new(zk.Stat)
	result := new(MutableAtomicValue)

	exists, err := av.getCurrentValue(result, stat)
	if err != nil {
		return err
	}

	if !exists {
		return errors.New("Value does not exists")
	}

	if !bytes.Equal(result.preValue, expected) {
		return errors.New("Wrong data version")
	}

	if _, err := av.client.setNodeData(av.path, newValue, stat.Version); err != nil {
		return err
	}

	result.postValue = newValue
	return nil
}

func (av *atomicValue) trySet(makeValue MakeValue) error {
	return av.tryOptimistic(makeValue)
}

// tryOptimistic tries to set the value. In case of error it
// will try again X (RetryCount) times with delay (RetryDelay).
// Each time it receives an error, RetryDelay is increased with
// with the following: RetryDelay = RetryDelay * 3 / 2 + 1
func (av *atomicValue) tryOptimistic(makeValue MakeValue) error {
	result := new(MutableAtomicValue)
	retryCount := 0
	retryDelay := av.RetryDelay

	for retryCount < av.MaxRetries {
		if err := av.tryOnce(result, makeValue); err == nil {
			return nil
		}

		time.Sleep(time.Duration(retryDelay) * av.RetryDelayUnit)
		retryDelay = retryDelay*3/2 + 1 // increase delay time
		retryCount++
	}
	return nil
}

func (av *atomicValue) tryOnce(result *MutableAtomicValue, makeValue MakeValue) error {
	stat := new(zk.Stat)

	exists, err := av.getCurrentValue(result, stat)
	if err != nil {
		return err
	}

	newValue := makeValue(result.preValue)
	if exists {
		if _, err := av.client.setNodeData(av.path, newValue, stat.Version); err != nil {
			return err
		}
	} else {
		if _, err := av.client.createParentNodeIfNotExists(av.path, newValue); err != nil {
			return err
		}
	}

	result.postValue = newValue
	return nil
}

func newAtomicValue(client *Client, path string) *atomicValue {
	return &atomicValue{
		client:         client,
		path:           path,
		MaxRetries:     3,
		RetryDelay:     2,
		RetryDelayUnit: time.Second,
	}
}
