package entities

type ContainerStartSettings struct {
	Image     string
	Host      string
	Cmd       []string
	NetworkID string
	Envs      []string
}

type Container struct {
	ID     string
	Names  []string
	Status string
}
