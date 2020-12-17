package testvectors

import (
	"github.com/oasisprotocol/oasis-core/go/common/cbor"
	"github.com/oasisprotocol/oasis-core/go/common/crypto/signature"
	memorySigner "github.com/oasisprotocol/oasis-core/go/common/crypto/signature/signers/memory"

	registry "github.com/oasisprotocol/metadata-registry-tools"
)

const keySeedPrefix = "oasis-metadata-registry test vectors: "

// EntityMetadataTestVector is an entity metadata test vector.
type EntityMetadataTestVector struct {
	Kind                    string                        `json:"kind"`
	SignatureContext        string                        `json:"signature_context"`
	EntityMeta              registry.EntityMetadata       `json:"entity_meta"`
	SignedEntityMeta        registry.SignedEntityMetadata `json:"signed_entity_meta"`
	EncodedEntityMeta       []byte                        `json:"encoded_entity_meta"`
	EncodedSignedEntityMeta []byte                        `json:"encoded_signed_entity_meta"`
	Valid                   bool                          `json:"valid"`
	SignerPrivateKey        []byte                        `json:"signer_private_key"`
	SignerPublicKey         signature.PublicKey           `json:"signer_public_key"`
}

// MakeEntityMetadataTestVector generates a new test vector from an entity metadata.
func MakeEntityMetadataTestVector(kind string, meta *registry.EntityMetadata, valid bool) EntityMetadataTestVector {
	signer := memorySigner.NewTestSigner(keySeedPrefix + kind)
	return MakeEntityMetadataTestVectorWithSigner(kind, meta, valid, signer)
}

// MakeEntityMetadataTestVectorWithSigner generates a new test vector from an entity metadata using a specific signer.
func MakeEntityMetadataTestVectorWithSigner(
	kind string,
	meta *registry.EntityMetadata,
	valid bool,
	signer signature.Signer,
) EntityMetadataTestVector {
	sigMeta, err := registry.SignEntityMetadata(signer, meta)
	if err != nil {
		panic(err)
	}

	sigCtx, err := signature.PrepareSignerContext(registry.EntityMetadataSignatureContext)
	if err != nil {
		panic(err)
	}

	return EntityMetadataTestVector{
		Kind:                    kind,
		SignatureContext:        string(sigCtx),
		EntityMeta:              *meta,
		SignedEntityMeta:        *sigMeta,
		EncodedEntityMeta:       cbor.Marshal(meta),
		EncodedSignedEntityMeta: cbor.Marshal(sigMeta),
		Valid:                   valid,
		SignerPrivateKey:        signer.(signature.UnsafeSigner).UnsafeBytes(),
		SignerPublicKey:         signer.Public(),
	}
}
