package workflow

import "errors"

type Job struct {
	Name  string
	Steps []Step
	Env   map[string]string
}

func (job *Job) Validate() error {
	var jobErrors []error

	for _, step := range job.Steps {
		jobErrors = append(jobErrors, step.Validate())
	}

	if len(jobErrors) > 0 {
		return errors.Join(jobErrors...)
	}
	return nil
}

func (job *Job) HasSteps() bool {
	return len(job.Steps) > 0
}
