package entities

type ContainerStartSettings struct {
	Image       string
	Host        string
	Cmd         []string
	NetworkID   string
	Envs        []string
	MemoryLimit int64
	CPU         float64
}

type Container struct {
	ID     string
	Names  []string
	Status string
}
