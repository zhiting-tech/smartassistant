package control

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"reflect"

	"google.golang.org/protobuf/proto"
)

var (
	ErrInvalidType = fmt.Errorf("invalid type")
)

type DefaultBinaryCoder struct{}

func (b *DefaultBinaryCoder) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (b *DefaultBinaryCoder) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (b *DefaultBinaryCoder) Encode(v reflect.Value) (data []byte, err error) {

	t := v.Type()
	if b.InvalidType(t, 0) {
		err = ErrInvalidType
		return
	}

	var pv reflect.Value
	var pt reflect.Type
	if t.Kind() == reflect.Ptr {
		pv = v
		pt = t
		v = v.Elem()
		t = v.Type()
	}

	switch t.Kind() {
	case reflect.Struct:
		if pt != nil && pt.Kind() == reflect.Ptr {
			v = pv
		}

		value := v.Interface()
		var message proto.Message
		var ok bool
		if message, ok = value.(proto.Message); ok {
			if data, err = proto.Marshal(message); err != nil {
				return
			}
		} else {
			if data, err = json.Marshal(value); err != nil {
				return
			}
		}

	case reflect.String:
		data = []byte(v.Interface().(string))
	case reflect.Int:
		// 默认当成32位处理
		var num int32 = int32(v.Interface().(int))
		writer := bytes.NewBuffer(nil)
		binary.Write(writer, binary.BigEndian, num)
		data = writer.Bytes()
	case reflect.Uint:
		// 同上
		var num uint32 = uint32(v.Interface().(uint))
		writer := bytes.NewBuffer(nil)
		binary.Write(writer, binary.BigEndian, num)
		data = writer.Bytes()
	default:
		writer := bytes.NewBuffer(nil)
		if err = binary.Write(writer, binary.BigEndian, v.Interface()); err != nil {
			return
		}
		data = writer.Bytes()
	}

	return
}

func (b *DefaultBinaryCoder) Decode(data []byte, t reflect.Type) (v reflect.Value, err error) {

	if b.InvalidType(t, 0) {
		err = ErrInvalidType
		return
	}

	ptr := false
	var nv reflect.Value
	if t.Kind() == reflect.Ptr {
		ptr = true
		t = t.Elem()
	}
	nv = reflect.New(t)

	switch t.Kind() {
	case reflect.Struct:
		var message proto.Message
		var ok bool
		if message, ok = nv.Interface().(proto.Message); ok {
			if err = proto.Unmarshal(data, message); err != nil {
				return
			}
		} else {
			if err = json.Unmarshal(data, nv.Interface()); err != nil {
				return
			}
		}

	case reflect.String:
		nv.Elem().SetString(string(data))
	case reflect.Int:
		// 默认当成32位处理
		var num int32
		reader := bytes.NewReader(data)
		binary.Read(reader, binary.BigEndian, num)
		nv.Elem().SetInt(int64(num))
	case reflect.Uint:
		// 同上
		var num uint32
		reader := bytes.NewReader(data)
		binary.Read(reader, binary.BigEndian, num)
		nv.Elem().SetUint(uint64(num))
	default:
		reader := bytes.NewReader(data)
		if err = binary.Read(reader, binary.BigEndian, nv.Interface()); err != nil {
			return
		}
	}

	if ptr {
		v = nv
	} else {
		v = nv.Elem()
	}

	return
}

func (b *DefaultBinaryCoder) InvalidType(t reflect.Type, deep int) bool {
	if deep > 1 {
		return true
	}

	switch t.Kind() {
	case reflect.Slice:
		if deep > 0 {
			return false
		}
	case reflect.Array:
		if deep > 0 {
			return false
		}
	case reflect.Invalid:
	case reflect.Uintptr:
	case reflect.Complex64:
	case reflect.Complex128:
	case reflect.Chan:
	case reflect.Func:
	case reflect.Interface:
	case reflect.Map:
	case reflect.UnsafePointer:
	case reflect.Ptr:
		return b.InvalidType(t.Elem(), deep+1)
	default:
		return false
	}

	return true
}
