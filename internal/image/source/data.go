package source

import (
	_ "embed"
)

//go:embed Dockerfile.gobuilder
var SrcGoBuilder []byte

//go:embed Dockerfile.pgbuildenv
var SrcPgBuildEnv []byte

//go:embed Dockerfile.pgdestenv
var SrcPgDestEnv []byte

//go:embed Dockerfile.sdmnode
var SrcSdmNode []byte

//go:embed Dockerfile.shardman
var SrcShardman []byte

//go:embed rc.local
var SrcRcLocal []byte

//go:embed sdmspec.json
var SrcSpec []byte
