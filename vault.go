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
	DefaultPathPrefix          = "untold"
	DefaultEnvironment         = "development"
	DefaultEnvironmentVariable = "UNTOLD_KEY"
)

type Vault interface {
	Load(interface{}) error
}

type vault struct {
	embeddedFiles                          embed.FS
	pathPrefix, environment, privateKeyEnv string
	publicKey, privateKey                  []byte
}

func NewVault(files embed.FS, options ...Option) Vault {
	v := vault{
		embeddedFiles: files,
		pathPrefix:    DefaultPathPrefix,
		environment:   DefaultEnvironment,
		privateKeyEnv: DefaultEnvironmentVariable,
	}

	for i := range options {
		options[i](&v)
	}

	return &v
}

func (v *vault) Load(dst interface{}) error {
	if err := v.loadSecrets(); err != nil {
		return err
	}

	return parse(dst, v.findSecret)
}

func (v *vault) loadSecrets() error {
	if v.privateKey != nil || v.publicKey != nil {
		return nil
	}

	base64PublicKey, err := v.embeddedFiles.ReadFile(filepath.Join(v.pathPrefix, v.environment+".public"))
	if err != nil {
		if os.IsNotExist(err) {
			return errors.Errorf("public key for \"%s\" environment not found", v.environment)
		}

		return errors.Wrapf(err, "read public key file for \"%s\" environment", v.environment)
	}

	base64PrivateKey := []byte(os.Getenv(v.privateKeyEnv))
	if len(base64PrivateKey) == 0 {
		base64PrivateKey, err = v.embeddedFiles.ReadFile(filepath.Join(v.pathPrefix, v.environment+".private"))
		if err != nil {
			if os.IsNotExist(err) {
				return errors.Errorf("can not find private key for environment \"%s\"", v.environment)
			}

			return errors.Wrapf(err, "read private key file for \"%s\" environment", v.environment)
		}
	}

	v.publicKey, err = Base64Decode(base64PublicKey)
	if err != nil {
		return errors.Wrap(err, "base64 decode public key")
	}

	v.privateKey, err = Base64Decode(base64PrivateKey)
	if err != nil {
		return errors.Wrapf(err, "base64 decode private key")
	}

	return nil
}

func (v *vault) findSecret(name string) (string, error) {
	md5Hash := md5.Sum([]byte(name))

	base64EncodedSecret, err := v.embeddedFiles.ReadFile(filepath.Join(v.pathPrefix, v.environment, hex.EncodeToString(md5Hash[:])))
	if err != nil {
		if os.IsNotExist(err) {
			return "", errors.Errorf("secret \"%s\" for \"%s\" environemnt not found", name, v.environment)
		}

		return "", errors.Wrapf(err, "get secret for \"%s\" for \"%s\" environment", name, v.environment)
	}

	decodedSecret, err := Base64Decode(base64EncodedSecret)
	if err != nil {
		return "", errors.Wrapf(err, "base64 decode \"%s\" for \"%s\" environment", name, v.environment)
	}

	var publicKey, privateKey [32]byte
	copy(publicKey[:], v.publicKey)
	copy(privateKey[:], v.privateKey)

	decrypted, ok := box.OpenAnonymous(nil, decodedSecret, &publicKey, &privateKey)
	if !ok {
		return "", errors.Errorf("can not decrypt secret \"%s\" for \"%s\" environemnt", name, v.environment)
	}

	return string(decrypted), nil
}
