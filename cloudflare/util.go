package cloudflare

import (
	"github.com/pkg/errors"
)

func FindDevice(devices []BoundDevice, deviceId string) (*BoundDevice, error) {
	for i := range devices {
		if devices[i].Id == deviceId {
			return &devices[i], nil
		}
	}
	return nil, errors.New("device not found in list")
}
