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
			log.Errorf("plugin local path %s doesn't exist", p.LocalPath)
			return err
		}
		if !fiInfo.IsDir() {
			log.Errorf("plugin local path %s is not a dir", p.LocalPath)
			return fmt.Errorf("invalid path %s for plugin %s", p.LocalPath, p.Name)
		}
		fiInfo, err = os.Stat(p.MainFile)
		if err != nil {
			log.Errorf("plugin main js %s doesn't exist", p.MainFile)
			return err
		}
		if fiInfo.IsDir() {
			log.Errorf("plugin main js %s is not a file", p.MainFile)
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
			if err == nil {
				status := string(output)
				if len(strings.TrimSpace(status)) == 0 {
					return p.checkPlugin()
				} else {
					log.Warnf("plugin local repo path %s is not clean will reclone", p.LocalPath)
				}
			}

		}

		if err = os.Remove(p.LocalPath); err != nil {
			if !os.IsNotExist(err) {
				log.Errorf("failed to remove existing plugin path %s: %v", p.LocalPath, err)
				return err
			}
		}

		if err = os.MkdirAll(p.LocalPath, os.ModePerm); err != nil {
			log.Errorf("failed to create plugin local path %s: %v", p.LocalPath, err)
			return err
		}

		cmd := exec.Command("git", "clone", "-b", p.Version, fmt.Sprintf("https://github.com/%s", p.Name), p.LocalPath)
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
		log.Errorf("plugin main js %s doesn't exist: %v", p.MainFile, err)
		return err
	}
	if !exists {
		log.Errorf("plugin main js %s doesn't exist", p.MainFile)
		return fmt.Errorf("%s doesn't exist", p.MainFile)
	}

	exists, err = file.FileExists(p.ManifestFile)
	if err != nil {
		log.Errorf("plugin manifest %s doesn't exist: %v", p.ManifestFile, err)
		return err
	}
	if !exists {
		log.Errorf("plugin mmanifest %s doesn't exist", p.ManifestFile)
		return fmt.Errorf("%s doesn't exist", p.ManifestFile)
	}
	return nil
}

func NewPluginMetadata(name, version, token, localPath string) (*PluginMetadata, error) {
	if len(version) == 0 {
		log.Errorf("version is missing for plugin %s", name)
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
	pm.ManifestFile = filepath.Join(pm.LocalPath, Manifest)

	if err := pm.Load(); err != nil {
		return nil, err
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
