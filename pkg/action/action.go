package action

type Action struct {
	Name    string
	Version string
	Env     map[string]string
	Jobs    []Job
}
