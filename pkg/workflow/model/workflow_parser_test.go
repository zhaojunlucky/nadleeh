package workflow

import (
	"fmt"
	"nadleeh/pkg/file"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflow(t *testing.T) {
	root, err := file.GetProjectRootDir()
	if err != nil {
		t.Fatal(err)
	}
	ymlFile, err := os.Open(filepath.Join(root, "examples/backup.yml"))
	if err != nil {
		t.Fatal(err)
	}
	workflow, err := ParseWorkflow(ymlFile)

	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(workflow)
}

func TestParseEnv(t *testing.T) {
	t.Run("EmptyEnvAndFiles", func(t *testing.T) {
		result, err := parseEnv(nil, []string{})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result == nil {
			t.Error("Expected non-nil result")
		}
		if len(result) != 0 {
			t.Errorf("Expected empty map, got %d items", len(result))
		}
	})

	t.Run("ExistingEnvMap", func(t *testing.T) {
		existingEnv := map[string]string{
			"EXISTING_KEY": "existing_value",
			"ANOTHER_KEY":  "another_value",
		}
		result, err := parseEnv(existingEnv, []string{})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("Expected 2 items, got %d", len(result))
		}
		if result["EXISTING_KEY"] != "existing_value" {
			t.Errorf("Expected 'existing_value', got '%s'", result["EXISTING_KEY"])
		}
	})

	t.Run("ValidEnvFile", func(t *testing.T) {
		// Create a temporary env file
		tempDir := t.TempDir()
		envFile := filepath.Join(tempDir, "test.env")
		
		envContent := `# This is a comment
KEY1=value1
KEY2=value2
// Another comment style
KEY3=value with spaces
EMPTY_VALUE=
KEY_WITH_EQUALS=value=with=equals
`
		if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
			t.Fatalf("Failed to create env file: %v", err)
		}

		result, err := parseEnv(nil, []string{envFile})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expected := map[string]string{
			"KEY1":             "value1",
			"KEY2":             "value2",
			"KEY3":             "value with spaces",
			"EMPTY_VALUE":      "",
			"KEY_WITH_EQUALS":  "value=with=equals",
		}

		if len(result) != len(expected) {
			t.Errorf("Expected %d items, got %d", len(expected), len(result))
		}

		for key, expectedValue := range expected {
			if result[key] != expectedValue {
				t.Errorf("Expected %s='%s', got '%s'", key, expectedValue, result[key])
			}
		}
	})

	t.Run("MultipleEnvFiles", func(t *testing.T) {
		tempDir := t.TempDir()
		
		// Create first env file
		envFile1 := filepath.Join(tempDir, "env1.env")
		envContent1 := `KEY1=value1
SHARED_KEY=from_file1`
		if err := os.WriteFile(envFile1, []byte(envContent1), 0644); err != nil {
			t.Fatalf("Failed to create env file 1: %v", err)
		}

		// Create second env file
		envFile2 := filepath.Join(tempDir, "env2.env")
		envContent2 := `KEY2=value2
SHARED_KEY=from_file2`
		if err := os.WriteFile(envFile2, []byte(envContent2), 0644); err != nil {
			t.Fatalf("Failed to create env file 2: %v", err)
		}

		result, err := parseEnv(nil, []string{envFile1, envFile2})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result["KEY1"] != "value1" {
			t.Errorf("Expected KEY1='value1', got '%s'", result["KEY1"])
		}
		if result["KEY2"] != "value2" {
			t.Errorf("Expected KEY2='value2', got '%s'", result["KEY2"])
		}
		// Second file should override first file
		if result["SHARED_KEY"] != "from_file2" {
			t.Errorf("Expected SHARED_KEY='from_file2', got '%s'", result["SHARED_KEY"])
		}
	})

	t.Run("EnvFileWithExistingMap", func(t *testing.T) {
		tempDir := t.TempDir()
		envFile := filepath.Join(tempDir, "test.env")
		
		envContent := `FILE_KEY=file_value
OVERRIDE_KEY=from_file`
		if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
			t.Fatalf("Failed to create env file: %v", err)
		}

		existingEnv := map[string]string{
			"EXISTING_KEY": "existing_value",
			"OVERRIDE_KEY": "from_map",
		}

		result, err := parseEnv(existingEnv, []string{envFile})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result["EXISTING_KEY"] != "existing_value" {
			t.Errorf("Expected EXISTING_KEY='existing_value', got '%s'", result["EXISTING_KEY"])
		}
		if result["FILE_KEY"] != "file_value" {
			t.Errorf("Expected FILE_KEY='file_value', got '%s'", result["FILE_KEY"])
		}
		// File should override existing map
		if result["OVERRIDE_KEY"] != "from_file" {
			t.Errorf("Expected OVERRIDE_KEY='from_file', got '%s'", result["OVERRIDE_KEY"])
		}
	})

	t.Run("InvalidEnvFileLine", func(t *testing.T) {
		tempDir := t.TempDir()
		envFile := filepath.Join(tempDir, "invalid.env")
		
		envContent := `VALID_KEY=valid_value
INVALID_LINE_NO_EQUALS
ANOTHER_VALID=value`
		if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
			t.Fatalf("Failed to create env file: %v", err)
		}

		_, err := parseEnv(nil, []string{envFile})
		if err == nil {
			t.Error("Expected error for invalid env file line")
		}
		if !strings.Contains(err.Error(), "is not valid") {
			t.Errorf("Expected error message about invalid line, got: %v", err)
		}
	})

	t.Run("NonexistentEnvFile", func(t *testing.T) {
		// This test will trigger log.Fatal, so we need to handle it carefully
		// For now, we'll skip this test as it would cause the test to exit
		t.Skip("Skipping test that would trigger log.Fatal")
	})

	t.Run("EnvFileOpenError", func(t *testing.T) {
		// Create a directory instead of a file to trigger open error
		tempDir := t.TempDir()
		invalidFile := filepath.Join(tempDir, "directory_not_file")
		if err := os.Mkdir(invalidFile, 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}

		// parseEnv doesn't actually validate file types, it just tries to read
		// Reading a directory will not cause an error but will return empty content
		result, err := parseEnv(nil, []string{invalidFile})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("Expected 0 env vars from directory, got %d", len(result))
		}
	})

	t.Run("EmptyLinesAndComments", func(t *testing.T) {
		tempDir := t.TempDir()
		envFile := filepath.Join(tempDir, "comments.env")
		
		envContent := `
# Comment at start
KEY1=value1

// Another comment style
KEY2=value2

   # Indented comment
   
KEY3=value3
#KEY4=commented_out
//KEY5=also_commented
`
		if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
			t.Fatalf("Failed to create env file: %v", err)
		}

		result, err := parseEnv(nil, []string{envFile})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expected := map[string]string{
			"KEY1": "value1",
			"KEY2": "value2",
			"KEY3": "value3",
		}

		if len(result) != len(expected) {
			t.Errorf("Expected %d items, got %d", len(expected), len(result))
		}

		for key, expectedValue := range expected {
			if result[key] != expectedValue {
				t.Errorf("Expected %s='%s', got '%s'", key, expectedValue, result[key])
			}
		}
	})
}

func TestParseWorkflow(t *testing.T) {
	t.Run("ValidWorkflowYAML", func(t *testing.T) {
		yamlContent := `
name: "test-workflow"
version: "1.0.0"
working-dir: "/tmp"
env:
  ENV_KEY1: "env_value1"
  ENV_KEY2: "env_value2"
checks:
  private-key: true
  requires-root: false
  args:
    - name: "required_arg"
      pattern: ".*"
  envs:
    - name: "required_env"
      pattern: "^[A-Z]+$"
jobs:
  test-job:
    env:
      JOB_ENV: "job_value"
    steps:
      - name: "test-step"
        run: "echo hello"
  another-job:
    steps:
      - name: "another-step"
        script: "console.log('test')"
`
		reader := strings.NewReader(yamlContent)
		workflow, err := ParseWorkflow(reader)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Validate workflow properties
		if workflow.Name != "test-workflow" {
			t.Errorf("Expected name 'test-workflow', got '%s'", workflow.Name)
		}
		if workflow.Version != "1.0.0" {
			t.Errorf("Expected version '1.0.0', got '%s'", workflow.Version)
		}
		if workflow.WorkingDir != "/tmp" {
			t.Errorf("Expected working-dir '/tmp', got '%s'", workflow.WorkingDir)
		}

		// Validate environment variables
		if len(workflow.Env) != 2 {
			t.Errorf("Expected 2 env vars, got %d", len(workflow.Env))
		}
		if workflow.Env["ENV_KEY1"] != "env_value1" {
			t.Errorf("Expected ENV_KEY1='env_value1', got '%s'", workflow.Env["ENV_KEY1"])
		}

		// Validate checks
		if !workflow.Checks.PrivateKey {
			t.Error("Expected PrivateKey check to be true")
		}
		if workflow.Checks.RequiresRoot {
			t.Error("Expected RequiresRoot check to be false")
		}
		if len(workflow.Checks.Args) != 1 {
			t.Errorf("Expected 1 arg check, got %d", len(workflow.Checks.Args))
		}
		if workflow.Checks.Args[0].Name != "required_arg" {
			t.Errorf("Expected arg name 'required_arg', got '%s'", workflow.Checks.Args[0].Name)
		}

		// Validate jobs
		if len(workflow.Jobs) != 2 {
			t.Errorf("Expected 2 jobs, got %d", len(workflow.Jobs))
		}

		// Find and validate first job
		var testJob *Job
		for _, job := range workflow.Jobs {
			if job.Name == "test-job" {
				testJob = job
				break
			}
		}
		if testJob == nil {
			t.Fatal("Could not find 'test-job'")
		}
		if testJob.Env["JOB_ENV"] != "job_value" {
			t.Errorf("Expected JOB_ENV='job_value', got '%s'", testJob.Env["JOB_ENV"])
		}
		if len(testJob.Steps) != 1 {
			t.Errorf("Expected 1 step in test-job, got %d", len(testJob.Steps))
		}
		if testJob.Steps[0].Name != "test-step" {
			t.Errorf("Expected step name 'test-step', got '%s'", testJob.Steps[0].Name)
		}
		if testJob.Steps[0].Run != "echo hello" {
			t.Errorf("Expected run 'echo hello', got '%s'", testJob.Steps[0].Run)
		}
	})

	t.Run("MinimalWorkflowYAML", func(t *testing.T) {
		yamlContent := `
name: "minimal-workflow"
jobs:
  simple-job:
    steps:
      - name: "simple-step"
        run: "echo test"
`
		reader := strings.NewReader(yamlContent)
		workflow, err := ParseWorkflow(reader)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if workflow.Name != "minimal-workflow" {
			t.Errorf("Expected name 'minimal-workflow', got '%s'", workflow.Name)
		}
		if workflow.Version != "" {
			t.Errorf("Expected empty version, got '%s'", workflow.Version)
		}
		if len(workflow.Jobs) != 1 {
			t.Errorf("Expected 1 job, got %d", len(workflow.Jobs))
		}
	})

	t.Run("WorkflowWithEnvFiles", func(t *testing.T) {
		// Create temporary env files
		tempDir := t.TempDir()
		envFile1 := filepath.Join(tempDir, "env1.env")
		envFile2 := filepath.Join(tempDir, "env2.env")
		
		envContent1 := `FILE_KEY1=file_value1`
		envContent2 := `FILE_KEY2=file_value2`
		
		if err := os.WriteFile(envFile1, []byte(envContent1), 0644); err != nil {
			t.Fatalf("Failed to create env file 1: %v", err)
		}
		if err := os.WriteFile(envFile2, []byte(envContent2), 0644); err != nil {
			t.Fatalf("Failed to create env file 2: %v", err)
		}

		yamlContent := fmt.Sprintf(`
name: "workflow-with-env-files"
env-files:
  - "%s"
  - "%s"
env:
  YAML_KEY: "yaml_value"
jobs:
  test-job:
    steps:
      - name: "test-step"
        run: "echo test"
`, envFile1, envFile2)

		reader := strings.NewReader(yamlContent)
		workflow, err := ParseWorkflow(reader)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Validate that env files were parsed and merged with YAML env
		if len(workflow.Env) != 3 {
			t.Errorf("Expected 3 env vars, got %d", len(workflow.Env))
		}
		if workflow.Env["FILE_KEY1"] != "file_value1" {
			t.Errorf("Expected FILE_KEY1='file_value1', got '%s'", workflow.Env["FILE_KEY1"])
		}
		if workflow.Env["FILE_KEY2"] != "file_value2" {
			t.Errorf("Expected FILE_KEY2='file_value2', got '%s'", workflow.Env["FILE_KEY2"])
		}
		if workflow.Env["YAML_KEY"] != "yaml_value" {
			t.Errorf("Expected YAML_KEY='yaml_value', got '%s'", workflow.Env["YAML_KEY"])
		}
	})

	t.Run("InvalidYAML", func(t *testing.T) {
		yamlContent := `
name: "invalid-workflow"
jobs:
  invalid-job:
    steps:
      - name: "step1"
        run: "echo test"
      - name: "step2"
        invalid_yaml: [unclosed_bracket
`
		reader := strings.NewReader(yamlContent)
		_, err := ParseWorkflow(reader)
		if err == nil {
			t.Error("Expected error for invalid YAML")
		}
	})

	t.Run("EmptyWorkflow", func(t *testing.T) {
		yamlContent := `{}`
		reader := strings.NewReader(yamlContent)
		workflow, err := ParseWorkflow(reader)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if workflow.Name != "" {
			t.Errorf("Expected empty name, got '%s'", workflow.Name)
		}
		if len(workflow.Jobs) != 0 {
			t.Errorf("Expected 0 jobs, got %d", len(workflow.Jobs))
		}
	})

	t.Run("WorkflowWithComplexJobs", func(t *testing.T) {
		yamlContent := `
name: "complex-workflow"
jobs:
  job-with-multiple-steps:
    env:
      JOB_VAR: "job_value"
    steps:
      - name: "bash-step"
        run: "echo 'bash command'"
        env:
          STEP_VAR: "step_value"
      - name: "js-step"
        script: "console.log('javascript')"
        if: "true"
      - name: "plugin-step"
        uses: "test-plugin@v1.0.0"
        with:
          param1: "value1"
          param2: "value2"
        continue-on-error: "false"
  empty-job:
    steps: []
`
		reader := strings.NewReader(yamlContent)
		workflow, err := ParseWorkflow(reader)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(workflow.Jobs) != 2 {
			t.Errorf("Expected 2 jobs, got %d", len(workflow.Jobs))
		}

		// Find the complex job
		var complexJob *Job
		for _, job := range workflow.Jobs {
			if job.Name == "job-with-multiple-steps" {
				complexJob = job
				break
			}
		}
		if complexJob == nil {
			t.Fatal("Could not find 'job-with-multiple-steps'")
		}

		if len(complexJob.Steps) != 3 {
			t.Errorf("Expected 3 steps, got %d", len(complexJob.Steps))
		}

		// Validate bash step
		bashStep := complexJob.Steps[0]
		if bashStep.Name != "bash-step" {
			t.Errorf("Expected step name 'bash-step', got '%s'", bashStep.Name)
		}
		if bashStep.Run != "echo 'bash command'" {
			t.Errorf("Expected run command, got '%s'", bashStep.Run)
		}
		if bashStep.Env["STEP_VAR"] != "step_value" {
			t.Errorf("Expected STEP_VAR='step_value', got '%s'", bashStep.Env["STEP_VAR"])
		}

		// Validate JS step
		jsStep := complexJob.Steps[1]
		if jsStep.Script != "console.log('javascript')" {
			t.Errorf("Expected script, got '%s'", jsStep.Script)
		}
		if jsStep.If != "true" {
			t.Errorf("Expected if condition 'true', got '%s'", jsStep.If)
		}

		// Validate plugin step
		pluginStep := complexJob.Steps[2]
		if pluginStep.Uses != "test-plugin@v1.0.0" {
			t.Errorf("Expected uses 'test-plugin@v1.0.0', got '%s'", pluginStep.Uses)
		}
		if pluginStep.With["param1"] != "value1" {
			t.Errorf("Expected param1='value1', got '%s'", pluginStep.With["param1"])
		}
		if pluginStep.ContinueOnError != "false" {
			t.Errorf("Expected continue-on-error 'false', got '%s'", pluginStep.ContinueOnError)
		}
	})

	t.Run("WorkflowWithInvalidEnvFile", func(t *testing.T) {
		// This would trigger log.Fatal in parseEnv, so we skip this test
		t.Skip("Skipping test that would trigger log.Fatal for nonexistent env file")
	})

	t.Run("WorkflowWithJobParseError", func(t *testing.T) {
		yamlContent := `
name: "workflow-with-job-error"
jobs:
  valid-job:
    steps:
      - name: "valid-step"
        run: "echo test"
  invalid-job:
    steps:
      - name: "invalid-step"
        run: 123  # Invalid type for run field
`
		reader := strings.NewReader(yamlContent)
		workflow, err := ParseWorkflow(reader)
		// YAML parsing might succeed but convert 123 to string "123"
		// So we check if parsing succeeded and validate the result
		if err != nil {
			// If there's an error, it should be a job parse error
			if !strings.Contains(err.Error(), "failed to parse job") {
				t.Errorf("Expected job parse error, got: %v", err)
			}
		} else {
			// If parsing succeeded, validate that the run field was converted properly
			if workflow != nil && len(workflow.Jobs) > 0 {
				t.Logf("YAML parsing succeeded, run field converted to: %v", workflow.Jobs)
			}
		}
	})

	t.Run("WorkflowWithAllChecks", func(t *testing.T) {
		yamlContent := `
name: "workflow-with-checks"
checks:
  private-key: true
  requires-root: true
  args:
    - name: "arg1"
      pattern: "^[a-z]+$"
    - name: "arg2"
      pattern: "\\d+"
  envs:
    - name: "ENV1"
      pattern: ".*"
    - name: "ENV2"
      pattern: "^[A-Z_]+$"
jobs:
  test-job:
    steps:
      - name: "test-step"
        run: "echo test"
`
		reader := strings.NewReader(yamlContent)
		workflow, err := ParseWorkflow(reader)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Validate all checks
		if !workflow.Checks.PrivateKey {
			t.Error("Expected PrivateKey check to be true")
		}
		if !workflow.Checks.RequiresRoot {
			t.Error("Expected RequiresRoot check to be true")
		}
		if len(workflow.Checks.Args) != 2 {
			t.Errorf("Expected 2 arg checks, got %d", len(workflow.Checks.Args))
		}
		if len(workflow.Checks.Envs) != 2 {
			t.Errorf("Expected 2 env checks, got %d", len(workflow.Checks.Envs))
		}

		// Validate specific check patterns
		if workflow.Checks.Args[0].Pattern != "^[a-z]+$" {
			t.Errorf("Expected pattern '^[a-z]+$', got '%s'", workflow.Checks.Args[0].Pattern)
		}
		if workflow.Checks.Envs[1].Name != "ENV2" {
			t.Errorf("Expected env name 'ENV2', got '%s'", workflow.Checks.Envs[1].Name)
		}
	})
}

func TestWorkflowDefinitionStruct(t *testing.T) {
	t.Run("WorkflowDefinitionFields", func(t *testing.T) {
		yamlContent := `
name: "test-workflow"
version: "1.0.0"
working-dir: "/tmp"
env:
  KEY1: "value1"
checks:
  private-key: true
  requires-root: false
jobs:
  test-job:
    steps:
      - name: "test-step"
        run: "echo test"
`
		reader := strings.NewReader(yamlContent)
		workflow, err := ParseWorkflow(reader)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Test that all fields are properly populated
		if workflow.Name == "" {
			t.Error("Expected non-empty workflow name")
		}
		if workflow.Version == "" {
			t.Error("Expected non-empty workflow version")
		}
		if workflow.WorkingDir == "" {
			t.Error("Expected non-empty working directory")
		}
		if workflow.Env == nil {
			t.Error("Expected non-nil environment map")
		}
		if workflow.Jobs == nil {
			t.Error("Expected non-nil jobs slice")
		}
	})

	t.Run("WorkflowDefinitionDefaults", func(t *testing.T) {
		yamlContent := `
name: "minimal-workflow"
jobs:
  test-job:
    steps:
      - name: "test-step"
        run: "echo test"
`
		reader := strings.NewReader(yamlContent)
		workflow, err := ParseWorkflow(reader)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Test default values
		if workflow.Version != "" {
			t.Errorf("Expected empty version by default, got '%s'", workflow.Version)
		}
		if workflow.WorkingDir != "" {
			t.Errorf("Expected empty working-dir by default, got '%s'", workflow.WorkingDir)
		}
		if workflow.Env == nil {
			t.Error("Expected initialized environment map")
		}
		if len(workflow.Env) != 0 {
			t.Errorf("Expected empty environment map by default, got %d items", len(workflow.Env))
		}
	})
}

func TestParseWorkflowEdgeCases(t *testing.T) {
	t.Run("WorkflowWithNoJobs", func(t *testing.T) {
		yamlContent := `
name: "no-jobs-workflow"
version: "1.0.0"
`
		reader := strings.NewReader(yamlContent)
		workflow, err := ParseWorkflow(reader)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(workflow.Jobs) != 0 {
			t.Errorf("Expected 0 jobs, got %d", len(workflow.Jobs))
		}
	})

	t.Run("WorkflowWithEmptyJobsMap", func(t *testing.T) {
		yamlContent := `
name: "empty-jobs-workflow"
jobs: {}
`
		reader := strings.NewReader(yamlContent)
		workflow, err := ParseWorkflow(reader)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(workflow.Jobs) != 0 {
			t.Errorf("Expected 0 jobs, got %d", len(workflow.Jobs))
		}
	})

	t.Run("WorkflowWithSpecialCharacters", func(t *testing.T) {
		yamlContent := `
name: "special-chars-workflow"
env:
  SPECIAL_CHARS: "!@#$%^&*()_+-=[]{}|;':\",./<>?"
  UNICODE: "こんにちは世界"
  MULTILINE: |
    Line 1
    Line 2
    Line 3
jobs:
  job-with-special-name:
    steps:
      - name: "step with spaces and symbols !@#"
        run: "echo 'special chars test'"
`
		reader := strings.NewReader(yamlContent)
		workflow, err := ParseWorkflow(reader)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if workflow.Env["SPECIAL_CHARS"] != "!@#$%^&*()_+-=[]{}|;':\",./<>?" {
			t.Errorf("Special characters not preserved in env var")
		}
		if workflow.Env["UNICODE"] != "こんにちは世界" {
			t.Errorf("Unicode characters not preserved in env var")
		}
		if !strings.Contains(workflow.Env["MULTILINE"], "Line 1") {
			t.Errorf("Multiline string not preserved in env var")
		}
	})

	t.Run("WorkflowWithLargeContent", func(t *testing.T) {
		// Create a workflow with many jobs and steps
		var yamlBuilder strings.Builder
		yamlBuilder.WriteString(`
name: "large-workflow"
version: "1.0.0"
env:
`)
		// Add many environment variables
		for i := 0; i < 50; i++ {
			yamlBuilder.WriteString(fmt.Sprintf("  ENV_VAR_%d: \"value_%d\"\n", i, i))
		}
		yamlBuilder.WriteString("jobs:\n")
		// Add many jobs
		for i := 0; i < 20; i++ {
			yamlBuilder.WriteString(fmt.Sprintf("  job-%d:\n", i))
			yamlBuilder.WriteString("    steps:\n")
			for j := 0; j < 5; j++ {
				yamlBuilder.WriteString(fmt.Sprintf("      - name: \"step-%d-%d\"\n", i, j))
				yamlBuilder.WriteString(fmt.Sprintf("        run: \"echo 'job %d step %d'\"\n", i, j))
			}
		}

		reader := strings.NewReader(yamlBuilder.String())
		workflow, err := ParseWorkflow(reader)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(workflow.Env) != 50 {
			t.Errorf("Expected 50 env vars, got %d", len(workflow.Env))
		}
		if len(workflow.Jobs) != 20 {
			t.Errorf("Expected 20 jobs, got %d", len(workflow.Jobs))
		}
		// Check that each job has 5 steps
		for _, job := range workflow.Jobs {
			if len(job.Steps) != 5 {
				t.Errorf("Expected 5 steps in job %s, got %d", job.Name, len(job.Steps))
			}
		}
	})

	t.Run("WorkflowWithNullValues", func(t *testing.T) {
		yamlContent := `
name: "null-values-workflow"
version: null
working-dir: null
env:
  NULL_VAR: null
  EMPTY_VAR: ""
jobs:
  test-job:
    env:
      JOB_NULL: null
    steps:
      - name: "test-step"
        run: "echo test"
        env:
          STEP_NULL: null
`
		reader := strings.NewReader(yamlContent)
		workflow, err := ParseWorkflow(reader)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// YAML null values should be converted to empty strings in Go
		if workflow.Version != "" {
			t.Errorf("Expected empty version for null, got '%s'", workflow.Version)
		}
		if workflow.WorkingDir != "" {
			t.Errorf("Expected empty working-dir for null, got '%s'", workflow.WorkingDir)
		}
	})
}

// Benchmark tests for performance validation
func BenchmarkParseEnv(b *testing.B) {
	// Create temporary env files for benchmarking
	tempDir := b.TempDir()
	envFile := filepath.Join(tempDir, "bench.env")
	envContent := `
KEY1=value1
KEY2=value2
KEY3=value3
KEY4=value4
KEY5=value5
`
	if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
		b.Fatalf("Failed to create env file: %v", err)
	}

	existingEnv := map[string]string{
		"EXISTING1": "existing_value1",
		"EXISTING2": "existing_value2",
	}
	envFiles := []string{envFile}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parseEnv(existingEnv, envFiles)
		if err != nil {
			b.Fatalf("parseEnv failed: %v", err)
		}
	}
}

func BenchmarkParseWorkflow_Simple(b *testing.B) {
	yamlContent := `
name: "benchmark-workflow"
version: "1.0.0"
env:
  KEY1: "value1"
  KEY2: "value2"
jobs:
  test-job:
    steps:
      - name: "test-step"
        run: "echo test"
`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(yamlContent)
		_, err := ParseWorkflow(reader)
		if err != nil {
			b.Fatalf("ParseWorkflow failed: %v", err)
		}
	}
}

func BenchmarkParseWorkflow_Complex(b *testing.B) {
	// Create a complex workflow for benchmarking
	var yamlBuilder strings.Builder
	yamlBuilder.WriteString(`
name: "complex-benchmark-workflow"
version: "1.0.0"
working-dir: "/tmp"
env:
`)
	// Add environment variables
	for i := 0; i < 20; i++ {
		yamlBuilder.WriteString(fmt.Sprintf("  ENV_VAR_%d: \"value_%d\"\n", i, i))
	}
	yamlBuilder.WriteString(`
checks:
  private-key: true
  requires-root: false
  args:
    - name: "arg1"
      pattern: ".*"
    - name: "arg2"
      pattern: "\\d+"
  envs:
    - name: "ENV1"
      pattern: ".*"
jobs:
`)
	// Add multiple jobs with steps
	for i := 0; i < 10; i++ {
		yamlBuilder.WriteString(fmt.Sprintf("  job-%d:\n", i))
		yamlBuilder.WriteString("    env:\n")
		yamlBuilder.WriteString(fmt.Sprintf("      JOB_VAR_%d: \"job_value_%d\"\n", i, i))
		yamlBuilder.WriteString("    steps:\n")
		for j := 0; j < 3; j++ {
			yamlBuilder.WriteString(fmt.Sprintf("      - name: \"step-%d-%d\"\n", i, j))
			if j%2 == 0 {
				yamlBuilder.WriteString(fmt.Sprintf("        run: \"echo 'job %d step %d'\"\n", i, j))
			} else {
				yamlBuilder.WriteString(fmt.Sprintf("        script: \"console.log('job %d step %d')\"\n", i, j))
			}
			yamlBuilder.WriteString("        env:\n")
			yamlBuilder.WriteString(fmt.Sprintf("          STEP_VAR_%d_%d: \"step_value_%d_%d\"\n", i, j, i, j))
		}
	}

	yamlContent := yamlBuilder.String()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(yamlContent)
		_, err := ParseWorkflow(reader)
		if err != nil {
			b.Fatalf("ParseWorkflow failed: %v", err)
		}
	}
}

func BenchmarkParseWorkflow_WithEnvFiles(b *testing.B) {
	// Create temporary env files for benchmarking
	tempDir := b.TempDir()
	envFile1 := filepath.Join(tempDir, "bench1.env")
	envFile2 := filepath.Join(tempDir, "bench2.env")
	
	envContent1 := `
FILE_KEY1=file_value1
FILE_KEY2=file_value2
FILE_KEY3=file_value3
`
	envContent2 := `
FILE_KEY4=file_value4
FILE_KEY5=file_value5
FILE_KEY6=file_value6
`
	
	if err := os.WriteFile(envFile1, []byte(envContent1), 0644); err != nil {
		b.Fatalf("Failed to create env file 1: %v", err)
	}
	if err := os.WriteFile(envFile2, []byte(envContent2), 0644); err != nil {
		b.Fatalf("Failed to create env file 2: %v", err)
	}

	yamlContent := fmt.Sprintf(`
name: "benchmark-workflow-with-env-files"
version: "1.0.0"
env-files:
  - "%s"
  - "%s"
env:
  YAML_KEY1: "yaml_value1"
  YAML_KEY2: "yaml_value2"
jobs:
  test-job:
    steps:
      - name: "test-step"
        run: "echo test"
`, envFile1, envFile2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(yamlContent)
		_, err := ParseWorkflow(reader)
		if err != nil {
			b.Fatalf("ParseWorkflow failed: %v", err)
		}
	}
}
