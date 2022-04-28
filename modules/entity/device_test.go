package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zhiting-tech/smartassistant/modules/types"
)

func TestAddDevice(t *testing.T) {
	ast := assert.New(t)

	var testSADevice = Device{
		ID:         1,
		IID:        "666",
		Model:      types.SaModel,
		PluginID:   "1",
		LocationID: 1,
		OwnerID:    1,
	}

	var testDevice2 = Device{
		ID:         2,
		IID:        "678",
		LocationID: 1,
		OwnerID:    2,
	}

	var errorIIDDevice = Device{
		ID:         3,
		IID:        "666",
		LocationID: 1,
		OwnerID:    2,
	}

	err := AddDevice(&testSADevice, GetDB())
	ast.NoError(err, "add device error: %v", err)

	err = AddDevice(&testDevice2, GetDB())
	ast.NoError(err, "add device error: %v", err)

	err = AddDevice(&testSADevice, GetDB())
	ast.Error(err, "add exist device error")

	err = AddDevice(&errorIIDDevice, GetDB())
	ast.Error(err, "add error device error")
}

func TestCheckSaDeviceCreator(t *testing.T) {
	ast := assert.New(t)

	err := CheckSaDeviceCreator(1)
	ast.NoError(err, "check sa device creator error: %v", err)
	//err = CheckSaDeviceCreator(2)
	//ast.Error(err, "check sa device creator error")
}

func TestIsSAOwner(t *testing.T) {
	ast := assert.New(t)

	ok := IsOwner(1)
	ast.True(ok, "is Sa Creator error")

	ok = IsOwner(2)
	ast.False(ok, "is Sa Creator error")
}

func TestGetDevice(t *testing.T) {
	ast := assert.New(t)

	const noExistDeviceID = 999
	const noExistIID = "999"

	devices, err := GetDevices()
	ast.NoError(err, "get devices error: %v", err)
	ast.NotEmpty(devices)

	device, err := GetSaDevice()
	ast.NoError(err, "get sa device error: %v", err)
	ast.NotEmpty(device)

	device, err = GetDeviceByID(1)
	ast.NoError(err, "get device by id error: %v", err)
	ast.NotEmpty(device)
	device, err = GetDeviceByID(noExistDeviceID)
	ast.Error(err, "get device by id error")
	ast.Empty(device)

	devices, err = GetDevicesByLocationID(1)
	ast.NoError(err, "get devices by location id: %v", err)
	ast.NotEmpty(devices)

	device, err = GetPluginDevice("666")
	ast.NoError(err, "get device by id error: %v", err)
	ast.NotEmpty(device)
	device, err = GetPluginDevice(noExistIID)
	ast.Error(err, "get device by id error")
	ast.Empty(device)
}

func TestUpdateDevice(t *testing.T) {
	ast := assert.New(t)

	const noExistDeviceID = 999
	const newOwnerID = 2

	var updateDevice = Device{
		ID:         2,
		IID:        "678",
		LocationID: 1,
		OwnerID:    newOwnerID,
	}

	err := UpdateDevice(updateDevice.ID, updateDevice)
	ast.NoError(err, "update device error: %v", err)
	device, _ := GetDeviceByID(updateDevice.ID)
	ast.Equal(device.OwnerID, newOwnerID)

	err = UpdateDevice(noExistDeviceID, updateDevice)
	ast.Error(err, "update device error")
}

func TestUnBindLocationDevice(t *testing.T) {
	ast := assert.New(t)

	err := UnBindLocationDevice(1)
	ast.NoError(err, "unbind location device error: %v", err)
	device, _ := GetDeviceByID(1)
	ast.Equal(device.LocationID, 0)

	err = UnBindLocationDevices(1)
	ast.NoError(err, "unbind location device error: %v", err)
	device, _ = GetDeviceByID(2)
	ast.Equal(device.LocationID, 0)
}

func TestDelDevice(t *testing.T) {
	ast := assert.New(t)

	err := DelDeviceByID(2)
	ast.NoError(err, "delete device by id error: %v", err)
	device, _ := GetDeviceByID(2)
	ast.Empty(device)

	err = DelDevicesByPlgID("1")
	ast.NoError(err, "delete device by plugin id error: %v", err)
	device, _ = GetDeviceByID(1)
	ast.Empty(device)
}
