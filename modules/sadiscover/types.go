package sadiscover

type Info struct {
	Model string `json:"model"`
	SwVer string `json:"sw_ver"`
	HwVer string `json:"hw_ver"`
	Port  int    `json:"port"`
	SaID  string `json:"sa_id"`
}

type Result struct {
	ID int  `json:"id"`
	Re Info `json:"result"`
}

type Protocol struct {
}
