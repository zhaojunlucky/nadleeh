package argument

import (
	"fmt"
	"github.com/akamensky/argparse"
)

func GetStringFromArg(arg argparse.Arg, required bool) (*string, error) {
	if required && !arg.GetParsed() {
		return nil, fmt.Errorf("argument %s is not provided", arg.GetLname())
	} else if !arg.GetParsed() {
		return nil, nil
	} else {
		str := arg.GetResult().(*string)
		return str, nil
	}
}

func GetInt64FromArg(arg argparse.Arg, required bool) (*int64, error) {
	if required && !arg.GetParsed() {
		return nil, fmt.Errorf("argument %s is not provided", arg.GetLname())
	} else if !arg.GetParsed() {
		return nil, nil
	} else {
		value := int64(*arg.GetResult().(*int))
		return &value, nil
	}
}

func GetBoolFromArg(arg argparse.Arg, required bool) (*bool, error) {
	return arg.GetResult().(*bool), nil
}
