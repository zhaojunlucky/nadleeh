package workflow

const (
	NoIf int = iota
	IfMatched
	IfNotMatched
	IfEvalErr
)

const (
	NoContinueOnError int = iota
	ContinueOnErrMatched
	ContinueOnErrNotMatched
	ContinueOnErrEvalErr
)

type ActionResult struct {
	Err           error
	ReturnCode    int
	Output        string
	If            int
	ContinueOnErr int
	Skipped       bool
}

func (action *ActionResult) Set(err error, returnCode int, output string) {
	action.Err = err
	action.ReturnCode = returnCode
	action.Output = output
}

func NewActionResult(err error, returnCode int, output string) *ActionResult {
	return &ActionResult{
		Err:           err,
		ReturnCode:    returnCode,
		Output:        output,
		If:            NoIf,
		ContinueOnErr: NoContinueOnError,
	}
}

func NewEmptyActionResult() *ActionResult {
	return &ActionResult{
		Err:           nil,
		ReturnCode:    0,
		Output:        "",
		If:            NoIf,
		ContinueOnErr: NoContinueOnError,
	}
}
