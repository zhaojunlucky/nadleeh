package encrypt

import (
	"encoding/base64"
	"fmt"
	"github.com/akamensky/argparse"
	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/golib/pkg/security"
	"io"
	"nadleeh/internal/argument"
	"os"
	"path"
	"strings"
)

func Encrypt(cmd *argparse.Command, argsMap map[string]argparse.Arg) {
	pPub, err := argument.GetStringFromArg(argsMap["public"], true)
	if err != nil {
		log.Fatal(err)
	}
	reader, err := os.Open(*pPub)
	if err != nil {
		log.Fatal(err)
	}
	pubKey, err := security.ReadPublicKey(reader)
	if err != nil {
		log.Fatal(err)
	}
	ecies := security.ECIESHelper{}

	pFileArg := argsMap["file"]
	if pFileArg.GetParsed() {
		filePath := pFileArg.GetResult().(string)
		log.Infof("encrypt file: %s", filePath)
		file, err := os.Open(filePath)
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

		outputFilePath := path.Join(path.Dir(filePath), fmt.Sprintf("%s-encrypted%s", path.Base(filePath),
			path.Ext(filePath)))
		log.Infof("write encrypted file: %s", outputFilePath)
		err = os.WriteFile(outputFilePath, []byte(fmt.Sprintf("ENC(%s)", string(encrypted))), 0644)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	pStr := argsMap["str"]
	if pStr.GetParsed() {
		pStr := pStr.GetResult().(*string)
		str := strings.TrimSpace(*pStr)
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
