package device

import (
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

func GetOrCreateUUID() string {
	appData := os.Getenv("APPDATA")
	uuidFile := filepath.Join(appData, "ka-ping-uuid.txt")
	data, err := os.ReadFile(uuidFile)
	if err == nil && len(data) > 0 {
		return string(data)
	}
	id := uuid.New().String()
	_ = os.WriteFile(uuidFile, []byte(id), 0644)
	return id
}
