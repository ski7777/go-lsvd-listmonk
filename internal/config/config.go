package config

import (
	"encoding/json"
	"io"
	"os"
)

type Config struct {
	Listmonk struct {
		Url           string `json:"url"`
		MemberIdField string `json:"memberidfield"`
		Credentials   struct {
			Username string `json:"username"`
			Password string `json:"password"`
		} `json:"credentials"`
		Lists struct {
			Managed          []uint `json:"managed"`
			OneTimeSubscribe []uint `json:"onetimesubscribe"`
		} `json:"lists"`
	} `json:"listmonk"`
	Sharepoint struct {
		Credentials struct {
			TenantId     string `json:"tenantid"`
			ClientId     string `json:"clientid"`
			ClientSecret string `json:"clientsecret"`
		} `json:"credentials"`
		DriveId      string `json:"driveid"`
		FolderItemId string `json:"folderitemid"`
	} `json:"sharepoint"`
}

func NewConfigFromBytes(bytes []byte) (c *Config, err error) {
	c = &Config{}
	err = json.Unmarshal(bytes, c)
	return
}
func NewConfigFromFile(filename string) (c *Config, err error) {
	jsonFile, err := os.Open(filename)
	defer func() {
		_ = jsonFile.Close()
	}()
	if err != nil {
		return
	}
	bytes, err := io.ReadAll(jsonFile)
	if err != nil {
		return
	}
	c, err = NewConfigFromBytes(bytes)
	return
}
