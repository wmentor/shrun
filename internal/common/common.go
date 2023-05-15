package common

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/docker/go-units"
)

const (
	SpecFile             = "sdmspec.json"
	RcLocalFile          = "rc.local"
	DockerfileEtcd       = "Dockerfile.etcd"
	DockerfileGoBuilder  = "Dockerfile.gobuilder"
	DockerfileGoTpc      = "Dockerfile.gotpc"
	DockerfilePgBuildEnv = "Dockerfile.pgbuildenv"
	DockerfilePgDestEnv  = "Dockerfile.pgdestenv"
	DockerfilePgDoc      = "Dockerfile.pgdoc"
	DockerfileSdmNode    = "Dockerfile.sdmnode"
	DockerfileShardman   = "Dockerfile.shardman"
	DockerfileStolonInt  = "Dockerfile.stolon_int"

	ArchAmd64 = "amd64"
	ArchArm64 = "arm64"

	CopyDebugToolCmd = "COPY --from=gbuilder /go/bin/dlv $APP/bin"
	BuildDefault     = "make all"
	BuildDebug       = "make all_debug"

	PgUser = "postgres"

	ShmSize = units.MB * 512
)

var (
	ObjectPrefix = "shr"
	PgVersion    = 14
	ClusterName  = "cluster0"

	CmdBash = []string{"/bin/bash"}

	dirConfig = os.Getenv("SHRDM_CONFIG_DIR")
	dirData   = os.Getenv("SHRDM_DATA_DIR")

	WorkArch = ""
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
	os.Mkdir(GetPgDataDir(), 0755)
}

func GetDefaultArch() string {
	if runtime.GOARCH != ArchAmd64 {
		return ArchArm64
	}
	return runtime.GOARCH
}

func GetObjectPrefix() string {
	return ObjectPrefix
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

func GetPgDataDir() string {
	return filepath.Join(GetDataDir(), "pgdata")
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

func GetNodeName(num int) string {
	return fmt.Sprintf("%sn%d", GetObjectPrefix(), num)
}

func GetEtcdName(num int) string {
	return fmt.Sprintf("%se%d", GetObjectPrefix(), num)
}

func GetNodeContainerName() string {
	name := "shardman:latest"
	if GetObjectPrefix() != "shr" {
		name = GetObjectPrefix() + name
	}
	return name
}

func GetSdmNodeImageName() string {
	name := "sdmnode:latest"
	if GetObjectPrefix() != "shr" {
		name = GetObjectPrefix() + name
	}
	return name
}

func GetEnvFileName() string {
	return filepath.Join(GetConfigDir(), GetObjectPrefix()+".env")
}

func defEnvs() []string {
	return []string{
		"CLUSTER_NAME=cluster0",
		"SDM_CLUSTER_NAME=cluster0",
		"SDM_LOG_LEVEL=debug",
		"SDM_STORE_ENDPOINTS=http://" + GetEtcdName(1) + ":2379",
	}
}

func GetEnvs() []string {
	res := []string{}

	rh, err := os.Open(GetEnvFileName())
	if err != nil {
		return defEnvs()
	}
	defer rh.Close()

	br := bufio.NewReader(rh)

	for {
		str, err := br.ReadString('\n')
		if err != nil && str == "" {
			break
		}
		if str = strings.TrimSpace(str); str != "" && strings.Contains(str, "=") {
			res = append(res, str)
		}
	}

	return res
}

func GetEtcdList() ([]string, error) {
	prefix := "SDM_STORE_ENDPOINTS="
	for _, env := range GetEnvs() {
		if strings.HasPrefix(env, prefix) {
			return strings.Split(strings.TrimPrefix(env, prefix), ","), nil
		}
	}
	return nil, errors.New("etcd list not found")
}

func GetGoModDir() string {
	uinfo, err := user.Current()
	if err != nil {
		log.Fatal("get current user info error")
	}

	modDir := filepath.Join(uinfo.HomeDir, "/gopath/go1.18/pkg")

	os.MkdirAll(modDir, 0755)

	return modDir
}
