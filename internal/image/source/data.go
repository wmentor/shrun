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
