package tmpl

import (
	_ "embed"
)

//go:embed Dockerfile.gobuilder.tmpl
var SrcGoBuilder []byte

//go:embed Dockerfile.gotpc.tmpl
var SrcGoTpc []byte

//go:embed Dockerfile.pgbuildenv.tmpl
var SrcPgBuildEnv []byte

//go:embed Dockerfile.pgdestenv.tmpl
var SrcPgDestEnv []byte

//go:embed Dockerfile.pgdoc.tmpl
var SrcPgDoc []byte

//go:embed Dockerfile.sdmnode.tmpl
var SrcSdmNode []byte

//go:embed Dockerfile.shardman.tmpl
var SrcShardman []byte

//go:embed rc.local.tmpl
var SrcRcLocal []byte

//go:embed sdmspec.json.tmpl
var SrcSpec []byte

//go:embed env.tmpl
var EnvFile []byte

//go:embed openssl.conf
var SrcOpenSSL []byte

//go:embed prometheus.yml.tmpl
var SrcPrometheusConf []byte

//go:embed Dockerfile.prometheus.tmpl
var SrcPrometheus []byte

//go:embed Dockerfile.grafana.tmpl
var SrcGrafana []byte

//go:embed datasource.yaml.tmpl
var SrcGrafanaDatasource []byte

//go:embed dashboard.yaml.tmpl
var SrcGrafanaBoard []byte

//go:embed Main.json.tmpl
var SrcGrafanaMainBoard []byte

//go:embed Dockerfile.core.tmpl
var SrcDockerfileCore []byte
