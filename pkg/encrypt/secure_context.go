package encrypt

import (
	"crypto/ecdsa"
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
	reader, err := os.Open(*priFile)
	if err != nil {
		log.Fatal(err)
	}
	priKey, err := security.ReadECPrivateKey(reader)
	if err != nil {
		log.Fatal(err)
	}
	ecies := security.ECIESHelper{}
	return SecureContext{
		privateKey:  priKey,
		eciesHelper: ecies,
		pattern:     regexp.MustCompile(`^ENC(.+)$`),
	}
}

func (s SecureContext) IsEncrypted(str string) bool {
	return s.pattern.MatchString(strings.TrimSpace(str))
}

func (s SecureContext) DecryptStr(str string) (string, error) {
	match := s.pattern.FindStringSubmatch(strings.TrimSpace(str))
	if match == nil {
		return str, nil
	}

	decrypted, err := s.eciesHelper.DecryptWithPrivate(s.privateKey, []byte(match[1]))
	if err != nil {
		return "", err
	}
	return string(decrypted), nil
}

func (s SecureContext) Decrypt(str string) ([]byte, error) {
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
