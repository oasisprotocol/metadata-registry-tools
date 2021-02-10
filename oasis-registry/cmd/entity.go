package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/oasisprotocol/oasis-core/go/common/crypto/signature"
	signerFile "github.com/oasisprotocol/oasis-core/go/common/crypto/signature/signers/file"
	signerPlugin "github.com/oasisprotocol/oasis-core/go/common/crypto/signature/signers/plugin"
	"github.com/oasisprotocol/oasis-core/go/common/logging"
	cmdCommon "github.com/oasisprotocol/oasis-core/go/oasis-node/cmd/common"
	cmdFlags "github.com/oasisprotocol/oasis-core/go/oasis-node/cmd/common/flags"
	cmdSigner "github.com/oasisprotocol/oasis-core/go/oasis-node/cmd/common/signer"

	registry "github.com/oasisprotocol/metadata-registry-tools"
)

// cfgSkipValidation configures whether the validation of the provided entity
// metadata should be skipped or not.
const cfgSkipValidation = "skip-validation"

var (
	entityCmd = &cobra.Command{
		Use:   "entity",
		Short: "entity-related subcommands",
	}

	entityUpdateCmd = &cobra.Command{
		Use:   "update",
		Short: "update (or create) an entity in the registry",
		Run:   doEntityUpdate,
	}

	entityFlags = flag.NewFlagSet("", flag.ContinueOnError)

	entityLogger = logging.GetLogger("cmd/entity")
)

func logErrorAndExit(msg string, err error) {
	entityLogger.Error(msg, "err", err)
	os.Exit(1)
}

func loadSigner() (signature.Signer, error) {
	signerDir, err := cmdSigner.CLIDirOrPwd()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve signer dir: %w", err)
	}
	signerFactory, err := cmdSigner.NewFactory(cmdSigner.Backend(), signerDir, signature.SignerEntity)
	if err != nil {
		return nil, fmt.Errorf("failed to create signer factory: %w", err)
	}
	signer, err := signerFactory.Load(signature.SignerEntity)
	if err != nil {
		return nil, fmt.Errorf("failed to load signer: %w", err)
	}
	return signer, nil
}

func doEntityUpdate(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		entityLogger.Error("expected a single argument")
		os.Exit(1)
	}

	p := newFsProvider()

	// Open and parse the passed entity metadata file.
	rawEntity, err := ioutil.ReadFile(args[0])
	if err != nil {
		logErrorAndExit("failed to read entity descriptor", err)
	}

	var entity registry.EntityMetadata
	if err = json.Unmarshal(rawEntity, &entity); err != nil {
		logErrorAndExit("failed to parse serialized entity metadata", err)
	}

	if !viper.GetBool(cfgSkipValidation) {
		if err = entity.ValidateBasic(); err != nil {
			logErrorAndExit("provided entity metadata is invalid", err)
		}
	}

	// Get the signer.
	signer, err := loadSigner()
	if err != nil {
		logErrorAndExit("failed to load signer", err)
	}

	// Show descriptor and ask for confirmation.
	fmt.Printf("You are about to sign the following entity metadata descriptor:\n")
	entity.PrettyPrint(context.Background(), "  ", os.Stdout)

	switch cmdSigner.Backend() {
	case signerFile.SignerName:
		if !cmdFlags.AssumeYes() {
			if !cmdCommon.GetUserConfirmation("\nAre you sure you want to continue? (y)es/(n)o: ") {
				os.Exit(1)
			}
		}
	case signerPlugin.SignerName:
		if cmdCommon.Isatty(os.Stdin.Fd()) {
			fmt.Println("\nYou may need to review the transaction on your device if you use a hardware-based signer plugin...")
		}
	}

	// Sign the descriptor.
	signed, err := registry.SignEntityMetadata(signer, &entity)
	if err != nil {
		logErrorAndExit("failed to sign metadata", err)
	}

	if err = p.UpdateEntity(signed); err != nil {
		logErrorAndExit("failed to update metadata", err)
	}

	fmt.Printf("Updated entity %s\n", signer.Public())
}

func init() {
	entityFlags.Bool(cfgSkipValidation, false, "skip metadata validation")
	entityFlags.AddFlagSet(cmdSigner.Flags)
	entityFlags.AddFlagSet(cmdSigner.CLIFlags)
	entityFlags.AddFlagSet(cmdFlags.AssumeYesFlag)
	_ = viper.BindPFlags(entityFlags)

	entityUpdateCmd.Flags().AddFlagSet(entityFlags)

	// Register all of the sub-commands.
	entityCmd.AddCommand(entityUpdateCmd)
}
