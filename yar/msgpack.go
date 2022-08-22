package yar

import (
	"bytes"
	"errors"
	"fmt"

	"gopkg.in/vmihailenco/msgpack.v2"
)

type msgpackPack struct {
	PackagerName [8]byte
}

func newMsgpackPack() *msgpackPack {
	name := [8]byte{'M', 'S', 'G', 'P', 'A', 'C', 'K'}
	return &msgpackPack{
		PackagerName: name,
	}
}

func (m *msgpackPack) Marshal(x interface{}) (data []byte, err error) {
	if x == nil {
		return nil, errors.New("yar: serverResponse null")
	}
	data, err = msgpack.Marshal(x)
	return
}

func (m *msgpackPack) GetName() [8]byte {
	return m.PackagerName
}

func (m *msgpackPack) Unmarshal(data []byte, x interface{}) error {
	dec := msgpack.NewDecoder(bytes.NewReader(data))
	dec.DecodeMapFunc = func(d *msgpack.Decoder) (interface{}, error) {
		n, err := d.DecodeMapLen()
		if err != nil {
			return nil, err
		}
		rtn := make(map[string]interface{}, n)
		for i := 0; i < n; i++ {
			mk, err := d.DecodeInterface()
			if err != nil {
				return nil, err
			}
			mv, err := d.DecodeInterface()
			if err != nil {
				return nil, err
			}
			// ATTENTION: unmarshal msgpack-php map struct, mk may be nil value
			if mk == nil {
				continue
			}
			key := encodeFixedPhpKey(mk)
			val := encodeFixedPhpVal(mv)
			rtn[key] = val
		}
		return rtn, nil
	}
	return dec.Decode(&x)
}

func encodeFixedPhpKey(data interface{}) string {
	switch val := data.(type) {
	case string:
		return val
	case int, int64, uint64:
		return fmt.Sprintf("%d", val)
	case float64:
		return fmt.Sprintf("%f", val)
	}
	if data == nil {
		return ""
	}
	return fmt.Sprintf("%v", data)
}

func encodeFixedPhpVal(data interface{}) interface{} {
	switch data.(type) {
	case int:
		return int64(data.(int))
	case int64:
		return int64(data.(int64))
	case uint64:
		return uint64(data.(uint64))
	case float64:
		return int64(data.(float64))
	}
	return data
}
