package types

import "strings"

type Scope struct {
	Scope string
	Info  string
}

var (
	// ScopeAll 拥有所有权限的特殊 Scope
	ScopeAll = Scope{"all", "获取所有权限"}

	ScopeUser         = Scope{"user", "获取你的登录状态"}
	ScopeArea         = Scope{"area", "获取你的家庭信息"}
	ScopeDevice       = Scope{"device", "设备相关权限"}
	ScopePlugin       = Scope{"plugin", "插件相关权限"}
	ScopeGetTokenBySC = Scope{"get_token_by_sc", "通过sc找回用户token"}
)

func WithScopes(scopes ...Scope) string {
	var ss []string
	for _, s := range scopes {
		ss = append(ss, s.Scope)
	}
	return strings.Join(ss, ",")
}

// Scopes 所有可用的 scope
var Scopes = map[string]Scope{
	ScopeUser.Scope:         ScopeUser,
	ScopeArea.Scope:         ScopeArea,
	ScopeDevice.Scope:       ScopeDevice,
	ScopePlugin.Scope:       ScopePlugin,
	ScopeGetTokenBySC.Scope: ScopeGetTokenBySC,
}

// ExtensionScope 扩展的 scope TODO 将扩展和插件作为 client 管理起来
var ExtensionScope = []Scope{
	ScopeUser, ScopeArea,
}
