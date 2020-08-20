package registry

import (
	"context"
	"errors"
	"testing"

	"github.com/oasisprotocol/oasis-core/go/common/crypto/signature"
	"github.com/stretchr/testify/require"
)

func TestGitProvider(t *testing.T) {
	require := require.New(t)

	gp, err := NewGitProvider(NewTestGitConfig())
	require.NoError(err, "NewGitProvider")

	ctx := context.Background()

	// Non-existing entity.
	_, err = gp.GetEntity(ctx, signature.PublicKey{})
	require.Equal(ErrNoSuchEntity, err, "GetEntity should fail for non-existing entity")

	// Existing entity which can be verified.
	var entityID signature.PublicKey
	_ = entityID.UnmarshalHex("9391840cee32fd13283e7e383838ad89f9704e3860a178db35a372b8b9c2cf10")
	entity, err := gp.GetEntity(ctx, entityID)
	require.NoError(err, "GetEntity")
	require.Equal("Hello World Entity", entity.Name)
	require.Equal("https://oasisprotocol.org", entity.URL)

	// Mismatched signer and filename.
	_ = entityID.UnmarshalHex("9391840cee32fd13283e7e383838ad89f9704e3860a178db35a372b8b9c2cf1f")
	_, err = gp.GetEntity(ctx, entityID)
	require.True(errors.Is(err, ErrCorruptedRegistry), "GetEntity should fail for mismatched signer/filename")

	// Corrupted payload.
	_ = entityID.UnmarshalHex("5fb36d105a6c85a21542abdd712c292ae37425e842b38db450970cdef0780bd8")
	_, err = gp.GetEntity(ctx, entityID)
	require.True(errors.Is(err, ErrCorruptedRegistry), "GetEntity should fail for corrupted payload")
}
