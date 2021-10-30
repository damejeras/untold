package vault

import (
	"context"
	"crypto/rand"
	"flag"
	"github.com/damejeras/untold"
	"github.com/damejeras/untold/internal/cli"
	"github.com/google/subcommands"
	"golang.org/x/crypto/nacl/box"
	"os"
	"path/filepath"
)

type createCmd struct {}

func NewCreateCommand() subcommands.Command { return &createCmd{} }

func (c *createCmd) Name() string { return "new-env" }

func (c *createCmd) Synopsis() string { return "create new environment" }

func (c *createCmd) Usage() string {
	return `untold new-env <environment_name>:
  Create new environment.
`
}

func (c createCmd) SetFlags(f *flag.FlagSet) {}

func (c createCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	environmentName := f.Arg(0)
	if environmentName == "" {
		cli.Errorf("argument \"environment_name\" is required")
		c.Usage()

		return subcommands.ExitUsageError
	}

	if _, err := os.Stat(environmentName); !os.IsNotExist(err) {
		cli.Errorf("directory %q already exists", environmentName)

		return subcommands.ExitUsageError
	}

	if _, err := os.Stat(environmentName+".public"); !os.IsNotExist(err) {
		cli.Errorf("file %q already exists", environmentName+".public")

		return subcommands.ExitUsageError
	}

	if _, err := os.Stat(environmentName+".private"); !os.IsNotExist(err) {
		cli.Errorf("file %q already exists", environmentName+".private")

		return subcommands.ExitUsageError
	}

	if err := os.Mkdir(environmentName, 0755); err != nil {
		cli.Wrapf(err, "create directory %q", environmentName)

		return subcommands.ExitFailure
	}

	if err := os.WriteFile(filepath.Join(environmentName, ".gitkeep"), []byte("*"), 0644); err != nil {
		cli.Wrapf(err, "create .gitkeep file for %q environment", environmentName)

		return subcommands.ExitFailure
	}

	publicKey, privateKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		cli.Wrapf(err, "generate keypair")

		return subcommands.ExitFailure
	}

	if err := os.WriteFile(filepath.Join(environmentName+".public"), untold.Base64Encode(publicKey[:]), 0644); err != nil {
		cli.Wrapf(err, "write public key for environment %q", environmentName)

		return subcommands.ExitFailure
	}

	if err := os.WriteFile(filepath.Join(environmentName+".private"), untold.Base64Encode(privateKey[:]), 0644); err != nil {
		cli.Wrapf(err, "write private key for environment %q", environmentName)

		return subcommands.ExitFailure
	}

	cli.Successf("Environment %q created.", environmentName)

	return subcommands.ExitSuccess
}
