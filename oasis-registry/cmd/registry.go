package cmd

import (
	"fmt"
	"os"

	"github.com/oasisprotocol/oasis-core/go/common/logging"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	registry "github.com/oasisprotocol/metadata-registry-tools"
)

const cfgUpdate = "update"

var (
	initCmd = &cobra.Command{
		Use:   "init",
		Short: "initialize a metadata registry in the current directory",
		Run:   doInit,
	}

	verifyCmd = &cobra.Command{
		Use:   "verify",
		Short: "verifies the integrity of the registry",
		Run:   doVerify,
	}

	verifyFlags = flag.NewFlagSet("", flag.ContinueOnError)

	registryLogger = logging.GetLogger("cmd/registry")
)

func newFsProvider() registry.MutableProvider {
	wd, err := os.Getwd()
	if err != nil {
		registryLogger.Error("failed to get current working directory",
			"err", err,
		)
		os.Exit(1)
	}

	p, err := registry.NewFilesystemPathProvider(wd)
	if err != nil {
		registryLogger.Error("failed to create filesystem provider",
			"err", err,
		)
		os.Exit(1)
	}

	return p
}

func doInit(cmd *cobra.Command, args []string) {
	p := newFsProvider()

	if err := p.Init(); err != nil {
		registryLogger.Error("failed to initialize registry",
			"err", err,
		)
		os.Exit(1)
	}

	fmt.Printf("Initialized metadata registry in %s\n", p.BaseDir())
}

func doVerify(cmd *cobra.Command, args []string) {
	p := newFsProvider()

	if err := p.Verify(); err != nil {
		registryLogger.Error("registry integrity verification failed",
			"err", err,
		)
		os.Exit(1)
	}

	updateFrom := viper.GetString(cfgUpdate)
	if updateFrom == "" {
		return
	}

	registryLogger.Info("verifying update from a previous registry snapshot",
		"src", updateFrom,
	)

	src, err := registry.NewFilesystemPathProvider(updateFrom)
	if err != nil {
		registryLogger.Error("failed to create filesystem provider for source registry",
			"err", err,
		)
		os.Exit(1)
	}

	if err = p.VerifyUpdate(src); err != nil {
		registryLogger.Error("update integrity verification failed",
			"err", err,
		)
		os.Exit(1)
	}
}

func init() { //nolint:gochecknoinits
	verifyFlags.String(cfgUpdate, "", "verify update from a previous registry snapshot")

	_ = viper.BindPFlags(verifyFlags)

	verifyCmd.Flags().AddFlagSet(verifyFlags)
}
