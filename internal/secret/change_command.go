package secret

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/damejeras/untold"
	"github.com/damejeras/untold/internal/cli"
	"github.com/google/subcommands"
	"golang.org/x/crypto/nacl/box"
	"os"
	"path/filepath"
)

type changeCmd struct {
	environment, privateKey string
}

func NewChangeCommand() subcommands.Command {
	return changeCmd{environment: untold.DefaultEnvironment}
}

func (c changeCmd) Name() string {
	return "change-secret"
}

func (c changeCmd) Synopsis() string {
	return "change secret's value"
}

func (c changeCmd) Usage() string {
	return `untold change-secret [-env={environment}] [-key={decryption_key}] <secret_name>:
  Change secret's value.
`
}

func (c changeCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.environment, "env", c.environment, "set environment")
	f.StringVar(&c.privateKey, "key", c.privateKey, "provide decryption key")
}

func (c changeCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	name := f.Arg(0)
	if name == "" {
		cli.Errorf("argument \"name\" is required")
		c.Usage()

		return subcommands.ExitUsageError
	}

	environment :=  c.environment
	if environment == "" || environment == untold.DefaultEnvironment {
		environment = untold.DefaultEnvironment
		cli.Warnf("No environment provided, using default - \"%s\"", environment)
	}

	md5Hash := md5.Sum([]byte(name))
	base64EncodedPrivateKey := []byte(c.privateKey)

	if _, err := os.Stat(filepath.Join(environment, hex.EncodeToString(md5Hash[:]))); os.IsNotExist(err) {
		cli.Errorf("secret \"%s\" for \"%s\" environment not found", name, environment)

		return subcommands.ExitUsageError
	}

	if _, err := os.Stat(environment + ".public"); os.IsNotExist(err) {
		cli.Errorf("public key for \"%s\" environment not found", environment)

		return subcommands.ExitFailure
	}

	if len(base64EncodedPrivateKey) == 0 {
		if _, err := os.Stat(environment + ".private"); os.IsNotExist(err) {
			cli.Errorf("private key for \"%s\" environment not found", environment)

			return subcommands.ExitFailure
		}
	}

	base64EncodedPublicKey, err := os.ReadFile(environment + ".public")
	if err != nil {
		cli.Wrapf(err, "read public key for \"%s\" environment", environment)

		return subcommands.ExitFailure
	}

	if len(base64EncodedPrivateKey) == 0 {
		base64EncodedPrivateKey, err = os.ReadFile(environment + ".private")
		if err != nil {
			cli.Wrapf(err, "read private key for \"%s\" environment", environment)

			return subcommands.ExitFailure
		}
	}

	decodedPublicKey, err := untold.Base64Decode(base64EncodedPublicKey)
	if err != nil {
		cli.Wrapf(err, "decode public key for \"%s\" environment", environment)

		return subcommands.ExitFailure
	}

	decodedPrivateKey, err := untold.Base64Decode(base64EncodedPrivateKey)
	if err != nil {
		cli.Wrapf(err, "decode private key for \"%s\" environment", environment)

		return subcommands.ExitFailure
	}

	if err != nil {
		cli.Wrapf(err, "decode secret \"%s\" key for \"%s\" environment", name, environment)

		return subcommands.ExitFailure
	}

	var publicKey, privateKey [32]byte
	copy(publicKey[:], decodedPublicKey)
	copy(privateKey[:], decodedPrivateKey)

	fmt.Printf("Enter value for \"%s\" secret in \"%s\" environment:\n", name, environment)

	var value string
	_, err = fmt.Scanln(&value)
	if err != nil {
		cli.Wrapf(err, "read user input")

		return subcommands.ExitFailure
	}


	encryptedValue, err := box.SealAnonymous(nil, []byte(value), &publicKey, rand.Reader)
	if err != nil {
		cli.Wrapf(err, "encrypt user input")

		return subcommands.ExitFailure
	}

	err = os.WriteFile(filepath.Join(environment, hex.EncodeToString(md5Hash[:])), untold.Base64Encode(encryptedValue), 0644)
	if err != nil {
		cli.Wrapf(err, "write secret \"%s\" for \"%s\" environment to file", name, environment)

		return subcommands.ExitFailure
	}

	cli.Successf("Secret's \"%s\" value changed", name)

	return subcommands.ExitSuccess
}


