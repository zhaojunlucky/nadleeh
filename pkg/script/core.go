package script

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/santhosh-tekuri/jsonschema/v5"
	log "github.com/sirupsen/logrus"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"
)

type NJSCore struct {
}

type CmdResult struct {
	Status int
	Stdout string
	Stderr string
}

func (n *NJSCore) ParseYAML(data string) (map[string]any, error) {
	var ret map[string]any
	err := yaml.Unmarshal([]byte(data), &ret)
	if err != nil {
		log.Errorf("failed to parse yaml: %v", err)
	}
	return ret, err
}

func (n *NJSCore) ParseYAMLFile(filePath string) (map[string]any, error) {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var ret map[string]any
	err = yaml.Unmarshal(fileContent, &ret)
	if err != nil {
		log.Errorf("failed to parse yaml: %v", err)
	}
	return ret, err
}

func (n *NJSCore) ValidateJSONSchema(data any, jsonSchema string) error {
	schema, err := jsonschema.CompileString("myschema.json", jsonSchema)
	if err != nil {
		log.Errorf("schema compilation error: %v", err)
		return fmt.Errorf("schema compilation error: %v", err)
	}
	if err = schema.Validate(data); err != nil {
		log.Errorf("validation failed: %v", err)
		return fmt.Errorf("validation failed: %v", err)
	} else {
		return nil
	}
}

func (n *NJSCore) RunCmd(name string, args *[]string, options map[string]any) *CmdResult {
	log.Debugf("run cmd %s with args: %v", name, args)
	var cmd *exec.Cmd
	if args != nil {
		cmd = exec.Command(name, *args...)
	} else {
		cmd = exec.Command(name)
	}
	if options != nil {
		if workingDir, ok := options["workingDir"]; ok {
			cmd.Dir = workingDir.(string)
		}
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run() // Runs the command and waits for it to complete

	ret := &CmdResult{
		Status: 0,
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}

	if err != nil {
		log.Error(err)
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			ret.Status = exitError.ExitCode()
		} else {
			ret.Status = 255
		}
	}
	return ret
}

// ReadLine reads a line of text from stdin (console input)
func (n *NJSCore) ReadLine(prompt string) (string, error) {
	if prompt != "" {
		fmt.Print(prompt)
	}
	
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	
	// Remove trailing newline
	if len(line) > 0 && line[len(line)-1] == '\n' {
		line = line[:len(line)-1]
		// Also remove \r if present (Windows)
		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}
	}
	
	return line, nil
}

// ReadPassword reads a password from stdin without echoing (hidden input)
func (n *NJSCore) ReadPassword(prompt string) (string, error) {
	if prompt != "" {
		fmt.Print(prompt)
	}
	
	// Read password without echoing
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	
	// Print newline after password input
	fmt.Println()
	
	return string(password), nil
}

// ReadKey reads a single key press from stdin
// Note: This requires terminal to be in raw mode
func (n *NJSCore) ReadKey() (string, error) {
	// Save the current terminal state
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)
	
	// Read a single byte
	b := make([]byte, 1)
	_, err = os.Stdin.Read(b)
	if err != nil {
		return "", err
	}
	
	return string(b), nil
}
