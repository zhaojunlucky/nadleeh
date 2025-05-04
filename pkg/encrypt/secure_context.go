package encrypt

import (
	"crypto/ecdsa"
	"encoding/base64"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/golib/pkg/security"
	"os"
	"regexp"
	"strings"
)

type SecureContext struct {
	privateKey  *ecdsa.PrivateKey
	eciesHelper security.ECIESHelper
	pattern     *regexp.Regexp
}

func NewSecureContext(priFile *string) SecureContext {
	var priKey *ecdsa.PrivateKey
	if priFile != nil && len(*priFile) > 0 {
		log.Infof("try to read private key from %s", *priFile)
		reader, err := os.Open(*priFile)
		if err != nil {
			log.Fatal(err)
		}
		priKey, err = security.ReadECPrivateKey(reader)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Warn("no private file passed")
	}

	ecies := security.ECIESHelper{}
	return SecureContext{
		privateKey:  priKey,
		eciesHelper: ecies,
		pattern:     regexp.MustCompile(`^ENC\((.+)\)$`),
	}
}

func (s SecureContext) HasPrivateKey() bool {
	return s.privateKey != nil
}

func (s SecureContext) IsEncrypted(str string) bool {
	str = strings.TrimSpace(str)

	match := s.pattern.FindStringSubmatch(strings.TrimSpace(str))
	if match == nil {
		return false
	}

	_, err := base64.StdEncoding.DecodeString(match[1])
	return err == nil
}

func (s SecureContext) DecryptStr(str string) (string, error) {
	if s.privateKey == nil {
		return "", fmt.Errorf("no private key")
	}
	match := s.pattern.FindStringSubmatch(strings.TrimSpace(str))
	if match == nil {
		return str, nil
	}
	data, err := base64.StdEncoding.DecodeString(match[1])
	if err != nil {
		return "", err
	}
	decrypted, err := s.eciesHelper.DecryptWithPrivate(s.privateKey, data)
	if err != nil {
		return "", err
	}
	return string(decrypted), nil
}

func (s SecureContext) Decrypt(str string) ([]byte, error) {
	if s.privateKey == nil {
		return nil, fmt.Errorf("no private key")
	}
	match := s.pattern.FindStringSubmatch(str)
	if match == nil {
		return nil, fmt.Errorf("invalid encrypted string")
	}

	decrypted, err := s.eciesHelper.DecryptWithPrivate(s.privateKey, []byte(strings.TrimSpace(match[1])))
	if err != nil {
		return nil, err
	}
	return decrypted, nil
}
