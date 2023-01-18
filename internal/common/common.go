package common

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

const (
	SpecFile             = "sdmspec.json"
	RcLocalFile          = "rc.local"
	DockerfileEtcd       = "Dockerfile.etcd"
	DockerfileGoBuilder  = "Dockerfile.gobuilder"
	DockerfilePgBuildEnv = "Dockerfile.pgbuildenv"
	DockerfilePgDestEnv  = "Dockerfile.pgdestenv"
	DockerfilePgDoc      = "Dockerfile.pgdoc"
	DockerfileSdmNode    = "Dockerfile.sdmnode"
	DockerfileShardman   = "Dockerfile.shardman"
)

var (
	objectPrefix = "shr"

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

	os.Mkdir(GetVolumeDir(), 0755)
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

func GetVolumeDir() string {
	return filepath.Join(GetDataDir(), "mntdata")
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

func GetEtcdList() ([]string, error) {
	srcFile := filepath.Join(GetConfigDir(), "Dockerfile.shardman")
	rh, err := os.Open(srcFile)
	if err != nil {
		log.Printf("open file %s error: %v", srcFile, err)
		return nil, err
	}
	defer rh.Close()

	prefix := "ARG SDM_STORE_ENDPOINTS="

	br := bufio.NewReader(rh)
	for {
		str, err := br.ReadString('\n')
		if err != nil && str == "" {
			return nil, ErrNotFound
		}

		str = strings.TrimSpace(str)
		if strings.HasPrefix(str, prefix) {
			str = strings.TrimPrefix(str, prefix)
			return strings.Split(str, ","), nil
		}
	}
}

func GetClusterName() (string, error) {
	srcFile := filepath.Join(GetConfigDir(), "Dockerfile.shardman")
	rh, err := os.Open(srcFile)
	if err != nil {
		log.Printf("open file %s error: %v", srcFile, err)
		return "", err
	}
	defer rh.Close()

	prefix := "ARG SDM_CLUSTER_NAME="

	br := bufio.NewReader(rh)
	for {
		str, err := br.ReadString('\n')
		if err != nil && str == "" {
			return "", ErrNotFound
		}

		str = strings.TrimSpace(str)
		if strings.HasPrefix(str, prefix) {
			str = strings.TrimPrefix(str, prefix)
			return str, nil
		}
	}
}

func GetLogLevel() (string, error) {
	srcFile := filepath.Join(GetConfigDir(), "Dockerfile.shardman")
	rh, err := os.Open(srcFile)
	if err != nil {
		log.Printf("open file %s error: %v", srcFile, err)
		return "", err
	}
	defer rh.Close()

	prefix := "ARG SDM_LOG_LEVEL="

	br := bufio.NewReader(rh)
	for {
		str, err := br.ReadString('\n')
		if err != nil && str == "" {
			return "", ErrNotFound
		}

		str = strings.TrimSpace(str)
		if strings.HasPrefix(str, prefix) {
			str = strings.TrimPrefix(str, prefix)
			return str, nil
		}
	}
}

func GetNodeName(num int) string {
	return fmt.Sprintf("%sn%d", GetObjectPrefix(), num)
}

func GetEtcdName(num int) string {
	return fmt.Sprintf("%se%d", GetObjectPrefix(), num)
}
