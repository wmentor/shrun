package entities

type ExportFileSettings struct {
	NoGoProxy   bool
	Repfactor   int
	Topology    string
	LogLevel    string
	ClusterName string
	EtcdCount   int
	PgMajor     int
}
