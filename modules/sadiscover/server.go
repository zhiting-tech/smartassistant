package sadiscover

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"net"

	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"github.com/zhiting-tech/smartassistant/pkg/rand"
)

var saID []byte
var key []byte

func initSaID() {
	// 转换saID字符串为数值类型
	hexID, err := hex.DecodeString(config.GetConf().SmartAssistant.ID)
	if err == nil {
		saID = hexID
		return
	}
	// 兼容旧版，将删除
	saIDBytes := []byte(config.GetConf().SmartAssistant.ID)
	data := uint32(0)
	for _, b := range saIDBytes {
		data = (data << 8) | uint32(b)
	}
	saID = make([]byte, 6)
	binary.BigEndian.PutUint32(saID, data)
}

type Server struct {
}

func NewSaDiscoverServer() *Server {
	return &Server{}
}

func (s *Server) Run(ctx context.Context) {
	initSaID()

	// 随机生成一个token
	m := md5.New()
	m.Write([]byte(rand.String(32)))
	token, _ := hex.DecodeString(hex.EncodeToString(m.Sum(nil)))

	go s.readFromUDP(token)

	<-ctx.Done()
	logger.Warning("sa discover server stopped")
}

func (s *Server) readFromUDP(token []byte) {
	addr, err := net.ResolveUDPAddr("udp", ":54321")
	if err != nil {
		logger.Error("[sa discover] Can't resolve address: ", err)
		return
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		logger.Error("[sa discover] Error listening:", err)
		return
	}
	defer conn.Close()

	data := make([]byte, 1024)
	logger.Debug("starting sa discover server")
	for {
		n, remoteAddr, err := conn.ReadFromUDP(data)
		if err != nil {
			logger.Error("[sa discover] failed to read UDP msg because of ", err.Error())
			continue
		}
		// 响应hello应答包
		if bytes.Compare(helloPacket, data[:n]) == 0 {
			_, err := conn.WriteTo(helloResponse(), remoteAddr)
			if err != nil {
				logger.Warnf("[sa discover] write response error %v", err)
			}
			continue
		}

		// 解密data
		packet, err := Decode(data)
		if err != nil {
			logger.Warn("[sa discover] decode error", err)
			continue
		}

		// 加密token，返回token
		if binary.BigEndian.Uint16(packet.Reserve) == reserveTokenKey {
			key = packet.MD5Sum
			msg := Encode(token, key, saID, packet.SerialNum)
			_, err := conn.WriteTo(msg, remoteAddr)
			if err != nil {
				logger.Warn("[sa discover] write to error", err)
			}
			continue
		}

		if len(packet.Data) == 0 {
			continue
		}
		// 返回sa信息
		result, err := Decrypt(packet.Data, token)
		if err != nil {
			logger.Warn("[sa discover] decrypt error", err)
		}
		//logger.Infof("[sa discover] get result bytes %v from %v", result, remoteAddr)
		//logger.Infof("[sa discover] get result %s from %v", string(result), remoteAddr)

		msg := make(map[string]interface{})
		if err = json.Unmarshal(result, &msg); err != nil {
			continue
		}
		method, ok := msg["method"]
		if !ok {
			continue
		}
		if method == "get_prop.info" {
			re := Result{
				ID: int(msg["id"].(float64)),
				Re: Info{
					Model: "smart_assistant",
					SwVer: types.Version,
					HwVer: types.HardwareVersion,
					SaID:  config.GetConf().SmartAssistant.ID,
				},
			}
			re.Re.Port, ok = config.GetConf().Datatunnel.GetPort("http")
			if !ok {
				continue
			}
			toMsg, _ := json.Marshal(re)
			buf := Encode(toMsg, token, saID, packet.SerialNum)
			_, err := conn.WriteTo(buf, remoteAddr)
			if err != nil {
				logger.Warn("[sa discover] write to error", err)
				continue
			}
		}
	}
}

func helloResponse() []byte {
	resp := struct {
		Model string `json:"model"`
	}{Model: types.SaModel}
	data, _ := json.Marshal(resp)
	length := len(data) + 32
	buf := make([]byte, length)
	binary.BigEndian.PutUint16(buf[0:2], packetHead)
	binary.BigEndian.PutUint16(buf[_packetLenOffset:_packetReserveOffset], uint16(length))
	binary.BigEndian.PutUint16(buf[_packetReserveOffset:_packetDeviceIDOffset], reserveHello)
	copy(buf[_packetDeviceIDOffset:_packetSerialNumOffset], saID)
	binary.BigEndian.PutUint32(buf[_packetSerialNumOffset:_packetMD5SumOffset], 0x01)
	copy(buf[_packetValidDataOffset:], data)
	sum := md5Hash(buf)
	copy(buf[_packetMD5SumOffset:_packetValidDataOffset], sum[:])
	return buf
}
