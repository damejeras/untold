package untold

import (
	"crypto/md5"
	"embed"
	"encoding/hex"
	"github.com/pkg/errors"
	"golang.org/x/crypto/nacl/box"
	"os"
	"path/filepath"
)

const (
	defaultPathPrefix                    = "untold"
	defaultEnvironment                   = "development"
	defaultPrivateKeyEnvironmentVariable = "UNTOLD_KEY"
)

type Loader interface {
	Load(interface{}) error
}

type loader struct {
	embeddedFiles                          embed.FS
	pathPrefix, environment, privateKeyEnv string
	publicKey, privateKey                  []byte
}

func NewLoader(files embed.FS, options ...Option) Loader {
	l := loader{
		embeddedFiles: files,
		pathPrefix:    defaultPathPrefix,
		environment:   defaultEnvironment,
		privateKeyEnv: defaultPrivateKeyEnvironmentVariable,
	}

	for i := range options {
		options[i](&l)
	}

	return &l
}

func (l *loader) Load(dst interface{}) error {
	if err := l.loadKeys(); err != nil {
		return err
	}

	return parse(dst, l.findSecret)
}

func (l *loader) loadKeys() error {
	if l.privateKey != nil || l.publicKey != nil {
		return nil
	}

	base64PublicKey, err := l.embeddedFiles.ReadFile(filepath.Join(l.pathPrefix, l.environment+".public"))
	if err != nil {
		if os.IsNotExist(err) {
			return errors.Errorf("public key for \"%s\" environment not found", l.environment)
		}

		return errors.Wrapf(err, "read public key file for \"%s\" environment", l.environment)
	}

	base64PrivateKey := []byte(os.Getenv(l.privateKeyEnv))
	if len(base64PrivateKey) == 0 {
		base64PrivateKey, err = l.embeddedFiles.ReadFile(filepath.Join(l.pathPrefix, l.environment+".private"))
		if err != nil {
			if os.IsNotExist(err) {
				return errors.Errorf("can not find private key for environment \"%s\"", l.environment)
			}

			return errors.Wrapf(err, "read private key file for \"%s\" environment", l.environment)
		}
	}

	l.publicKey, err = Base64Decode(base64PublicKey)
	if err != nil {
		return errors.Wrap(err, "base64 decode public key")
	}

	l.privateKey, err = Base64Decode(base64PrivateKey)
	if err != nil {
		return errors.Wrapf(err, "base64 decode private key")
	}

	return nil
}

func (l *loader) findSecret(name string) (string, error) {
	md5Hash := md5.Sum([]byte(name))

	base64EncodedSecret, err := l.embeddedFiles.ReadFile(filepath.Join(l.pathPrefix, l.environment, hex.EncodeToString(md5Hash[:])))
	if err != nil {
		if os.IsNotExist(err) {
			return "", errors.Errorf("secret \"%s\" for \"%s\" environemnt not found", name, l.environment)
		}

		return "", errors.Wrapf(err, "get secret for \"%s\" for \"%s\" environment", name, l.environment)
	}

	decodedSecret, err := Base64Decode(base64EncodedSecret)
	if err != nil {
		return "", errors.Wrapf(err, "base64 decode \"%s\" for \"%s\" environment", name, l.environment)
	}

	decrypted, ok := box.OpenAnonymous(nil, decodedSecret, (*[32]byte)(l.publicKey), (*[32]byte)(l.privateKey))
	if !ok {
		return "", errors.Errorf("can not decrypt secret \"%s\" for \"%s\" environemnt", name, l.environment)
	}

	return string(decrypted), nil
}
