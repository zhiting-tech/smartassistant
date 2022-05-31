package area

import (
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"strconv"
	"testing"

	"github.com/zhiting-tech/smartassistant/modules/api/test"
)

func TestArea(t *testing.T) {
	// 添加area记录
	test.CreateRecord(&entity.Area{Name: "demo"})
	test.CreateRecord(&entity.Area{Name: "demo2"})
	area := test.GetAreas()
	test.CreateRecord(&entity.Device{Name: "demo2", Model: types.SaModel, AreaID: area[1].ID})

	areaID := strconv.FormatUint(area[0].ID, 10)
	areaID2 := strconv.FormatUint(area[1].ID, 10)

	userCases := []test.ApiTestCase{
		{
			Method: "GET",
			Path:   "/areas",
			Status: 0,
			IsArray: []string{
				"data.areas",
			},
		},
		{
			Method: "PUT",
			Path:   "/areas/3",
			Body:   "{\n  \"name\": \"areaNotExist\" \n}",
			Status: status.Deny,
		},
		{
			Method: "PUT",
			Path:   "/areas/" + areaID,
			Body:   "{\n  \"name\": \"areaExist\" \n}",
			Status: 0,
		},
		// 前端传入空参
		{
			Method: "PUT",
			Path:   "/areas/" + areaID,
			Body:   "{\n  \"name\": \"\" \n}",
			Status: status.AreaNameInputNilErr,
		},
		// 前端传入参数长度超出限制
		{
			Method: "PUT",
			Path:   "/areas/" + areaID,
			Body:   "{\n  \"name\": \"ashdoiauhfuioasodiugfiuagoiuahoiudgfyausgtdoyftguyatgfydsoaihdushdiuf\" \n}",
			Status: status.AreaNameLengthLimit,
		},
		{
			Method: "DELETE",
			Path:   "/areas/" + areaID,
			Body:   "{\"areas\": \"" + areaID + "\"}",
			Status: status.Deny,
		},
		{
			Method: "GET",
			Path:   "/areas/" + areaID,
			Status: 0,
		},
		{
			Method: "GET",
			Path:   "/areas/" + areaID,
			Status: 0,
		},
		{
			Method: "DELETE",
			Path:   "/areas/" + areaID + "/users/1",
			Status: 0,
		},
		// 用户已退出当前area
		{
			Method: "DELETE",
			Path:   "/areas/" + areaID + "/users/2",
			Status: status.InvalidUserCredentials,
		},
	}

	// 管理员角色用户的测试
	test.RunApiTest(t, RegisterAreaRouter, userCases, test.WithRoles("管理员"), test.WithAreas(area[0].ID))

	adminCases := []test.ApiTestCase{
		{
			Method: "GET",
			Path:   "/areas",
			Status: 0,
			IsArray: []string{
				"data.areas",
			},
		},
		// 成员角色用户没有权限修改area
		{
			Method: "PUT",
			Path:   "/areas/" + areaID2,
			Body:   "{\"name\": \"area\"}",
			Status: status.Deny,
		},
		{
			Method: "GET",
			Path:   "/areas/" + areaID2,
			Status: 0,
		},
		{
			Method: "DELETE",
			Path:   "/areas/" + areaID2,
			Body:   "{\"is_del_cloud_disk\": false}",
			Status: status.Deny,
		},
		// 用户已退出当前area
		{
			Method: "DELETE",
			Path:   "/areas/" + areaID2 + "/users/2",
			Status: status.Deny,
		},
	}

	// 成员角色用户的测试
	test.RunApiTest(t, RegisterAreaRouter, adminCases, test.WithRoles("成员"), test.WithAreas(area[1].ID))
}

func TestMain(m *testing.M) {
	test.InitApiTest(m)
}
