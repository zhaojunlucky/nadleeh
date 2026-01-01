package script

import (
	"log"
	"testing"
)

// Disabled
func TestDockerCli(t *testing.T) {
	host := "tcp://10.53.1.66:2375"
	cli := NewDockerCli(&host)

	resp, err := cli.InspectContainer("infisical-db")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v\n", resp)
}

func TestDockerCli2(t *testing.T) {
	host := "tcp://10.53.1.66:2375"
	cli := NewDockerCli(&host)

	resp, err := cli.ListContainers(map[string]any{"all": true})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v\n", resp)
}

func TestDockerCli3(t *testing.T) {
	host := "tcp://10.53.1.66:2375"
	cli := NewDockerCli(&host)

	cmds := []string{
		"tar", "-czf", `/backup/postgres.tar.gz`, "-C", "/data", ".",
	}
	output, err := cli.RunImage("busybox:latest", cmds, map[string]any{
		"user":          "root",
		"autoRemove":    true,
		"captureOutput": true,
		"volumes": []string{
			"infisical_pg_data:/data",
			"/tmp/postgres:/backup",
		},
	})
	log.Println(output)
	if err != nil {
		log.Fatal(err)
	}
}
