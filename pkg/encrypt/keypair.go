package encrypt

import (
	"fmt"
	"nadleeh/internal/argument"

	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/golib/pkg/security"

	"os"
	"path"
)

func GenerateKeyPair(args *argument.KeypairArgs) {
	pri, err := security.GenerateECKeyPair("secp256r1")
	if err != nil {
		log.Fatal(err)
	}

	priFile := path.Join(args.Dir, fmt.Sprintf("%s-private.pem", args.Name))
	pubFile := path.Join(args.Dir, fmt.Sprintf("%s-public.pem", args.Name))

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
