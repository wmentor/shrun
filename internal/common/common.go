package common

import (
	"context"
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

	log.Printf("config dir : %s (env SHRDM_CONFIG_DIR or ~/.shrun)", dirConfig)
	log.Printf("data dir: %s (env SHRDM_DATA_DIR or ~/build)", dirData)
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

func CopyFile(ctx context.Context, src string, dest string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		log.Printf("read file %s error: %v", src, err)
		return err
	}

	if err = os.WriteFile(dest, data, 0755); err != nil {
		log.Printf("write file %s error: %v", dest, err)
		return err
	}
	log.Printf("copy %s -> %s", src, dest)

	return nil
}
