package registry

import (
	"context"
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/oasisprotocol/oasis-core/go/common/cbor"
	memorySigner "github.com/oasisprotocol/oasis-core/go/common/crypto/signature/signers/memory"
	"github.com/stretchr/testify/require"
)

func TestFilesystemProvider(t *testing.T) {
	require := require.New(t)

	fs := memfs.New()
	fp, err := NewFilesystemProvider(fs)
	require.NoError(err, "NewFilesystemProvider")

	err = fp.Init()
	require.NoError(err, "Init")
	err = fp.Init()
	require.Error(err, "Init should fail when already initialized")

	err = fp.Verify()
	require.NoError(err, "Verify should work on an empty registry")

	signer := memorySigner.NewTestSigner("metadata-registry-tools test entity signer")
	entity := &EntityMetadata{
		Versioned: cbor.NewVersioned(1),
		Serial:    1,
		Name:      "hello world",
		URL:       "https://helloworld.io",
		Email:     "hello@world.org",
		Keybase:   "helloworld",
		Twitter:   "helloworld",
	}
	signed, err := SignEntityMetadata(signer, entity)
	require.NoError(err, "SignEntityMetadata")
	err = fp.UpdateEntity(signed)
	require.NoError(err, "UpdateEntity")
	err = fp.UpdateEntity(signed)
	require.Error(err, "UpdateEntity should fail if serial number is not bumped")

	fetchedEntity, err := fp.GetEntity(context.Background(), signer.Public())
	require.NoError(err, "GetEntity")
	require.EqualValues(entity, fetchedEntity, "GetEntity should return the same entity")

	entity.Name = "another hello world"
	entity.Serial++
	signed, err = SignEntityMetadata(signer, entity)
	require.NoError(err, "SignEntityMetadata")
	err = fp.UpdateEntity(signed)
	require.NoError(err, "UpdateEntity")

	fetchedEntity, err = fp.GetEntity(context.Background(), signer.Public())
	require.NoError(err, "GetEntity")
	require.EqualValues(entity, fetchedEntity, "GetEntity should return the same entity")

	entities, err := fp.GetEntities(context.Background())
	require.NoError(err, "GetEntities")
	require.Len(entities, 1)
	require.EqualValues(entity, entities[signer.Public()])
}
