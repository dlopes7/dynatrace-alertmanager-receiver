package utils

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/twmb/murmur3"
	"os"
	"strconv"
	"strings"
)

func GenerateGroupAndCustomDeviceID(groupID string, deviceID string) (string, string) {

	// customDeviceId, customDeviceGroupId -> c0c06b85c0d34e60, c4130405a7d2b02
	// alertmanager, alertmanager -> E1ABC2CBF8723322, EBFD2154C71FC3F7

	namespace := ""
	namespaceLenBigEndian := make([]byte, 8)
	binary.BigEndian.PutUint32(namespaceLenBigEndian, uint32(len(namespace)))

	groupLenBigEndian := make([]byte, 8)
	binary.BigEndian.PutUint32(groupLenBigEndian, uint32(len(groupID)))

	var groupIdBytes []byte
	groupIdBytes = append(groupIdBytes, []byte(namespace)...)
	groupIdBytes = append(groupIdBytes, namespaceLenBigEndian[:len(namespaceLenBigEndian)-4]...)
	groupIdBytes = append(groupIdBytes, []byte(groupID)...)
	groupIdBytes = append(groupIdBytes, groupLenBigEndian[:len(groupLenBigEndian)-4]...)

	dtGroupID := dtMurMur3(groupIdBytes)
	dtGroupIDUint64, _ := strconv.ParseUint(dtGroupID, 16, 64)

	dtGroupIDBigEndian := make([]byte, 8)
	binary.BigEndian.PutUint64(dtGroupIDBigEndian, dtGroupIDUint64)

	deviceIDLenBigEndian := make([]byte, 8)
	binary.BigEndian.PutUint32(deviceIDLenBigEndian, uint32(len(deviceID)))

	var customDeviceBytes []byte
	customDeviceBytes = append(customDeviceBytes, dtGroupIDBigEndian...)
	customDeviceBytes = append(customDeviceBytes, []byte(deviceID)...)
	customDeviceBytes = append(customDeviceBytes, deviceIDLenBigEndian[:len(deviceIDLenBigEndian)-4]...)

	dtCustomDeviceID := dtMurMur3(customDeviceBytes)

	return fmt.Sprintf("CUSTOM_DEVICE_GROUP-%s", dtGroupID), fmt.Sprintf("CUSTOM_DEVICE-%s", dtCustomDeviceID)
}

func dtMurMur3(data []byte) string {

	h1, _ := murmur3.SeedSum128(0, 0, data)
	return strings.TrimLeft(fmt.Sprintf("%X", h1), "0")

}

func GetTempDir() string {
	tmpDir := fmt.Sprintf("%s/dynatrace-receiver", os.TempDir())
	if os.Getenv("WEBHOOK_LOG_FOLDER") != "" {
		tmpDir = os.Getenv("WEBHOOK_LOG_FOLDER")
	}
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		log.WithFields(log.Fields{"path": tmpDir}).Info("Creating temporary directory")
		err := os.MkdirAll(tmpDir, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}

	return tmpDir
}

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func Hash(data string) string {
	fullHash := fmt.Sprintf("%x", md5.Sum([]byte(data)))
	half := fullHash[:len(fullHash)-16]
	return half
}
