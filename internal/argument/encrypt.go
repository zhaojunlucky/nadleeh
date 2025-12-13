package argument

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func addEncryptCmd(rootCmd *cobra.Command, handlers *CommandHandlers) {
	encryptArgs := &EncryptArgs{}

	encryptCmd := &cobra.Command{
		Use:   "encrypt",
		Short: "Encrypt the given string data or file",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Validate public key file
			fi, err := os.Stat(encryptArgs.Public)
			if err != nil {
				return err
			}
			if fi.IsDir() {
				return fmt.Errorf("public file must be a valid file")
			}
			// Validate file if provided
			if encryptArgs.File != "" {
				fi, err := os.Stat(encryptArgs.File)
				if err != nil {
					return err
				}
				if fi.IsDir() {
					return fmt.Errorf("file must be a file")
				}
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			if handlers != nil && handlers.EncryptHandler != nil {
				handlers.EncryptHandler(encryptArgs)
			}
		},
	}

	encryptCmd.Flags().StringVar(&encryptArgs.Public, "public", "", "The public key file")
	encryptCmd.Flags().StringVarP(&encryptArgs.File, "file", "f", "", "The file to encrypt")
	encryptCmd.Flags().StringVarP(&encryptArgs.Str, "str", "s", "", "The string to encrypt")

	_ = encryptCmd.MarkFlagRequired("public")

	rootCmd.AddCommand(encryptCmd)
}
