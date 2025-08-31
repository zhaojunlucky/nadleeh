package encrypt

import (
	"fmt"
	"nadleeh/internal/argument"

	"github.com/akamensky/argparse"
	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/golib/pkg/security"

	"os"
	"path"
)

func GenerateKeyPair(cmd *argparse.Command, argsMap map[string]argparse.Arg) {

	pName, err := argument.GetStringFromArg(argsMap["name"], true)
	if err != nil {
		log.Fatal(err)
	}

	pDir, err := argument.GetStringFromArg(argsMap["dir"], true)
	if err != nil {
		log.Fatal(err)
	}

	pri, err := security.GenerateECKeyPair("secp256r1")
	if err != nil {
		log.Fatal(err)
	}

	priFile := path.Join(*pDir, fmt.Sprintf("%s-private.pem", *pName))
	pubFile := path.Join(*pDir, fmt.Sprintf("%s-public.pem", *pName))

	log.Infof("Saving public key %s", pubFile)

	pubWriter, err := os.Create(pubFile)
	if err != nil {
		log.Fatal(err)
	}

	err = security.WritePublicKey(&pri.PublicKey, pubWriter)
	if err != nil {
		log.Fatal(err)
	}

	priWriter, err := os.Create(priFile)
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("Saving private key %s", priFile)

	err = security.WriteECPrivateKey(pri, priWriter)
	if err != nil {
		log.Fatal(err)
	}
}
