package argument

import (
	"github.com/spf13/cobra"
)

// RunArgs holds arguments for the run command
type RunArgs struct {
	File        string
	Provider    string
	Check       bool
	Args        []string
	PrivateFile string
}

// WorkflowArgs holds arguments for the wf command
type WorkflowArgs struct {
	ConfigFile string
	Args       []string
}

// KeypairArgs holds arguments for the keypair command
type KeypairArgs struct {
	Name string
	Dir  string
}

// EncryptArgs holds arguments for the encrypt command
type EncryptArgs struct {
	Public string
	File   string
	Str    string
}

// CommandHandlers holds the handler functions for each command
type CommandHandlers struct {
	RunHandler     func(args *RunArgs)
	WfHandler      func(args *WorkflowArgs)
	KeypairHandler func(args *KeypairArgs)
	EncryptHandler func(args *EncryptArgs)
}

// Verbose is a global flag for verbose logging
var Verbose bool

// rootCmd is the base command
var rootCmd = &cobra.Command{
	Use:   "nadleeh",
	Short: "Nadleeh workflow",
	Long:  "Nadleeh workflow automation tool",
}

// NewNadleehCliParser creates the root cobra command with all subcommands
func NewNadleehCliParser(handlers *CommandHandlers) *cobra.Command {
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "Enable verbose log")

	addRunCmd(rootCmd, handlers)
	addEncryptCmd(rootCmd, handlers)
	addKeypairCmd(rootCmd, handlers)
	addWfCmd(rootCmd, handlers)

	return rootCmd
}
