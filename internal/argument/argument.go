package argument

import (
	"fmt"
	"os"
	"regexp"

	"github.com/akamensky/argparse"
)

func NewNadleehCliParser() *argparse.Parser {
	parser := argparse.NewParser("nadleeh", "Nadleeh workflow")
	parser.String("h", "help", &argparse.Options{
		Required: false,
		Validate: nil,
		Help:     "Display help",
		Default:  nil,
	})

	parser.Flag("v", "verbose", &argparse.Options{
		Required: false,
		Validate: nil,
		Help:     "Enable verbose log",
		Default:  nil,
	})

	addRunCmd(parser)
	addEncryptCmd(parser)
	addGenerateKeyPairCmd(parser)
	addWorkflowCmd(parser)

	return parser
}

func addWorkflowCmd(parser *argparse.Parser) {
	runCmd := parser.NewCommand("wf", "Run the given workflow config file")
	runCmd.StringPositional(&argparse.Options{
		Required: true,
		Validate: nil,
		Help:     "Run the workflow config file",
		Default:  nil,
	})

	re := regexp.MustCompile(`^[a-zA-Z0-9_]+=.*$`)

	runCmd.StringList("a", "arg", &argparse.Options{
		Required: false,
		Validate: func(args []string) error {
			if len(args) <= 0 {
				return nil
			}
			for _, arg := range args {
				if !re.MatchString(arg) {
					return fmt.Errorf("invalid argment %s", arg)
				}
			}
			return nil
		},
		Help:    "Arguments variables",
		Default: nil,
	})
}

func addRunCmd(parser *argparse.Parser) {
	runCmd := parser.NewCommand("run", "Run the given workflow file")
	runCmd.String("f", "file", &argparse.Options{
		Required: true,
		Validate: nil,
		Help:     "Run the workflow file",
		Default:  nil,
	})

	runCmd.String("p", "provider", &argparse.Options{
		Required: false,
		Validate: nil,
		Help:     "The workflow provider",
		Default:  nil,
	})

	runCmd.Flag("c", "check", &argparse.Options{
		Required: false,
		Validate: nil,
		Help:     "Only check the workflow",
		Default:  nil,
	})

	re := regexp.MustCompile(`^[a-zA-Z0-9_]+=.*$`)

	runCmd.StringList("a", "arg", &argparse.Options{
		Required: false,
		Validate: func(args []string) error {
			if len(args) <= 0 {
				return nil
			}
			for _, arg := range args {
				if !re.MatchString(arg) {
					return fmt.Errorf("invalid argment %s", arg)
				}
			}
			return nil
		},
		Help:    "Arguments variables",
		Default: nil,
	})

	runCmd.String("", "private", &argparse.Options{
		Required: false,
		Validate: func(args []string) error {
			if len(args) <= 0 || len(args[0]) <= 0 {
				return nil
			}
			fi, err := os.Stat(args[0])
			if err != nil {
				return err
			}
			if fi.IsDir() {
				return fmt.Errorf("private file must be a valid file")
			}
			return nil
		},
		Help:    "Private key file to decrypt the encrypted data",
		Default: nil,
	})
}

func addGenerateKeyPairCmd(parser *argparse.Parser) {
	keyPairCmd := parser.NewCommand("keypair", "Generate key pair")
	keyPairCmd.String("", "name", &argparse.Options{
		Required: true,
		Validate: nil,
		Help:     "The name of the key pair",
		Default:  nil,
	})

	keyPairCmd.String("", "dir", &argparse.Options{
		Required: true,
		Validate: nil,
		Help:     "The directory to save the generated key pair",
		Default:  nil,
	})
}

func addEncryptCmd(parser *argparse.Parser) {
	encryptCmd := parser.NewCommand("encrypt", "encrypt the given string data or file")

	encryptCmd.String("", "public", &argparse.Options{
		Required: true,
		Validate: func(args []string) error {
			if len(args) <= 0 || len(args[0]) <= 0 {
				return fmt.Errorf("public key file is required")
			}
			fi, err := os.Stat(args[0])
			if err != nil {
				return err
			}
			if fi.IsDir() {
				return fmt.Errorf("public file must be a valid file")
			}
			return nil
		},
		Help:    "The public key file",
		Default: nil,
	})

	encryptCmd.String("f", "file", &argparse.Options{
		Required: false,
		Validate: func(args []string) error {
			if len(args) <= 0 || len(args[0]) <= 0 {
				return nil
			}
			fi, err := os.Stat(args[0])
			if err != nil {
				return err
			}
			if fi.IsDir() {
				return fmt.Errorf("file must be a file")
			}
			return nil
		},
		Help:    "The string to encrypt",
		Default: nil,
	})

	encryptCmd.String("s", "str", &argparse.Options{
		Required: false,
		Validate: nil,
		Help:     "The string to encrypt",
		Default:  nil,
	})
}
