package entities

type ContainerStartSettings struct {
	Image       string
	Host        string
	Cmd         []string
	NetworkID   string
	Ports       []string
	Envs        []string
	MemoryLimit int64
	CPU         float64
	MountData   bool
	Debug       bool
}

type Container struct {
	ID     string
	Names  []string
	Status string
}
