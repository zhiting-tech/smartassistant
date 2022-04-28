package main

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/v2"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/v2/demo/pkg"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	p := sdk.NewPluginServer(discover)
	sdk.Run(p)

}

func discover(ctx context.Context, devices chan<- sdk.Device) {

	for i := 0; i < 10; i++ {
		time.Sleep(time.Second)
		demo := pkg.NewDemo(uuid.New().String())
		devices <- demo
	}
}
