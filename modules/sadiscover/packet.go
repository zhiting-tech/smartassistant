package sadiscover

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
)

var helloPacket = []byte{0x21, 0x31, 0x00, 0x20, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

const (
	packetHead = 0x2131

	reserveTokenKey = 0xfffe
	reserveHello    = 0xffff

	_packetHeaderOffset = 0 // 包头偏移
	_packetHeaderSize   = 2 // 包头大小

	_packetLenOffset = _packetHeaderOffset + _packetHeaderSize // 包长度偏移
	_packetLenSize   = 2                                       // 包长度大小

	_packetReserveOffset = _packetLenOffset + _packetLenSize // 预留偏移
	_packetReserveSize   = 2                                 // 预留大小

	_packetDeviceIDOffset = _packetReserveOffset + _packetReserveSize // 设备id偏移
	_packetDeviceIDSize   = 6                                         // 设备id大小

	_packetSerialNumOffset = _packetDeviceIDOffset + _packetDeviceIDSize // 序列号偏移
	_packetSerialNumSize   = 4                                           // 序列号大小

	_packetMD5SumOffset = _packetSerialNumOffset + _packetSerialNumSize // md5校验偏移
	_packetMD5SumSize   = 16                                            // md5校验大小

	_packetValidDataOffset = _packetMD5SumOffset + _packetMD5SumSize // 有效数据偏移
	_packetValidDataSize   = 32                                      // 有效数据大小
)

type Packet struct {
	Head      []byte
	Len       int
	Reserve   []byte
	DeviceID  []byte
	SerialNum uint32
	MD5Sum    []byte
	Data      []byte
}

func Encode(msg, token, deviceID []byte, serialNum uint32) (buf []byte) {

	cryptoMsg := Encrypt(msg, token)
	length := len(cryptoMsg) + 32
	buf = make([]byte, length)

	binary.BigEndian.PutUint16(buf[0:2], packetHead)
	binary.BigEndian.PutUint16(buf[_packetLenOffset:_packetReserveOffset], uint16(length))
	binary.BigEndian.PutUint16(buf[_packetReserveOffset:_packetDeviceIDOffset], 0)
	copy(buf[_packetDeviceIDOffset:_packetSerialNumOffset], deviceID)
	binary.BigEndian.PutUint32(buf[_packetSerialNumOffset:_packetMD5SumOffset], serialNum)

	copy(buf[_packetValidDataOffset:], cryptoMsg)
	sum := md5Hash(buf)
	copy(buf[_packetMD5SumOffset:_packetValidDataOffset], sum[:])
	return
}
func EncodeKey(key, deviceID []byte) (buf []byte) {

	length := 32

	buf = make([]byte, length)

	binary.BigEndian.PutUint16(buf[0:2], packetHead)
	binary.BigEndian.PutUint16(buf[_packetLenOffset:_packetReserveOffset], uint16(32))
	binary.BigEndian.PutUint16(buf[_packetReserveOffset:_packetDeviceIDOffset], reserveTokenKey)
	copy(buf[_packetDeviceIDOffset:_packetSerialNumOffset], deviceID)
	copy(buf[_packetMD5SumOffset:_packetValidDataOffset], key[:16])
	return
}

func DecodeHello(buf []byte) (id []byte, model string, err error) {

	var packet Packet
	packet, err = Decode(buf)
	if err != nil {
		logrus.Error("binary read err", err)
		return
	}

	resp := struct {
		Model string
	}{}
	if err = json.Unmarshal(packet.Data, &resp); err != nil {
		return
	}
	return packet.DeviceID, resp.Model, nil
}

func Decode(buf []byte) (packet Packet, err error) {

	head := buf[:_packetLenOffset]
	if binary.BigEndian.Uint16(head) != 0x2131 {
		err = fmt.Errorf("invalid packet %x\n", head)
		return
	}
	packetLenBytes := buf[_packetLenOffset:_packetReserveOffset]
	packet.Len = int(binary.BigEndian.Uint16(packetLenBytes))

	// fmt.Printf("<<-- %x\n", buf[:packet.Len])
	packet.Reserve = buf[_packetReserveOffset:_packetDeviceIDOffset]

	packet.DeviceID = buf[_packetDeviceIDOffset:_packetSerialNumOffset]

	serialNumBytes := buf[_packetSerialNumOffset:_packetMD5SumOffset]
	packet.SerialNum = binary.BigEndian.Uint32(serialNumBytes)

	packet.MD5Sum = buf[_packetMD5SumOffset:_packetValidDataOffset]

	packet.Data = buf[_packetValidDataOffset:packet.Len]
	// fmt.Printf("<<-- packet:%v\n", packet)
	return
}

func Encrypt(msg, token []byte) (dst []byte) {

	key, iv := getKeyAndIV(token)
	// fmt.Printf("key:%x,iv:%x\n", key, iv)

	// Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		logrus.Errorf("new cipher err: %s", err.Error())
		return
	}

	bm := cipher.NewCBCEncrypter(block, iv)

	msg, _ = pkcs7Pad(msg, block.BlockSize())
	dst = make([]byte, len(msg))
	bm.CryptBlocks(dst, msg)
	return
}

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("pkcs7: Data is empty")
	}
	if length%blockSize != 0 {
		return nil, errors.New("pkcs7: Data is not block-aligned")
	}
	padLen := int(data[length-1])
	ref := bytes.Repeat([]byte{byte(padLen)}, padLen)
	if padLen > blockSize || padLen == 0 || !bytes.HasSuffix(data, ref) {
		return nil, errors.New("pkcs7: Invalid padding")
	}
	return data[:length-padLen], nil
}

// pkcs7pad add pkcs7 padding
func pkcs7Pad(data []byte, blockSize int) ([]byte, error) {
	if blockSize < 0 || blockSize > 256 {
		return nil, fmt.Errorf("pkcs7: Invalid block size %d", blockSize)
	} else {
		padLen := blockSize - len(data)%blockSize
		padding := bytes.Repeat([]byte{byte(padLen)}, padLen)
		return append(data, padding...), nil
	}
}
func Decrypt(enc []byte, token []byte) (result []byte, err error) {
	key, iv := getKeyAndIV(token)

	// Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}
	bm := cipher.NewCBCDecrypter(block, iv)
	result = make([]byte, len(enc))
	bm.CryptBlocks(result, enc)
	result, err = pkcs7Unpad(result, bm.BlockSize())
	if err != nil {
		logrus.Error(err.Error())
		return
	}
	return
}

func md5Hash(datas ...[]byte) (result []byte) {

	hash := md5.New()
	for _, data := range datas {
		_, err := hash.Write(data)
		if err != nil {
			logrus.Errorf("write hash err: %s", err.Error())
		}
	}
	return hash.Sum(nil)
}

func getKeyAndIV(t []byte) ([]byte, []byte) {
	key := md5Hash(t)
	// iv := md5Hash(md5Hash(key), t)
	iv := md5Hash(key, t)
	return key, iv
}
