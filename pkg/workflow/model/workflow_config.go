package workflow

type WorkflowConfig struct {
	Workflow string            `yaml:"workflow"`
	Provider string            `yaml:"provider"`
	Args     map[string]string `yaml:"args"`
	Private  string            `yaml:"private"`
}
