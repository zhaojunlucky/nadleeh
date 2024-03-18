package googledrive

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	drive "google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"log"
	"nadleeh/pkg/env"
	"net/http"
	"os"
)

const (
	SCOPE = drive.DriveScope
)

type GoogleDrive struct {
	name       string
	path       string
	remotePath string
	config     map[string]string
	cred       string
}

func (g *GoogleDrive) Init(config map[string]string) error {
	g.config = config

	return nil
}

func (g *GoogleDrive) Run(parent env.Env) error {
	fmt.Println("Run Google Drive plugin")
	err := g.validate(parent)
	if err != nil {
		return err
	}
	client := ServiceAccount(g.cred)
	srv, err := drive.NewService(context.Background(), option.WithHTTPClient(client), option.WithScopes(SCOPE))
	if err != nil {
		return err
	}
	file, err := os.Open(g.path)
	if err != nil {
		return err
	}
	defer file.Close()
	f := &drive.File{Name: g.name, Parents: []string{g.remotePath}}
	res, err := srv.Files.
		Create(f).
		Media(file).
		ProgressUpdater(func(now, size int64) { fmt.Printf("%d, %d\r", now, size) }).
		Do()
	if err != nil {
		return err
	}
	fmt.Printf("https://drive.google.com/file/d/%s/view?usp=drive_link\n", res.Id)
	return nil
}

// ServiceAccount : Use Service account
func ServiceAccount(credentialFile string) *http.Client {
	b, err := os.ReadFile(credentialFile)
	if err != nil {
		log.Fatal(err)
	}
	var c = struct {
		Email      string `json:"client_email"`
		PrivateKey string `json:"private_key"`
	}{}
	json.Unmarshal(b, &c)
	config := &jwt.Config{
		Email:      c.Email,
		PrivateKey: []byte(c.PrivateKey),
		Scopes: []string{
			drive.DriveScope,
		},
		TokenURL: google.JWTTokenURL,
	}
	client := config.Client(context.TODO())
	return client
}

func (g *GoogleDrive) validate(parent env.Env) error {
	g.name = parent.Expand(g.config["name"])
	if len(g.name) <= 0 {
		return fmt.Errorf("invalid name")
	}
	g.path = parent.Expand(g.config["path"])
	if len(g.path) <= 0 {
		return fmt.Errorf("invalid path")
	}
	fi, err := os.Stat(g.path)
	if err != nil {
		return err
	}

	if fi.IsDir() {
		return fmt.Errorf("invalid path, it's not a file")
	}
	g.remotePath = parent.Expand(g.config["remote-path"])

	g.cred = parent.Expand(g.config["cred"])
	if len(g.cred) <= 0 {
		return fmt.Errorf("invalid cred")
	}
	return nil
}
