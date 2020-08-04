package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/oasisprotocol/oasis-core/go/common/crypto/signature"
	"github.com/oasisprotocol/oasis-core/go/common/logging"
	cmdSigner "github.com/oasisprotocol/oasis-core/go/oasis-node/cmd/common/signer"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	registry "github.com/oasisprotocol/metadata-registry-tools"
)

var (
	entityCmd = &cobra.Command{
		Use:   "entity",
		Short: "entity-related subcommands",
	}

	entityCreateCmd = &cobra.Command{
		Use:   "update",
		Short: "updates an entity in the registry",
		Run:   doEntityCreate,
	}

	entityFlags = flag.NewFlagSet("", flag.ContinueOnError)

	entityLogger = logging.GetLogger("cmd/entity")
)

func doEntityCreate(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		entityLogger.Error("expected a single argument")
		os.Exit(1)
	}

	p := newFsProvider()

	// Open and parse the passed entity metadata file.
	rawEntity, err := ioutil.ReadFile(args[0])
	if err != nil {
		entityLogger.Error("failed to read entity descriptor",
			"err", err,
		)
		os.Exit(1)
	}

	var entity registry.EntityMetadata
	if err = json.Unmarshal(rawEntity, &entity); err != nil {
		entityLogger.Error("failed to parse serialized entity metadata",
			"err", err,
		)
		os.Exit(1)
	}

	if err = entity.ValidateBasic(); err != nil {
		entityLogger.Error("provided entity metadata is invalid",
			"err", err,
		)
		os.Exit(1)
	}

	// Get the signer.
	signerDir, err := cmdSigner.CLIDirOrPwd()
	if err != nil {
		entityLogger.Error("failed to retrieve signer dir",
			"err", err,
		)
		os.Exit(1)
	}
	signerFactory, err := cmdSigner.NewFactory(cmdSigner.Backend(), signerDir, signature.SignerEntity)
	if err != nil {
		entityLogger.Error("failed to create signer factory",
			"err", err,
		)
		os.Exit(1)
	}
	signer, err := signerFactory.Load(signature.SignerEntity)
	if err != nil {
		entityLogger.Error("failed to load signer",
			"err", err,
		)
		os.Exit(1)
	}

	// Sign the descriptor.
	signed, err := registry.SignEntityMetadata(signer, &entity)
	if err != nil {
		entityLogger.Error("failed to sign metadata",
			"err", err,
		)
		os.Exit(1)
	}

	if err = p.UpdateEntity(signed); err != nil {
		entityLogger.Error("failed to update metadata",
			"err", err,
		)
		os.Exit(1)
	}

	fmt.Printf("Updated entity %s\n", signer.Public())
}

func init() {
	entityFlags.AddFlagSet(cmdSigner.Flags)
	entityFlags.AddFlagSet(cmdSigner.CLIFlags)
	entityCreateCmd.Flags().AddFlagSet(entityFlags)

	// Register all of the sub-commands.
	entityCmd.AddCommand(entityCreateCmd)
}
