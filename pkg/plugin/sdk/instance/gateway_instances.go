package instance

type GatewayServices struct {
	Services []interface{}
}

func (s GatewayServices) InstanceName() string {
	return "gateway_services"
}
