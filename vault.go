package untold

import (
	"crypto/md5"
	"embed"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/nacl/box"
	"os"
	"path/filepath"
)

const (
	DefaultPathPrefix          = "untold"
	DefaultEnvironment         = "development"
	DefaultEnvironmentVariable = "UNTOLD_KEY"
)

var (
	zeroKey [32]byte
)

type Vault interface {
	Load(interface{}) error
}

type vault struct {
	embeddedFiles                          embed.FS
	pathPrefix, environment, privateKeyEnv string
	publicKey, privateKey                  [32]byte
}

func NewVault(files embed.FS, options ...Option) Vault {
	v := vault{
		embeddedFiles: files,
		pathPrefix:    DefaultPathPrefix,
		environment:   DefaultEnvironment,
		privateKeyEnv: DefaultEnvironmentVariable,
		privateKey:    zeroKey,
		publicKey:     zeroKey,
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
	if v.privateKey != zeroKey || v.publicKey != zeroKey {
		return nil
	}

	base64PublicKey, err := v.embeddedFiles.ReadFile(filepath.Join(v.pathPrefix, v.environment+".public"))
	if err != nil {
		return fmt.Errorf("read public key file for %q environment: %s", v.environment, err)
	}

	base64PrivateKey := []byte(os.Getenv(v.privateKeyEnv))
	if len(base64PrivateKey) == 0 {
		base64PrivateKey, err = v.embeddedFiles.ReadFile(filepath.Join(v.pathPrefix, v.environment+".private"))
		if err != nil {
			return fmt.Errorf("read private key file for %q environment: %s", v.environment, err)
		}
	}

	v.publicKey, err = DecodeBase64Key(base64PublicKey)
	if err != nil {
		return fmt.Errorf("decode base64 encoded public key: %s", err)
	}

	v.privateKey, err = DecodeBase64Key(base64PrivateKey)
	if err != nil {
		return fmt.Errorf("decode base64 encoded private key: %s", err)
	}

	return nil
}

func (v *vault) findSecret(name string) (string, error) {
	md5Hash := md5.Sum([]byte(name))

	base64EncodedSecret, err := v.embeddedFiles.ReadFile(filepath.Join(v.pathPrefix, v.environment, hex.EncodeToString(md5Hash[:])))
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("secret %q for %q environemnt not found", name, v.environment)
		}

		return "", fmt.Errorf("get secret for %q for %q environment: %s", name, v.environment, err)
	}

	decodedSecret, err := Base64Decode(base64EncodedSecret)
	if err != nil {
		return "", fmt.Errorf("base64 decode %q for %q environment: %s", name, v.environment, err)
	}

	decrypted, ok := box.OpenAnonymous(nil, decodedSecret, &v.publicKey, &v.privateKey)
	if !ok {
		return "", fmt.Errorf("can not decrypt secret %q for %q environemnt: %s", name, v.environment, err)
	}

	return string(decrypted), nil
}
