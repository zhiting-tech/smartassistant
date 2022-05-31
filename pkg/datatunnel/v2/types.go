package datatunnel

import "github.com/zhiting-tech/smartassistant/pkg/datatunnel/v2/control"

const (
	methodAuthenticate    = "Authenticate"
	methodRegisterService = "RegisterService"
	notifyNewConnection   = "NewConnection"
)

const (
	clientVersion = 1
)

const (
	InvalidSAIDOrSAKey = iota + control.CustomCodeStart + 1
	DuplicateRegisterService
)
