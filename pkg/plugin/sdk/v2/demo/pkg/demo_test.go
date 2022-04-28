package pkg

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/proto/v2"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/v2"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/v2/definer"
	"github.com/zhiting-tech/smartassistant/pkg/thingmodel"
)

const (
	// iid = "0x0000000017995bc5" // yeelight color
	// iid = "84f7033a1b71"
	iid = "84f703a5ead5" // 后面开关
)

func TestDemo(t *testing.T) {

	d := NewDemo()
	def := definer.NewThingModelDefiner("123", nil, nil)
	d.Define(def)

	tm := def.ThingModel()
	data, err := json.Marshal(tm)
	logger.Println(string(data), err)
}

func TestGetDevices(t *testing.T) {
	conn, err := grpc.Dial("0.0.0.0:5555", grpc.WithInsecure())
	if err != nil {
		logrus.Panic(err)
	}

	cli := proto.NewPluginClient(conn)

	resp, err := cli.Discover(context.Background(), &emptypb.Empty{})
	if err != nil {
		logrus.Panic(err)
	}
	for {
		d, err := resp.Recv()
		if err != nil {
			logrus.Panic(err.Error())
		}
		m, _ := json.Marshal(d)
		logrus.Println(string(m))
	}
}

func TestConnectDevices(t *testing.T) {

	conn, err := grpc.Dial("0.0.0.0:5555", grpc.WithInsecure())
	if err != nil {
		logrus.Panic(err)
	}

	cli := proto.NewPluginClient(conn)

	iids := []string{"84f7033bb9d9", "84f7033aa275", "84f703a5ead5", "7cdfa1a44b15", "84f7033a1b71"}

	for _, i := range iids {
		_, err = ConnectDevice(cli, i)
	}

}
func ConnectDevice(cli proto.PluginClient, iid string) (tm thingmodel.ThingModel, err error) {

	resp, err := cli.Connect(context.Background(), &proto.AuthReq{Iid: iid})
	if err != nil {
		logrus.Panic(err)
	}

	if err != nil {
		logrus.Panic(err)
	}
	logrus.Println(resp)

	err = json.Unmarshal(resp.Instances, &tm.Instances)
	return

}
func TestConnect(t *testing.T) {
	conn, err := grpc.Dial("0.0.0.0:5555", grpc.WithInsecure())
	if err != nil {
		logrus.Panic(err)
	}

	cli := proto.NewPluginClient(conn)
	tm, err := ConnectDevice(cli, iid)
	if err != nil {
		logrus.Panic(err)
	}

	logrus.Printf("%+v", tm)
	// data, _ := tm.MarshalJSON()
	// logrus.Println(string(data))

}

func TestGetInstances(t *testing.T) {
	conn, err := grpc.Dial("0.0.0.0:5555", grpc.WithInsecure())
	if err != nil {
		logrus.Panic(err)
	}

	cli := proto.NewPluginClient(conn)

	req := proto.GetInstancesReq{Iid: iid}
	resp, err := cli.GetInstances(context.Background(), &req)
	if err != nil {
		logrus.Panic(err)
	}
	logrus.Println(resp)

	var tm thingmodel.ThingModel

	json.Unmarshal(resp.Instances, &tm.Instances)

	logrus.Printf("%+v", tm)
	// data, _ := tm.MarshalJSON()
	// logrus.Println(string(data))

}

func TestSetAttributes(t *testing.T) {
	conn, err := grpc.Dial("0.0.0.0:5555", grpc.WithInsecure())
	if err != nil {
		logrus.Panic(err)
	}

	cli := proto.NewPluginClient(conn)

	// setReq := server.SetRequest{Attributes: []server.SetAttribute{{
	//	IID: "0x0000000017995bc5",
	//	AID: 2,
	//	Val: 1,
	// }}}
	setReq := sdk.SetRequest{Attributes: []sdk.SetAttribute{{
		IID: iid,
		AID: 5,
		Val: "off",
	}}}
	data, _ := json.Marshal(setReq)
	req := proto.SetAttributesReq{Data: data}
	resp, err := cli.SetAttributes(context.Background(), &req)
	if err != nil {
		logrus.Panic(err)
	}
	logrus.Println(resp)

}

func TestStateChange(t *testing.T) {
	conn, err := grpc.Dial("0.0.0.0:5555", grpc.WithInsecure())
	if err != nil {
		logrus.Panic(err)
	}

	cli := proto.NewPluginClient(conn)

	empty := emptypb.Empty{}
	resp, err := cli.Subscribe(context.Background(), &empty)
	if err != nil {
		logrus.Panic(err)
	}

	for {
		attrs, err := resp.Recv()
		if err != nil {
			logrus.Panic(err.Error())
		}
		logrus.Println(attrs.String())
	}
}

func TestOTA(t *testing.T) {
	conn, err := grpc.Dial("0.0.0.0:5555", grpc.WithInsecure())
	if err != nil {
		logrus.Panic(err)
	}

	cli := proto.NewPluginClient(conn)

	// switch 1.0.19
	// url := "https://zt-editor.oss-cn-guangzhou.aliyuncs.com/a70bdb48d89222c8bab1dbc1b6bdcdb74ad273fc2848a8ec3235c306249551f3.bin"

	// switch 1.0.20
	url := "https://zt-editor.oss-cn-guangzhou.aliyuncs.com/d346a708a119c24e0fe56433201c4e6f45c33ab625f2d936ef4288cae65dddbe.bin"
	resp, err := cli.OTA(context.Background(), &proto.OTAReq{Iid: iid, FirmwareUrl: url})
	if err != nil {
		logrus.Panic(err)
	}

	for {
		attrs, err := resp.Recv()
		if err != nil {
			logrus.Panic(err.Error())
		}
		logrus.Println(attrs.String())
	}
}
