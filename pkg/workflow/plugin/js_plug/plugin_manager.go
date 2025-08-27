package js_plug

import (
	"fmt"
	"nadleeh/pkg/file"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

var (
	LocalPath     = filepath.Join(os.Getenv("HOME"), ".nadleeh", "plugins")
	LocalDataPath = filepath.Join(LocalPath, "data")
	LocalLockPath = filepath.Join(LocalPath, ".locks")
	RemotePath    = "https://github.com"
	Main          = "main.js"
	Manifest      = "manifest.yml"

	Local  = 1
	Remote = 2
)

var PM = NewPluginManager()

type PluginMetadata struct {
	scheme       int
	Name         string
	Version      string
	Key          string
	LocalPath    string
	RemotePath   string
	token        string
	MainFile     string
	ManifestFile string
	lockFile     string
}

func (p *PluginMetadata) Load() error {
	if p.scheme == Local {
		fiInfo, err := os.Stat(p.LocalPath)
		if err != nil {
			return err
		}
		if !fiInfo.IsDir() {
			return fmt.Errorf("invalid path %s for plugin %s", p.LocalPath, p.Name)
		}
		fiInfo, err = os.Stat(p.MainFile)
		if err != nil {
			return err
		}
		if fiInfo.IsDir() {
			return fmt.Errorf("invalid main file %s for plugin %s", p.MainFile, p.Name)
		}
		return p.checkPlugin()
	} else {

		fileLock := file.NewFileLock(p.lockFile)
		if err := fileLock.Lock(); err != nil {
			return err
		}
		defer fileLock.Unlock()

		isDir, err := file.DirExists(p.LocalPath)
		if err != nil {
			return err
		}
		if isDir {
			cmd := exec.Command("git", "-C", p.LocalPath, "status", "--porcelain")
			output, err := cmd.CombinedOutput()
			if err != nil {
				return err
			}
			status := string(output)
			if len(strings.TrimSpace(status)) == 0 {
				return p.checkPlugin()
			}

		}

		if err = os.Remove(p.LocalPath); err != nil {
			return err
		}

		if err = os.MkdirAll(filepath.Dir(p.LocalPath), os.ModePerm); err != nil {
			return err
		}

		cmd := exec.Command("git", "-C", p.LocalPath, "clone", "-b", p.Version)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return err
		}
		log.Info(string(output))
		return p.checkPlugin()
	}

}

func (p *PluginMetadata) checkPlugin() error {
	exists, err := file.FileExists(p.MainFile)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("%s doesn't exist", p.MainFile)
	}

	exists, err = file.FileExists(p.ManifestFile)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("%s doesn't exist", p.ManifestFile)
	}
	return nil
}

func NewPluginMetadata(name, version, token, localPath string) (*PluginMetadata, error) {
	if len(version) == 0 {
		return nil, fmt.Errorf("version is not specified")
	}

	pm := &PluginMetadata{
		Name:    name,
		Version: version,
		token:   token,
		Key:     formatPluginKey(name, version),
	}
	if len(localPath) > 0 {
		pm.scheme = Local
		pm.LocalPath = localPath
		pm.RemotePath = localPath
	} else {
		pm.scheme = Remote
		pm.LocalPath = filepath.Join(LocalDataPath, name, version)
		pm.RemotePath = fmt.Sprintf("%s/%s", RemotePath, name)
		pm.lockFile = filepath.Join(LocalLockPath, fmt.Sprintf("%s-%s.lock", name, version))
	}
	pm.MainFile = filepath.Join(pm.LocalPath, Main)

	exist, err := file.FileExists(pm.MainFile)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, fmt.Errorf("plugin main.js %s doesn't exist", pm.MainFile)
	}
	pm.ManifestFile = filepath.Join(pm.LocalPath, Manifest)

	exist, err = file.FileExists(pm.ManifestFile)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, fmt.Errorf("plugin manifest.yml %s doesn't exist", pm.ManifestFile)
	}

	return pm, nil
}

func formatPluginKey(name, version string) string {
	return fmt.Sprintf("plugin:%s@%s", name, version)
}

type PluginManager struct {
	LoadedPlugin map[string]*PluginMetadata
}

func (p *PluginManager) LoadPlugin(name, version, token, localPath string) (*PluginMetadata, error) {
	key := formatPluginKey(name, version)
	pm := p.LoadedPlugin[key]
	if pm != nil {
		return pm, nil
	}
	pm, err := NewPluginMetadata(name, version, token, localPath)
	if err != nil {
		return nil, err
	}
	p.LoadedPlugin[key] = pm
	return pm, nil
}

func prepareDir(path string) {
	exists, err := file.DirExists(path)
	if err != nil {
		log.Fatal(err)
	}
	if !exists {
		if err = os.MkdirAll(path, os.ModePerm); err != nil {
			log.Fatal(err)
		}
	}

}

func NewPluginManager() *PluginManager {
	prepareDir(LocalDataPath)
	prepareDir(LocalLockPath)

	pm := &PluginManager{
		LoadedPlugin: make(map[string]*PluginMetadata),
	}
	return pm
}
