package common

import (
	"log"
	"os"
	"os/user"
	"path/filepath"
)

var (
	objectPrefix = "shrdm_"

	dirConfig = os.Getenv("SHRDM_CONFIG_DIR")
	dirData   = os.Getenv("SHRDM_DATA_DIR")
)

func init() {
	uinfo, err := user.Current()
	if err != nil {
		log.Fatal("get current user info error")
	}

	if dirConfig == "" {
		dirConfig = filepath.Join(uinfo.HomeDir, ".shrun")
	}

	if finfo, err := os.Lstat(dirConfig); err != nil {
		if os.IsNotExist(err) {
			if err := os.Mkdir(dirConfig, 0755); err != nil {
				log.Fatalf("create directory %s error: %v", dirConfig, err.Error())
			}
		} else {
			panic(err)
		}

	} else {
		if !finfo.IsDir() {
			log.Fatalf("%s must be dir", dirConfig)
		}
	}

	if dirData == "" {
		dirData = filepath.Join(uinfo.HomeDir, "build")
	}

	log.Printf("config dir: %s", dirConfig)
	log.Printf("data dir: %s", dirData)
}

func GetObjectPrefix() string {
	return objectPrefix
}

func GetConfigDir() string {
	return dirConfig
}

func GetDataDir() string {
	return dirData
}
