package testcases

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func entityMetadataValidateBasic(require *require.Assertions, tc EntityMetadataTestCase) {
	err := tc.EntityMeta.ValidateBasic()
	switch tc.Valid {
	case true:
		require.NoError(err, "ValidateBasic should not fail on %s", tc.Name)
	case false:
		require.Error(err, "ValidateBasic should fail on %s", tc.Name)
	}
}

func TestEntityMetadata(t *testing.T) {
	require := require.New(t)

	for _, tc := range EntityMetadataBasicVersionAndSize {
		entityMetadataValidateBasic(require, tc)
	}

	for _, tc := range EntityMetadataExtendedVersionAndSize {
		entityMetadataValidateBasic(require, tc)
	}

	for _, tc := range EntityMetadataFieldSemantics {
		entityMetadataValidateBasic(require, tc)
	}
}
