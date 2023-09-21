package psql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wmentor/shrun/internal/common"
)

var (
	ErrInvalidContainerManager = errors.New("invalid container manager")
	ErrInvalidHostname         = errors.New("invalid hostname")
	ErrCommandFailed           = errors.New("psql failed")
)

type Case struct {
	hostname string
	port     int
	manager  ContainerManager
}

func NewCase(manager ContainerManager, hostname string) (*Case, error) {
	if manager == nil || manager == ContainerManager(nil) {
		return nil, ErrInvalidContainerManager
	}

	if strings.TrimSpace(hostname) == "" {
		return nil, ErrInvalidHostname
	}

	c := &Case{
		manager:  manager,
		hostname: hostname,
	}

	return c, nil
}

func (c *Case) WithPort(port int) *Case {
	if port > 0 && port <= 0xffff {
		c.port = port
	}
	return c
}

func (c *Case) Exec(ctx context.Context) error {
	opts, err := c.getParams()
	if err != nil {
		return err
	}

	connstr := c.makeConnstr(opts)

	return c.manager.ShellCommand(ctx, c.hostname, "postgres", "", []string{"psql", "-d", connstr})
}

type params struct {
	port       int
	username   string
	password   string
	cert       string
	key        string
	rootCert   string
	verifyMode string
}

type specData struct {
	InitialPort int    `json:"PGsInitialPort"`
	Username    string `json:"PgSuUsername"`
	Password    string `json:"PgSuPassword"`
	Cert        string `json:"PgSuSSLCert"`
	Key         string `json:"PgSuSSLKey"`
	VerifyMode  string `json:"PgSSLMode"`
	RootCert    string `json:"PgSSLRootCert"`
}

func (c *Case) makeConnstr(opts params) string {
	if opts.cert != "" {
		return fmt.Sprintf("dbname=postgres host=%s port=%d user=%s sslcert=%s sslkey=%s sslrootcert=%s sslmode=%s",
			c.hostname, opts.port, opts.username, opts.cert, opts.key, opts.rootCert, opts.verifyMode)
	}
	return fmt.Sprintf("dbname=postgres host=%s password=%s port=%d user=%s", c.hostname, opts.password, opts.port, opts.username)
}

func (c *Case) getParams() (params, error) {
	var opts params

	specFile := filepath.Join(common.GetConfigDir(), "sdmspec.json")

	rf, err := os.Open(specFile)
	if err != nil {
		return opts, fmt.Errorf("open %s file error: %w", specFile, err)
	}
	defer rf.Close()

	var resObj specData

	if err = json.NewDecoder(rf).Decode(&resObj); err != nil {
		return opts, fmt.Errorf("decode %s file error: %w", specFile, err)
	}

	opts.port = resObj.InitialPort
	opts.username = resObj.Username
	opts.password = resObj.Password
	opts.cert = resObj.Cert
	opts.key = resObj.Key
	opts.rootCert = resObj.RootCert
	opts.verifyMode = resObj.VerifyMode

	if c.port != 0 {
		opts.port = c.port
	}

	if opts.port == 0 {
		opts.port = common.DefaultPgPort
	}

	return opts, nil
}
