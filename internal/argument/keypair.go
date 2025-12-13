package argument

import (
	"github.com/spf13/cobra"
)

func addKeypairCmd(rootCmd *cobra.Command, handlers *CommandHandlers) {
	keypairArgs := &KeypairArgs{}

	keypairCmd := &cobra.Command{
		Use:   "keypair",
		Short: "Generate key pair",
		Run: func(cmd *cobra.Command, args []string) {
			if handlers != nil && handlers.KeypairHandler != nil {
				handlers.KeypairHandler(keypairArgs)
			}
		},
	}

	keypairCmd.Flags().StringVar(&keypairArgs.Name, "name", "", "The name of the key pair")
	keypairCmd.Flags().StringVar(&keypairArgs.Dir, "dir", "", "The directory to save the generated key pair")

	_ = keypairCmd.MarkFlagRequired("name")
	_ = keypairCmd.MarkFlagRequired("dir")

	rootCmd.AddCommand(keypairCmd)
}
