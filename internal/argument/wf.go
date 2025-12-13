package argument

import (
	"fmt"
	"regexp"

	"github.com/spf13/cobra"
)

func addWfCmd(rootCmd *cobra.Command, handlers *CommandHandlers) {
	wfArgs := &WorkflowArgs{}

	re := regexp.MustCompile(`^[a-zA-Z0-9_]+=.*$`)

	wfCmd := &cobra.Command{
		Use:   "wf <config-file>",
		Short: "Run the given workflow config file",
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			for _, arg := range wfArgs.Args {
				if !re.MatchString(arg) {
					return fmt.Errorf("invalid argument %s", arg)
				}
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			wfArgs.ConfigFile = args[0]
			if handlers != nil && handlers.WfHandler != nil {
				handlers.WfHandler(wfArgs)
			}
		},
	}

	wfCmd.Flags().StringArrayVarP(&wfArgs.Args, "arg", "a", nil, "Arguments variables")

	rootCmd.AddCommand(wfCmd)
}
