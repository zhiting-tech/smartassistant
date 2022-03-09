package role

import (
	"github.com/zhiting-tech/smartassistant/modules/device"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

func wrapRole(role entity.Role, c *gin.Context) (roleInfo, error) {

	r := roleInfo{
		RoleInfo: entity.RoleInfo{
			ID:   role.ID,
			Name: role.Name,
		},
		IsManager: role.IsManager,
	}

	// 请求数据库判断是否有改权限
	ps, err := wrapRolePermissions(role, c)
	if err != nil {
		return roleInfo{}, err
	}
	r.Permissions = &ps
	return r, nil
}

func wrapRolePermissions(role entity.Role, c *gin.Context) (ps Permissions, err error) {

	ps, err = getPermissionsWithDevices(c)
	if err != nil {
		return
	}
	curArea, _ := entity.GetAreaByID(session.Get(c).AreaID)

	wrapPermissions(role, ps.Device)
	for _, a := range ps.DeviceAdvanced.Locations {
		for _, d := range a.Devices {
			wrapPermissions(role, d.Permissions)
		}
	}
	for _, a := range ps.DeviceAdvanced.Departments {
		for _, d := range a.Devices {
			wrapPermissions(role, d.Permissions)
		}
	}
	if entity.IsHome(curArea.AreaType) {
		wrapPermissions(role, ps.Area)
		wrapPermissions(role, ps.Location)
	}else {
		wrapPermissions(role, ps.Company)
		wrapPermissions(role, ps.Department)
	}

	wrapPermissions(role, ps.Role)
	wrapPermissions(role, ps.Scene)
	return
}

// wrapPermissions 根据权限更新配置
func wrapPermissions(role entity.Role, ps []Permission) {
	for i, v := range ps {
		ps[i].Allow = entity.IsPermit(role.ID, v.Permission.Action, v.Permission.Target, v.Permission.Attribute, entity.GetDB())
	}
}

// getPermissionsWithDevices 获取所有可配置的权限(包括设备高级)
func getPermissionsWithDevices(c *gin.Context) (Permissions, error) {
	curArea, err := entity.GetAreaByID(session.Get(c).AreaID)
	if err != nil {
		return  Permissions{}, err
	}
	var locations, departments []Location
	if entity.IsHome(curArea.AreaType) {
		locations, err = getLocationsWithDevice(c)
		if err != nil {
			return Permissions{}, err
		}
	}else {
		departments, err = getDepartmentsWithDevice(c)
		if err != nil {
			return Permissions{}, err
		}
	}
	permission := Permissions{
		Device:         wrapPs(types.DevicePermission),
		DeviceAdvanced: DeviceAdvanced{Locations: locations, Departments: departments},
		Role:           wrapPs(types.RolePermission),
		Scene:          wrapPs(types.ScenePermission),
	}

	if entity.IsHome(curArea.AreaType){
		permission.Location = wrapPs(types.LocationPermission)
		permission.Area  = wrapPs(types.AreaPermission)
		return permission, nil
	}
	permission.Department = wrapPs(types.DepartmentPermission)
	permission.Company  = wrapPs(types.CompanyPermission)

	return permission, nil
}

// getPermissions 获取所有可配置的权限
func getPermissions() (Permissions, error) {

	return Permissions{
		Device:   	 wrapPs(types.DevicePermission),
		Area:     	 wrapPs(types.AreaPermission),
		Location: 	 wrapPs(types.LocationPermission),
		Role:     	 wrapPs(types.RolePermission),
		Scene:    	 wrapPs(types.ScenePermission),
		Company:  	 wrapPs(types.CompanyPermission),
		Department:  wrapPs(types.DepartmentPermission),
	}, nil
}

func wrapPs(ps []types.Permission) []Permission {
	var res []Permission

	for _, v := range ps {
		var a Permission
		a.Permission = v
		res = append(res, a)
	}
	return res
}

type Map struct {
	sync.RWMutex
	m map[int][]Device
}

func getLocationsWithDevice(c *gin.Context) (locations []Location, err error) {
	sessionUser := session.Get(c)
	devices, err := entity.GetDevices(sessionUser.AreaID)
	if err != nil {
		return
	}
	// 按区域划分
	var locationDevice Map
	locationDevice.m = make(map[int][]Device)
	var wg sync.WaitGroup
	wg.Add(len(devices))
	for _, d := range devices {
		go func(d entity.Device) {
			defer wg.Done()
			ps, e := device.Permissions(d)
			if e != nil {
				logger.Error("DevicePermissionsErr:", e.Error())
				return
			}
			dd := Device{Name: d.Name, Permissions: wrapPs(ps)}
			locationDevice.Lock()
			defer locationDevice.Unlock()
			value, ok := locationDevice.m[d.LocationID]
			if ok {
				locationDevice.m[d.LocationID] = append(value, dd)
			} else {
				locationDevice.m[d.LocationID] = []Device{dd}
			}
		}(d)
	}
	wg.Wait()
	for locationID, ds := range locationDevice.m {
		a, _ := entity.GetLocationByID(locationID)
		aa := Location{
			Name:    a.Name,
			Devices: ds,
		}
		if aa.Name == "" {
			aa.Name = "其他"
		}
		locations = append(locations, aa)
	}

	return
}

// getDepartmentsWithDevice 获取每个部门下面设备的权限
func getDepartmentsWithDevice(c *gin.Context) (department []Location, err error) {
	sessionUser := session.Get(c)
	devices, err := entity.GetDevices(sessionUser.AreaID)
	if err != nil {
		return
	}
	// 按部门划分
	var departmentDevice Map
	departmentDevice.m = make(map[int][]Device)
	var wg sync.WaitGroup
	wg.Add(len(devices))
	for _, d := range devices {
		go func(d entity.Device) {
			defer wg.Done()
			ps, e := device.Permissions(d)
			if e != nil {
				logger.Error("DevicePermissionsErr:", e.Error())
				return
			}
			dd := Device{Name: d.Name, Permissions: wrapPs(ps)}
			departmentDevice.Lock()
			defer departmentDevice.Unlock()
			value, ok := departmentDevice.m[d.DepartmentID]
			if ok {
				departmentDevice.m[d.DepartmentID] = append(value, dd)
			} else {
				departmentDevice.m[d.DepartmentID] = []Device{dd}
			}
		}(d)
	}
	wg.Wait()
	for departmentID, ds := range departmentDevice.m {
		a, _ := entity.GetDepartmentByID(departmentID)
		aa := Location{
			Name:    a.Name,
			Devices: ds,
		}
		if aa.Name == "" {
			aa.Name = "其他"
		}
		department = append(department, aa)
	}

	return
}
