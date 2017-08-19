package supervisor

import "github.com/samuel/go-zookeeper/zk"

// MakeValue the function that tries to save data will call this to transform the value
type MakeValue func(preValue []byte) []byte

// MutableAtomicValue holds value before and after save
type MutableAtomicValue struct {
	preValue  []byte
	postValue []byte
}

type atomicValue struct {
	client *Client
	path   string
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

func (av *atomicValue) trySet(makeValue MakeValue) error {
	stat := new(zk.Stat)
	result := new(MutableAtomicValue)

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
		client: client,
		path:   path,
	}
}
