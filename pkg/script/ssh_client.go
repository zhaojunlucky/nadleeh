package script

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/pkg/sftp"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type SSHConfig struct {
	Host           string
	Port           int
	Username       string
	Password       string
	PrivateKeyPath string
	Timeout        int // in seconds, 0 means no timeout
}

type NSSSHClient struct {
	client   *ssh.Client
	sessions []*NSSHSession
}

// SSHConfig holds SSH connection configuration

// Dial connects to the SSH server using the provided configuration
func (s *NSSSHClient) dial(config SSHConfig) error {
	// Build auth methods
	var authMethods []ssh.AuthMethod

	// Password auth
	if config.Password != "" {
		authMethods = append(authMethods, ssh.Password(config.Password))
	}

	// Private key auth from file
	if config.PrivateKeyPath != "" {
		key, err := os.ReadFile(config.PrivateKeyPath)
		if err != nil {
			return fmt.Errorf("failed to read private key file: %w", err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return fmt.Errorf("failed to parse private key from file: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if len(authMethods) == 0 {
		return errors.New("no authentication method provided")
	}

	// Build SSH client config
	sshConfig := &ssh.ClientConfig{
		User: config.Username,
		Auth: authMethods,
	}

	if sshConfig.HostKeyCallback == nil {
		sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	}

	// Connect
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return fmt.Errorf("failed to dial %s: %w", addr, err)
	}

	s.client = client
	log.Debugf("SSH connected to %s", addr)
	return nil
}

func (s *NSSSHClient) Close() {
	if s.client != nil {
		for _, session := range s.sessions {
			session.Close()
		}
		err := s.client.Close()
		if err != nil {
			log.Warnf("failed to close client: %v", err)
		}
		s.client = nil
		s.sessions = nil
	}
}

func (s *NSSSHClient) WriteFile(content, remotePath string) error {

	// Create SFTP client
	sftpClient, err := sftp.NewClient(s.client)
	if err != nil {
		return fmt.Errorf("failed to create SFTP client: %w", err)
	}
	defer sftpClient.Close()

	// Create remote file
	remoteFile, err := sftpClient.Create(remotePath)
	if err != nil {
		return fmt.Errorf("failed to create remote file %s: %w", remotePath, err)
	}
	defer remoteFile.Close()

	// Write content
	_, err = remoteFile.Write([]byte(content))
	if err != nil {
		return fmt.Errorf("failed to write to remote file %s: %w", remotePath, err)
	}

	log.Debugf("wrote %d bytes to %s", len(content), remotePath)
	return nil
}

func (s *NSSSHClient) ReadFile(remotePath string) (string, error) {

	// Create SFTP client
	sftpClient, err := sftp.NewClient(s.client)
	if err != nil {
		return "", fmt.Errorf("failed to create SFTP client: %w", err)
	}
	defer sftpClient.Close()

	// Open remote file
	remoteFile, err := sftpClient.Open(remotePath)
	if err != nil {
		return "", fmt.Errorf("failed to open remote file %s: %w", remotePath, err)
	}
	defer remoteFile.Close()

	// Read content
	content, err := io.ReadAll(remoteFile)
	if err != nil {
		return "", fmt.Errorf("failed to read remote file %s: %w", remotePath, err)
	}

	log.Debugf("read %d bytes from %s", len(content), remotePath)
	return string(content), nil
}

func (s *NSSSHClient) UploadFile(localPath, remotePath string) error {

	// Open local file
	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file %s: %w", localPath, err)
	}
	defer localFile.Close()

	// Create SFTP client
	sftpClient, err := sftp.NewClient(s.client)
	if err != nil {
		return fmt.Errorf("failed to create SFTP client: %w", err)
	}
	defer sftpClient.Close()

	// Create remote file
	remoteFile, err := sftpClient.Create(remotePath)
	if err != nil {
		return fmt.Errorf("failed to create remote file %s: %w", remotePath, err)
	}
	defer remoteFile.Close()

	// Copy local file to remote
	bytes, err := io.Copy(remoteFile, localFile)
	if err != nil {
		return fmt.Errorf("failed to upload file to %s: %w", remotePath, err)
	}

	log.Debugf("uploaded %d bytes from %s to %s", bytes, localPath, remotePath)
	return nil
}

func (s *NSSSHClient) DownloadFile(remotePath, localPath string) error {

	// Create SFTP client
	sftpClient, err := sftp.NewClient(s.client)
	if err != nil {
		return fmt.Errorf("failed to create SFTP client: %w", err)
	}
	defer sftpClient.Close()

	// Open remote file
	remoteFile, err := sftpClient.Open(remotePath)
	if err != nil {
		return fmt.Errorf("failed to open remote file %s: %w", remotePath, err)
	}
	defer remoteFile.Close()

	// Create local file
	localFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file %s: %w", localPath, err)
	}
	defer localFile.Close()

	// Copy remote file to local
	bytes, err := io.Copy(localFile, remoteFile)
	if err != nil {
		return fmt.Errorf("failed to download file from %s: %w", remotePath, err)
	}

	log.Debugf("downloaded %d bytes from %s to %s", bytes, remotePath, localPath)
	return nil
}

func (s *NSSSHClient) NewSession() (*NSSHSession, error) {
	if s.client == nil {
		return nil, errors.New("client is not initialized")
	}

	sshSession, err := s.client.NewSession()
	if err != nil {
		return nil, err
	}
	session := &NSSHSession{session: sshSession}
	s.sessions = append(s.sessions, session)
	return session, nil
}

func NewNSSSHClient(config SSHConfig) (*NSSSHClient, error) {
	client := &NSSSHClient{}
	if err := client.dial(config); err != nil {
		return nil, err
	}

	return client, nil
}
