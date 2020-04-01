package common

import uuid "github.com/satori/go.uuid"

// GenerateID 生成任务id
func GenerateID() string {
	return uuid.NewV4().String()
}
