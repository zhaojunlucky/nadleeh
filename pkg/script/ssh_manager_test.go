package script

import (
	"testing"
)

func TestNSSSHManager(t *testing.T) {
	sshMgr := &NSSSHManager{}
	defer sshMgr.Close()
	sshClient, err := sshMgr.Dial("10.53.1.28", 22, "jun", map[string]any{
		"password": "jun",
	})

	if err != nil {
		t.Fatal(err)
	}

	defer sshClient.Close()
	session, err := sshClient.NewSession()
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	ret, err := session.RunCmd("ls", nil, map[string]any{
		"workingDir": "/tmp",
	})
	if err != nil {
		t.Fatal(err)
	}
	if ret.Status != 0 {
		t.Errorf("expected status 0, got %d", ret.Status)
	}
	t.Logf("stdout: %s", ret.Stdout)
	t.Logf("stderr: %s", ret.Stderr)

	str, err := sshClient.ReadFile("/home/jun/ssh.txt")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("file content: %s", str)
}
