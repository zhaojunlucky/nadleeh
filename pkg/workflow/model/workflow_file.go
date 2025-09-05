package workflow

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"nadleeh/pkg/file"
	"nadleeh/pkg/workflow/core"
	"net/http"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/go-github/v74/github"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var (
	bearer = "Bearer"
	basic  = "Basic"

	githubProvider = "github"
	httpsProvider  = "https"
	defaultExt     = ".yml"
)

type workflowCred struct {
	Type     string `yaml:"type"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}
type workflowProvider struct {
	Type   string       `yaml:"type"`
	Server string       `yaml:"server"`
	Owner  string       `yaml:"owner"`
	Name   string       `yaml:"name"`
	Cred   workflowCred `yaml:"cred"`
}

var defaultGitHubProvider = workflowProvider{
	Type:   githubProvider,
	Server: "https://github.com",
	Owner:  "nadleehz",
	Name:   "workflows",
	Cred: workflowCred{
		Type: "",
	},
}

func (w *workflowProvider) Download(name string) (io.Reader, error) {
	log.Infof("download workflow file %s provided by provider %s", name, w.Type)

	switch w.Type {
	case githubProvider:
		return w.downloadGitHub(name)
	case httpsProvider:
		return w.downloadHTTP(name)
	default:
		return nil, fmt.Errorf("unsupported provider type '%s'", w.Type)
	}
}

func (w *workflowProvider) downloadHTTP(name string) (io.Reader, error) {
	if len(w.Server) == 0 || !strings.HasPrefix(w.Server, "https") {
		log.Errorf("invalid server url for https provider")
		return nil, fmt.Errorf("for https provider, the server is required")
	}
	url := fmt.Sprintf("%s/%s", strings.TrimSuffix(w.Server, "/"), strings.TrimPrefix(name, "/"))
	return w.downloadURL(url)
}

func (w *workflowProvider) getOwnerRepoAndName(name string) (string, string, string, error) {
	if strings.HasPrefix(name, "@") {
		if len(w.Name) == 0 || len(w.Owner) == 0 {
			return "", "", "", fmt.Errorf("invalid workflow file name %s, you want to use default nadleehz/workflows repo, but owner or name is not specified", name)
		}
		return w.Owner, w.Name, name[1:], nil
	} else if len(w.Name) > 0 && len(w.Owner) > 0 {
		return w.Owner, w.Name, name, nil
	} else {
		segs := strings.Split(name, "/")
		if len(segs) < 3 {
			return "", "", "", fmt.Errorf("invalid github workflow file, it should bt <owner>/<repo>/<one or more paths>")
		}

		return segs[0], segs[1], strings.Join(segs[2:], "/"), nil
	}

}

func (w *workflowProvider) downloadGitHub(name string) (io.Reader, error) {

	owner, repo, wfPath, err := w.getOwnerRepoAndName(name)
	if err != nil {
		return nil, err
	}

	log.Infof("workflow file %s/%s/%s", owner, repo, wfPath)

	if w.Cred.Type != "" && w.Cred.Type != bearer {
		return nil, fmt.Errorf("only Bearer or empty cred type supported for github provider")
	}
	client := github.NewClient(nil)
	if w.Cred.Type != "" {
		client = client.WithAuthToken(w.Cred.Password)
	}

	fileContent, _, _, err := client.Repositories.GetContents(context.Background(), owner, repo, wfPath, nil)
	if err != nil {
		return nil, err
	}
	if fileContent == nil {
		return nil, fmt.Errorf("the workflow file %s is't a file wfPath", name)
	}
	content, err := fileContent.GetContent()
	if err != nil {
		log.Fatalf("error decoding content: %v", err)
		return nil, err
	}
	return strings.NewReader(content), nil
}

func (w *workflowProvider) downloadURL(url string) (io.Reader, error) {
	var urlRegex = regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	if !urlRegex.MatchString(url) {
		return nil, fmt.Errorf("invalid url %s", url)
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Errorf("failed to create http request: %v", err)
		return nil, err
	}

	if err = w.addHTTPHeader(req); err != nil {
		log.Errorf("failed to set authorization header: %v", err)
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorf("failed to execute http request: %v", err)
		return nil, err
	}
	log.Infof("response code: %d", resp.StatusCode)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("invalid response code %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("failed to read http response: %v", err)
		return nil, err
	}
	return bytes.NewReader(data), nil
}

func (w *workflowProvider) addHTTPHeader(req *http.Request) error {
	switch w.Cred.Type {
	case bearer:
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", w.Cred.Password))
	case basic:
		{
			auth := fmt.Sprintf("%s:%s", w.Cred.Username, w.Cred.Password)
			req.Header.Add("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(auth))))
		}
	default:

	}
	return nil
}

func LoadWorkflowFile(yml string, wa *core.WorkflowArgs) (io.Reader, error) {
	if !strings.HasSuffix(yml, defaultExt) && !strings.HasSuffix(yml, ".yaml") {
		log.Infof("add default yaml ext .yml to workflow name")
		yml = fmt.Sprintf("%s%s", yml, defaultExt)
	}

	if wa.Provider != nil {
		if len(*wa.Provider) == 0 {
			return nil, fmt.Errorf("provider is empty")
		}
		currentUser, err := user.Current()
		if err != nil {
			return nil, fmt.Errorf("failed to get current user: %w", err)
		}

		providerFile := filepath.Join(currentUser.HomeDir, ".nadleeh/providers/", *wa.Provider)
		log.Infof("provider file %s", providerFile)
		val, err := file.FileExists(providerFile)
		if err != nil {
			log.Errorf("failed to check provider file %s", *wa.Provider)
			return nil, err
		}
		var wp workflowProvider
		if !val && *wa.Provider == githubProvider {
			log.Infof("github provider file %s doesn't exist, use default", *wa.Provider)
			wp = defaultGitHubProvider
		} else {
			pFile, err := os.Open(providerFile)
			if err != nil {
				return nil, fmt.Errorf("failed to read provider %s: %v", *wa.Provider, err)
			}
			defer pFile.Close()

			if err := yaml.NewDecoder(pFile).Decode(&wp); err != nil {
				return nil, err
			}
		}

		return wp.Download(yml)

	}

	ext := strings.ToLower(path.Ext(yml))
	if ext != ".yaml" && ext != ".yml" {
		log.Fatalf("%s must be a yaml file", yml)
	}
	fi, err := os.Stat(yml)
	if err != nil {
		log.Fatalf("failed to get yaml file %v", err)
	}
	if fi.IsDir() {
		log.Fatalf("%s must be a file", yml)
	}

	return os.Open(yml)
}
