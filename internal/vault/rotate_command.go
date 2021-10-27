package vault

import (
	"context"
	"crypto/rand"
	"flag"
	"github.com/damejeras/untold"
	"github.com/damejeras/untold/internal/cli"
	"github.com/google/subcommands"
	"golang.org/x/crypto/nacl/box"
	"io/ioutil"
	"os"
	"path/filepath"
)

type rotateCmd struct {
	privateKey string
}

func NewRotateCommand() subcommands.Command {
	return rotateCmd{}
}
func (r rotateCmd) Name() string {
	return "rotate-keys"
}

func (r rotateCmd) Synopsis() string {
	return "rotate environment keys"
}

func (r rotateCmd) Usage() string {
	return `rotate-keys <environment_name>:
  Rotate environment keys.
`
}

func (r rotateCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&r.privateKey, "key", r.privateKey, "provide decryption key")
}

func (r rotateCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	environmentName := f.Arg(0)
	if environmentName == "" {
		cli.Errorf("argument \"environment_name\" is required")
		r.Usage()

		return subcommands.ExitUsageError
	}

	if _, err := os.Stat(environmentName); os.IsNotExist(err) {
		cli.Errorf("directory for \"%s\" environment not found", environmentName)

		return subcommands.ExitUsageError
	}

	if _, err := os.Stat(environmentName+".public"); os.IsNotExist(err) {
		cli.Errorf("public key for \"%s\" environment not found", environmentName)

		return subcommands.ExitUsageError
	}

	if _, err := os.Stat(environmentName+".private"); os.IsNotExist(err) {
		cli.Errorf("private key for \"%s\" environment not found", environmentName)

		return subcommands.ExitUsageError
	}

	base64EncodedPublicKey, err := os.ReadFile(environmentName + ".public")
	if err != nil {
		cli.Wrapf(err, "read public key for \"%s\" environment", environmentName)

		return subcommands.ExitFailure
	}

	base64EncodedPrivateKey := []byte(r.privateKey)

	if len(base64EncodedPrivateKey) == 0 {
		base64EncodedPrivateKey, err = os.ReadFile(environmentName + ".private")
		if err != nil {
			cli.Wrapf(err, "read private key for \"%s\" environment", environmentName)

			return subcommands.ExitFailure
		}
	}

	decodedPublicKey, err := untold.Base64Decode(base64EncodedPublicKey)
	if err != nil {
		cli.Wrapf(err, "decode public key for \"%s\" environment", environmentName)

		return subcommands.ExitFailure
	}

	decodedPrivateKey, err := untold.Base64Decode(base64EncodedPrivateKey)
	if err != nil {
		cli.Wrapf(err, "decode private key for \"%s\" environment", environmentName)

		return subcommands.ExitFailure
	}

	var publicKey, privateKey [32]byte
	copy(publicKey[:], decodedPublicKey)
	copy(privateKey[:], decodedPrivateKey)

	files, err := ioutil.ReadDir(environmentName)
	if err != nil {
		cli.Wrapf(err, "read environment \"%s\" secrets", environmentName)

		return subcommands.ExitFailure
	}

	values := make(map[string]string)

	for _, file := range files {
		filename := file.Name()
		if filename == ".gitkeep" {
			continue
		}

		base64EncodedSecret, err := os.ReadFile(filepath.Join(environmentName, file.Name()))
		if err != nil {
			cli.Wrapf(err, "read \"%s\" file", filepath.Join(environmentName, file.Name()))

			return subcommands.ExitFailure
		}

		decodedSecret, err := untold.Base64Decode(base64EncodedSecret)
		if err != nil {
			cli.Wrapf(err, "decode secret \"%s\"", file.Name())

			return subcommands.ExitFailure
		}

		decryptedSecret, ok := box.OpenAnonymous(nil, decodedSecret, &publicKey, &privateKey)
		if !ok {
			cli.Errorf("can not decode secret \"%s\"", file.Name())

			return subcommands.ExitFailure
		}

		values[file.Name()] = string(decryptedSecret)
	}

	newPublicKey, newPrivateKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		cli.Wrapf(err, "generate new keypair")

		return subcommands.ExitFailure
	}

	for filename, value := range values {
		encryptedValue, err := box.SealAnonymous(nil, []byte(value), newPublicKey, rand.Reader)
		if err != nil {
			cli.Wrapf(err, "encrypt \"%s\" value", filename)

			return subcommands.ExitFailure
		}

		err = os.WriteFile(filepath.Join(environmentName, filename), untold.Base64Encode(encryptedValue), 0644)
		if err != nil {
			cli.Wrapf(err, "write encrypted value to \"%s\"", filename)

			return subcommands.ExitFailure
		}
	}

	if err := os.WriteFile(environmentName+".public", untold.Base64Encode(newPublicKey[:]), 0644); err != nil {
		cli.Wrapf(err, "write new public key")

		return subcommands.ExitFailure
	}

	if err := os.WriteFile(environmentName+".private", untold.Base64Encode(newPrivateKey[:]), 0644); err != nil {
		cli.Wrapf(err, "write new private key")

		return subcommands.ExitFailure
	}

	cli.Successf("Keys for environment \"%s\" rotated", environmentName)

	return subcommands.ExitSuccess
}
