package core

type Validator interface {
	Precheck() error
}
