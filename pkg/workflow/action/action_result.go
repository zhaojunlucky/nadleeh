package workflow

type ActionResult struct {
	Err        error
	ReturnCode int
	Output     string
}

func NewActionResult(err error, returnCode int, output string) *ActionResult {
	return &ActionResult{
		Err:        err,
		ReturnCode: returnCode,
		Output:     output,
	}
}
