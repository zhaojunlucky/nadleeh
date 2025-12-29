package argument

import (
	"fmt"
	"os"
	"regexp"

	"github.com/spf13/cobra"
)

func addRunCmd(rootCmd *cobra.Command, handlers *CommandHandlers) {
	runArgs := &RunArgs{}

	re := regexp.MustCompile(`^[a-zA-Z0-9_]+(=.*)?$`)

	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run the given workflow file",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Validate args
			for _, arg := range runArgs.Args {
				if !re.MatchString(arg) {
					return fmt.Errorf("invalid argument %s", arg)
				}
			}
			// Validate private file if provided
			if runArgs.PrivateFile != "" {
				fi, err := os.Stat(runArgs.PrivateFile)
				if err != nil {
					return err
				}
				if fi.IsDir() {
					return fmt.Errorf("private file must be a valid file")
				}
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			if handlers != nil && handlers.RunHandler != nil {
				handlers.RunHandler(runArgs)
			}
		},
	}

	runCmd.Flags().StringVarP(&runArgs.File, "file", "f", "", "Run the workflow file")
	runCmd.Flags().StringVarP(&runArgs.Provider, "provider", "p", "", "The workflow provider (e.g., github)")
	runCmd.Flags().BoolVarP(&runArgs.Check, "check", "c", false, "Only check the workflow")
	runCmd.Flags().BoolVar(&runArgs.Usage, "usage", false, "Show usage")
	runCmd.Flags().StringArrayVarP(&runArgs.Args, "arg", "a", nil, "Arguments variables")
	runCmd.Flags().StringVar(&runArgs.PrivateFile, "private", "", "Private key file to decrypt the encrypted data")

	_ = runCmd.MarkFlagRequired("file")

	runCmd.Flags().Lookup("provider").NoOptDefVal = "github"

	rootCmd.AddCommand(runCmd)
}
