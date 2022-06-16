package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenerateGroupAndCustomDeviceID(t *testing.T) {
	groupID := "alertmanager"
	deviceID := "alertmanager"
	group, customDevice := GenerateGroupAndCustomDeviceID(groupID, deviceID)
	assert.Equal(t, "CUSTOM_DEVICE_GROUP-E1ABC2CBF8723322", group)
	assert.Equal(t, "CUSTOM_DEVICE-EBFD2154C71FC3F7", customDevice)

	groupID = "Alertmanager OCP4"
	group, customDevice = GenerateGroupAndCustomDeviceID(groupID, deviceID)
	assert.Equal(t, "CUSTOM_DEVICE_GROUP-D46A0C2A46644A07", group)
	assert.Equal(t, "CUSTOM_DEVICE-4264CFC18F3CBC35", customDevice)
}
