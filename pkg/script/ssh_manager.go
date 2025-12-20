package script

import "errors"

type NSSSHManager struct {
	clients []*NSSSHClient
}

func (s *NSSSHManager) Dial(host string, port int, username string, options map[string]any) (*NSSSHClient, error) {
	if options == nil {
		return nil, errors.New("options cannot be nil")
	}
	config := SSHConfig{
		Host:     host,
		Port:     port,
		Username: username,
	}
	if val, ok := options["password"].(string); ok {
		config.Password = val
	}
	if val, ok := options["privateKeyPath"].(string); ok {
		config.PrivateKeyPath = val
	}
	if val, ok := options["timeout"].(int); ok {
		config.Timeout = val
	}

	if config.Password == "" && config.PrivateKeyPath == "" {
		return nil, errors.New("password or privateKeyPath must be provided")
	}

	client, err := NewNSSSHClient(config)
	if err != nil {
		return nil, err
	}
	
	// Track client for cleanup
	s.clients = append(s.clients, client)
	return client, nil
}

func (s *NSSSHManager) Close() {
	for _, client := range s.clients {
		client.Close()
	}
}
