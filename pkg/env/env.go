package env

type Env interface {
	Get(key string) string
	Set(key, value string)
	SetAll(envs map[string]string)

	GetAll() map[string]string

	Expand(s string) string
}
