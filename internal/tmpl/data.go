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
