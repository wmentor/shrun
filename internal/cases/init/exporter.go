package init

import (
	"github.com/wmentor/shrun/internal/entities"
)

type Exporter interface {
	ExportFiles(settings entities.ExportFileSettings) error
}
