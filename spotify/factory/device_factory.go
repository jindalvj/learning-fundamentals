package factory

import (
	"app/device"
	"app/enum"
	"app/external"
)

func CreateDevice(deviceType enum.DeviceType) device.IAudioOutputDevice {
	switch deviceType {
	case enum.Bluetooth:
		return device.NewBluetoothSpeakerAdapter(&external.BluetoothSpeakerAPI{})
	case enum.Wired:
		return device.NewWiredSpeakerAdapter(&external.WiredSpeakerAPI{})
	default: // Headphones
		return device.NewHeadphonesAdapter(&external.HeadphonesAPI{})
	}
}
