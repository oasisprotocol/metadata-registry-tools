// gen_vectors generates test vectors for entity metadata descriptors.
package main

import (
	"encoding/json"
	"fmt"

	"github.com/oasisprotocol/oasis-core/go/common/crypto/hash"
	"github.com/oasisprotocol/oasis-core/go/common/crypto/signature"

	"github.com/oasisprotocol/metadata-registry-tools/testcases"
	"github.com/oasisprotocol/metadata-registry-tools/testvectors"
)

func main() {
	// Configure chain context for all signatures using chain domain separation.
	var chainContext hash.Hash
	chainContext.FromBytes([]byte("metadata registry test vectors"))
	signature.SetChainContext(chainContext.String())

	var vectors []testvectors.EntityMetadataTestVector

	for _, tc := range testcases.EntityMetadataBasicVersionAndSize {
		vec := testvectors.MakeEntityMetadataTestVector(
			"EntityMetadataBasicVersionAndSize", &tc.EntityMeta, tc.Valid,
		)
		vectors = append(vectors, vec)
	}

	for _, tc := range testcases.EntityMetadataExtendedVersionAndSize {
		vec := testvectors.MakeEntityMetadataTestVector(
			"EntityMetadataExtendedVersionAndSize", &tc.EntityMeta, tc.Valid,
		)
		vectors = append(vectors, vec)
	}

	// Generate output.
	jsonOut, _ := json.MarshalIndent(&vectors, "", "  ")
	fmt.Printf("%s", jsonOut)
}
