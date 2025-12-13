package encrypt

import (
	"encoding/base64"
	"fmt"
	"io"
	"nadleeh/internal/argument"
	"os"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/golib/pkg/security"
)

func Encrypt(args *argument.EncryptArgs) {
	reader, err := os.Open(args.Public)
	if err != nil {
		log.Fatal(err)
	}
	pubKey, err := security.ReadPublicKey(reader)
	if err != nil {
		log.Fatal(err)
	}
	ecies := security.ECIESHelper{}

	if args.File != "" {
		log.Infof("encrypt file: %s", args.File)
		file, err := os.Open(args.File)
		if err != nil {
			log.Fatal(err)
		}

		data, err := io.ReadAll(file)
		if err != nil {
			log.Fatal(err)
		}

		encrypted, err := ecies.EncryptWithPublic(pubKey, data)
		if err != nil {
			log.Fatal(err)
		}

		outputFilePath := path.Join(path.Dir(args.File), fmt.Sprintf("%s-encrypted%s", path.Base(args.File),
			path.Ext(args.File)))
		log.Infof("write encrypted file: %s", outputFilePath)
		err = os.WriteFile(outputFilePath, []byte(fmt.Sprintf("ENC(%s)", string(encrypted))), 0644)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	if args.Str != "" {
		str := strings.TrimSpace(args.Str)
		log.Infof("encrypt string: %s", str)
		encrypted, err := ecies.EncryptWithPublic(pubKey, []byte(str))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("entryped string is: ENC(%s)\n", base64.StdEncoding.EncodeToString(encrypted))
		return
	}
	log.Fatal("invalid argument for decrypt")
}
