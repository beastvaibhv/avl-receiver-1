package handlers

import (
	"github.com/404minds/avl-receiver/internal/devices"
	"github.com/404minds/avl-receiver/internal/store"
)

func NewTcpHandler(datadir string) tcpHandler {
	return tcpHandler{
		connToProtocolMap:     make(map[string]devices.DeviceProtocol),
		registeredDeviceTypes: []devices.AVLDeviceType{devices.Teltonika, devices.Wanway}, // registered device types can be made configurable to enable/disable a device-type at once
		connToStoreMap:        make(map[string]store.Store),
		dataDir:               datadir,
	}

}
