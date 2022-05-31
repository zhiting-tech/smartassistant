package datatunnel

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

var (
	fingerprint = []byte{0x00, 0x5a, 0x68, 0x69, 0x74, 0x69, 0x6e, 0x67}
	version     = []byte{0x00, 0x01}
)

type proxyConnectProtocol struct {
	values   [][]byte
	dataSize int64
}

func (p *proxyConnectProtocol) Add(data []byte) {
	p.values = append(p.values, data)
	p.dataSize += int64(len(data) + 4)
}

func (p *proxyConnectProtocol) Encode() ([]byte, error) {

	var bufSize int = len(fingerprint) + len(version) + 2 + 8 + int(p.dataSize)
	buffer := bytes.NewBuffer(nil)
	buffer.Write(fingerprint)
	buffer.Write(version)
	binary.Write(buffer, binary.BigEndian, int16(len(p.values)))
	binary.Write(buffer, binary.BigEndian, int64(p.dataSize))
	for _, value := range p.values {
		binary.Write(buffer, binary.BigEndian, int32(len(value)))
		buffer.Write(value)
	}

	bytes := buffer.Bytes()
	if len(bytes) != int(bufSize) {
		return nil, fmt.Errorf("encode error")
	}

	return bytes, nil
}

func (p *proxyConnectProtocol) Decode(reader io.Reader) (err error) {
	buf := make([]byte, len(fingerprint)+len(version)+2+8)
	if _, err = io.ReadFull(reader, buf); err != nil {
		return
	}

	bufReader := bytes.NewReader(buf)

	var f []byte = make([]byte, len(fingerprint))
	if _, err = bufReader.Read(f); err != nil {
		return
	}
	if !bytes.Equal(f, fingerprint) {
		err = fmt.Errorf("invalid package")
		return
	}

	var v []byte = make([]byte, len(version))
	if _, err = bufReader.Read(v); err != nil {
		return
	}

	var num int16
	if err = binary.Read(bufReader, binary.BigEndian, &num); err != nil {
		return
	}

	if err = binary.Read(bufReader, binary.BigEndian, &p.dataSize); err != nil {
		return
	}

	buf = make([]byte, p.dataSize)
	if _, err = io.ReadFull(reader, buf); err != nil {
		return
	}
	bufReader = bytes.NewReader(buf)

	for i := 0; i < int(num); i++ {
		var size int32
		if err = binary.Read(bufReader, binary.BigEndian, &size); err != nil {
			return
		}

		value := make([]byte, size)
		if _, err = bufReader.Read(value); err != nil {
			return
		}

		p.values = append(p.values, value)
	}

	return
}

func (p *proxyConnectProtocol) Values() [][]byte {
	return p.values
}
