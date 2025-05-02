package workflow

type WorkflowResult struct {
	workflowRunAction *WorkflowRunAction
}

func (r *WorkflowResult) Success() bool {
	if r.workflowRunAction.result != nil && r.workflowRunAction.result.ReturnCode != 0 {
		return false
	}
	for _, job := range r.workflowRunAction.jobActions {
		if job.result != nil && job.result.ReturnCode != 0 {
			return false
		}
	}
	return true
}

func (r *WorkflowResult) Failure() bool {
	return !r.Success()
}

func (r *WorkflowResult) Reason() string {
	if r.Failure() {
		return r.workflowRunAction.result.Err.Error()
	}
	return ""
}

type WorkflowJobResult struct {
	jobAction *JobAction
}

func (r *WorkflowJobResult) Success() bool {
	if r.jobAction.result != nil && r.jobAction.result.ReturnCode != 0 {
		return false
	}
	for _, step := range r.jobAction.stepActions {
		if step.result != nil && step.result.ReturnCode != 0 {
			return false
		}
	}
	return true
}

func (r *WorkflowJobResult) Failure() bool {
	return !r.Success()
}
