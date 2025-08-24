package core

import "strings"

var (
	Fail     = "Fail"
	Pass     = "Pass"
	NotStart = "NotStart"
	Running  = "Running"
	Skipped  = "Skipped"
)

type RunnableStatus struct {
	name          string
	rType         string
	status        string
	errs          []string
	childs        []*RunnableStatus
	childMap      map[string]*RunnableStatus
	ContinueOnErr bool
}

func (r *RunnableStatus) Status() string {
	return r.status
}

func (r *RunnableStatus) errors() []string {
	var allErrs []string
	allErrs = append(allErrs, r.errs...)
	for _, child := range r.childs {
		allErrs = append(allErrs, child.errors()...)
	}
	return allErrs
}

func (r *RunnableStatus) Reason() string {
	return strings.Join(r.errors(), "\n")
}

func (r *RunnableStatus) FutureStatus() string {
	if r.status == Fail && !r.ContinueOnErr {
		return Fail
	}
	for _, child := range r.childs {
		if child.FutureStatus() == Fail {
			return Fail
		}
	}
	return Pass
}

func (r *RunnableStatus) Finish(errs ...error) {
	if len(errs) > 0 {
		r.status = Fail
	} else {
		r.status = Pass
	}
	for _, err := range errs {
		r.errs = append(r.errs, err.Error())
	}
}

func (r *RunnableStatus) Skipped() {
	r.status = Skipped
}

func (r *RunnableStatus) Start() {
	r.status = Running
}

func (r *RunnableStatus) AddChild(child *RunnableStatus) {
	r.childs = append(r.childs, child)
	if len(child.name) > 0 {
		r.childMap[child.name] = child
	}
}

func (r *RunnableStatus) GetChild(name string) *RunnableStatus {
	return r.childMap[name]
}

func (r *RunnableStatus) GetChildByIndex(i int) *RunnableStatus {
	return r.childs[i]
}

func NewRunnableStatus(name, rType string) *RunnableStatus {
	return &RunnableStatus{
		name:     name,
		rType:    rType,
		status:   NotStart,
		errs:     nil,
		childs:   nil,
		childMap: make(map[string]*RunnableStatus),
	}
}
