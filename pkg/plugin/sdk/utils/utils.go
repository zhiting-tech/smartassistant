package utils

import (
	"strconv"
	"strings"
)

// ParserIdentity 解析设备的identity
// 把identity用-切成切片，且判断第一个元素是否childDevice (例子：childDevice-abc-4)
// 如果是childDevice，则是子设备，第二个元素为父设备的identity，第三个元素为在父子设备中的instanceId
func ParserIdentity(identity string) (isChildDevice bool, pIdentity string, instanceId int) {
	identityArr := strings.Split(identity, "-")
	if identityArr[0] == "childDevice" {
		isChildDevice = true
		pIdentity = identityArr[1]
		instanceId, _ = strconv.Atoi(identityArr[2])
	}
	return
}
