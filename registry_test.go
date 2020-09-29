package registry

import (
	"testing"

	"github.com/oasisprotocol/oasis-core/go/common/cbor"
	"github.com/stretchr/testify/require"
)

func TestEntityMetadata(t *testing.T) {
	require := require.New(t)

	for _, tc := range []struct {
		name  string
		meta  EntityMetadata
		valid bool
	}{
		{"InvalidVersion1", EntityMetadata{Versioned: cbor.NewVersioned(0)}, false},
		{"InvalidVersion2", EntityMetadata{Versioned: cbor.NewVersioned(2)}, false},
		{"ValidName", EntityMetadata{Versioned: cbor.NewVersioned(1), Name: "this is a name"}, true},
		{"TooLongName", EntityMetadata{Versioned: cbor.NewVersioned(1), Name: "this is a name but it is soooooooooooooooooooo long"}, false},
		{"ValidURL", EntityMetadata{Versioned: cbor.NewVersioned(1), URL: "https://hello.world/bar/goo"}, true},
		{"TooLongURL", EntityMetadata{Versioned: cbor.NewVersioned(1), URL: "https://too.too.too.too.too.too.too.too.too.too.too.too.too.too.long"}, false},
		{"BadSchemeURL", EntityMetadata{Versioned: cbor.NewVersioned(1), URL: "http://hello.world/bar/goo"}, false},
		{"BadQueryURL", EntityMetadata{Versioned: cbor.NewVersioned(1), URL: "https://hello.world/bar?goo=1"}, false},
		{"BadFragmentURL", EntityMetadata{Versioned: cbor.NewVersioned(1), URL: "https://hello.world/bar#goo"}, false},
		{"BadPortURL", EntityMetadata{Versioned: cbor.NewVersioned(1), URL: "https://hello.world:123/bar"}, false},
		{"BadURL1", EntityMetadata{Versioned: cbor.NewVersioned(1), URL: "hello.world"}, false},
		{"BadURL2", EntityMetadata{Versioned: cbor.NewVersioned(1), URL: "127.0.0.1:1234"}, false},
		{"ValidEmail", EntityMetadata{Versioned: cbor.NewVersioned(1), Email: "hello@world.org"}, true},
		{"TooLongEmail", EntityMetadata{Versioned: cbor.NewVersioned(1), Email: "too@too.too.too.too.too.too.too.long"}, false},
		{"BadEmail1", EntityMetadata{Versioned: cbor.NewVersioned(1), Email: "hello world.org"}, false},
		{"BadEmail2", EntityMetadata{Versioned: cbor.NewVersioned(1), Email: "Hello World <hello@world.org>"}, false},
		{"BadEmail3", EntityMetadata{Versioned: cbor.NewVersioned(1), Email: "@world.org"}, false},
		{"BadEmail4", EntityMetadata{Versioned: cbor.NewVersioned(1), Email: "hello@.org"}, false},
		{"ValidKeybase", EntityMetadata{Versioned: cbor.NewVersioned(1), Keybase: "Hello_world42"}, true},
		{"TooLongKeybase", EntityMetadata{Versioned: cbor.NewVersioned(1), Keybase: "tootootootootootootootootootoolong"}, false},
		{"BadKeybase1", EntityMetadata{Versioned: cbor.NewVersioned(1), Keybase: "helloworld-"}, false},
		{"BadKeybase2", EntityMetadata{Versioned: cbor.NewVersioned(1), Keybase: "https://keybase.io/hello"}, false},
		{"BadKeybase3", EntityMetadata{Versioned: cbor.NewVersioned(1), Keybase: "foo-bar"}, false},
		{"BadKeybase4", EntityMetadata{Versioned: cbor.NewVersioned(1), Keybase: "foo:bar"}, false},
		{"ValidTwitter", EntityMetadata{Versioned: cbor.NewVersioned(1), Twitter: "Hello_world42"}, true},
		{"TooLongTwitter", EntityMetadata{Versioned: cbor.NewVersioned(1), Twitter: "tootootootootootootootootootoolong"}, false},
		{"BadTwitter1", EntityMetadata{Versioned: cbor.NewVersioned(1), Twitter: "helloworld-"}, false},
		{"BadTwitter2", EntityMetadata{Versioned: cbor.NewVersioned(1), Twitter: "https://twitter.com/hello"}, false},
		{"BadTwitter3", EntityMetadata{Versioned: cbor.NewVersioned(1), Twitter: "foo-bar"}, false},
		{"BadTwitter4", EntityMetadata{Versioned: cbor.NewVersioned(1), Twitter: "foo:bar"}, false},
	} {
		err := tc.meta.ValidateBasic()
		switch tc.valid {
		case true:
			require.NoError(err, "ValidateBasic should not fail on %s", tc.name)
		case false:
			require.Error(err, "ValidateBasic should fail on %s", tc.name)
		}
	}
}
