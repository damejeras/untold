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

func NewChangeCommand() subcommands.Command { return &changeCmd{environment: untold.DefaultEnvironment} }

func (c *changeCmd) Name() string { return "change-secret" }

func (c *changeCmd) Synopsis() string { return "change secret's value" }

func (c *changeCmd) Usage() string {
	return `untold change-secret [-env={environment}] [-key={decryption_key}] <secret_name>:
  Change secret's value.
`
}

func (c *changeCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.environment, "env", c.environment, "set environment")
	f.StringVar(&c.privateKey, "key", c.privateKey, "provide decryption key")
}

func (c *changeCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	name := f.Arg(0)
	if name == "" {
		cli.Errorf("argument \"name\" is required")
		c.Usage()

		return subcommands.ExitUsageError
	}

	environment :=  c.environment
	if environment == "" || environment == untold.DefaultEnvironment {
		environment = untold.DefaultEnvironment
		cli.Warnf("No environment provided, using default - %q", environment)
	}

	md5Hash := md5.Sum([]byte(name))
	base64EncodedPrivateKey := []byte(c.privateKey)

	if _, err := os.Stat(filepath.Join(environment, hex.EncodeToString(md5Hash[:]))); os.IsNotExist(err) {
		cli.Errorf("secret %q for %q environment not found", name, environment)

		return subcommands.ExitUsageError
	}

	if _, err := os.Stat(environment + ".public"); os.IsNotExist(err) {
		cli.Errorf("public key for %q environment not found", environment)

		return subcommands.ExitFailure
	}

	if len(base64EncodedPrivateKey) == 0 {
		if _, err := os.Stat(environment + ".private"); os.IsNotExist(err) {
			cli.Errorf("private key for %q environment not found", environment)

			return subcommands.ExitFailure
		}
	}

	base64EncodedPublicKey, err := os.ReadFile(environment + ".public")
	if err != nil {
		cli.Wrapf(err, "read public key for %q environment", environment)

		return subcommands.ExitFailure
	}

	if len(base64EncodedPrivateKey) == 0 {
		base64EncodedPrivateKey, err = os.ReadFile(environment + ".private")
		if err != nil {
			cli.Wrapf(err, "read private key for %q environment", environment)

			return subcommands.ExitFailure
		}
	}

	base64EncodedContent, err := os.ReadFile(filepath.Join(environment, hex.EncodeToString(md5Hash[:])))
	if err != nil {
		cli.Wrapf(err, "read secret %q for %q environment", name, environment)

		return subcommands.ExitFailure
	}

	var publicKey, privateKey [32]byte
	publicKey, err = untold.DecodeBase64Key(base64EncodedPublicKey)
	if err != nil {
		cli.Wrapf(err, "decode base64 encoded public key for %q environment", environment)

		return subcommands.ExitFailure
	}

	privateKey, err = untold.DecodeBase64Key(base64EncodedPrivateKey)
	if err != nil {
		cli.Wrapf(err, "decode base64 encoded private key for %q environment", environment)

		return subcommands.ExitFailure
	}

	encryptedSecret, err := untold.Base64Decode(base64EncodedContent)
	if err != nil {
		cli.Wrapf(err, "decode base64 secret %q for %q environment", name, environment)

		return subcommands.ExitFailure
	}

	_, ok := box.OpenAnonymous(nil, encryptedSecret, &publicKey, &privateKey)
	if !ok {
		cli.Errorf("can not decrypt %q secret for %q environment", name, environment)

		return subcommands.ExitFailure
	}

	fmt.Printf("Enter new value for %q secret in %q environment:\n", name, environment)

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
		cli.Wrapf(err, "write secret %q for %q environment to file", name, environment)

		return subcommands.ExitFailure
	}

	cli.Successf("Secret's %q value changed", name)

	return subcommands.ExitSuccess
}
