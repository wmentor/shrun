package source

import (
	_ "embed"
)

//go:embed Dockerfile.gobuilder
var SrcGoBuilder []byte
