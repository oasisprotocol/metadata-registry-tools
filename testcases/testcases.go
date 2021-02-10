package testcases

import (
	"math"
	"strconv"

	"github.com/oasisprotocol/oasis-core/go/common/cbor"

	registry "github.com/oasisprotocol/metadata-registry-tools"
)

const (
	EntityValidName      = "this is a name"
	EntityTooLongName    = "this is a name but it is soooooooooooooooooooo long"
	EntityValidURL       = "https://hello.world/bar/goo"
	EntityTooLongURL     = "https://too.too.too.too.too.too.too.too.too.too.too.too.too.too.long"
	EntityValidEmail     = "hello@world.org"
	EntityTooLongEmail   = "too@too.too.too.too.too.too.too.long"
	EntityValidKeybase   = "Hello_world42"
	EntityTooLongKeybase = "tootootootootootootootootootoolong"
	EntityValidTwitter   = "Hello_world42"
	EntityTooLongTwitter = "tootootootootootootootootootoolong"
)

// EntityMetadataTestCase is an entity metadata test case.
type EntityMetadataTestCase struct {
	Name       string
	EntityMeta registry.EntityMetadata
	Valid      bool
}

var (
	// EntityMetadataBasicVersionAndSize are the entity metadata test cases that
	// contain test cases for basic version and field sizes checks.
	EntityMetadataBasicVersionAndSize []EntityMetadataTestCase = []EntityMetadataTestCase{
		{"InvalidVersion1", registry.EntityMetadata{Versioned: cbor.NewVersioned(0)}, false},
		{"InvalidVersion2", registry.EntityMetadata{Versioned: cbor.NewVersioned(2)}, false},
		{"ValidName", registry.EntityMetadata{Versioned: cbor.NewVersioned(1), Name: EntityValidName}, true},
		{"TooLongName", registry.EntityMetadata{Versioned: cbor.NewVersioned(1), Name: EntityTooLongName}, false},
		{"ValidURL", registry.EntityMetadata{Versioned: cbor.NewVersioned(1), URL: EntityValidURL}, true},
		{"TooLongURL", registry.EntityMetadata{Versioned: cbor.NewVersioned(1), URL: EntityTooLongURL}, false},
		{"ValidEmail", registry.EntityMetadata{Versioned: cbor.NewVersioned(1), Email: EntityValidEmail}, true},
		{"TooLongEmail", registry.EntityMetadata{Versioned: cbor.NewVersioned(1), Email: EntityTooLongEmail}, false},
		{"ValidKeybase", registry.EntityMetadata{Versioned: cbor.NewVersioned(1), Keybase: EntityValidKeybase}, true},
		{"TooLongKeybase", registry.EntityMetadata{Versioned: cbor.NewVersioned(1), Keybase: EntityTooLongKeybase}, false},
		{"ValidTwitter", registry.EntityMetadata{Versioned: cbor.NewVersioned(1), Twitter: EntityValidTwitter}, true},
		{"TooLongTwitter", registry.EntityMetadata{Versioned: cbor.NewVersioned(1), Twitter: EntityTooLongTwitter}, false},
	}

	// EntityMetadataExtendedVersionAndSize are the entity metadata test cases
	// that contain test cases for extended version and field sizes checks.
	// NOTE: All these test cases contain full entity metadata structs (i.e. no
	// fields are empty).
	EntityMetadataExtendedVersionAndSize []EntityMetadataTestCase

	// EntityMetadataFieldSemantics are the entity metadata test cases that
	// contain test cases for checking fields' semantics.
	EntityMetadataFieldSemantics []EntityMetadataTestCase = []EntityMetadataTestCase{
		{"ValidURL", registry.EntityMetadata{Versioned: cbor.NewVersioned(1), URL: EntityValidURL}, true},
		{"BadSchemeURL", registry.EntityMetadata{Versioned: cbor.NewVersioned(1), URL: "http://hello.world/bar/goo"}, false},
		{"BadQueryURL", registry.EntityMetadata{Versioned: cbor.NewVersioned(1), URL: "https://hello.world/bar?goo=1"}, false},
		{"BadFragmentURL", registry.EntityMetadata{Versioned: cbor.NewVersioned(1), URL: "https://hello.world/bar#goo"}, false},
		{"BadPortURL", registry.EntityMetadata{Versioned: cbor.NewVersioned(1), URL: "https://hello.world:123/bar"}, false},
		{"BadURL1", registry.EntityMetadata{Versioned: cbor.NewVersioned(1), URL: "hello.world"}, false},
		{"BadURL2", registry.EntityMetadata{Versioned: cbor.NewVersioned(1), URL: "127.0.0.1:1234"}, false},
		{"ValidEmail", registry.EntityMetadata{Versioned: cbor.NewVersioned(1), Email: EntityValidEmail}, true},
		{"BadEmail1", registry.EntityMetadata{Versioned: cbor.NewVersioned(1), Email: "hello world.org"}, false},
		{"BadEmail2", registry.EntityMetadata{Versioned: cbor.NewVersioned(1), Email: "Hello World <hello@world.org>"}, false},
		{"BadEmail3", registry.EntityMetadata{Versioned: cbor.NewVersioned(1), Email: "@world.org"}, false},
		{"BadEmail4", registry.EntityMetadata{Versioned: cbor.NewVersioned(1), Email: "hello@.org"}, false},
		{"ValidKeybase", registry.EntityMetadata{Versioned: cbor.NewVersioned(1), Keybase: EntityValidKeybase}, true},
		{"BadKeybase1", registry.EntityMetadata{Versioned: cbor.NewVersioned(1), Keybase: "helloworld-"}, false},
		{"BadKeybase2", registry.EntityMetadata{Versioned: cbor.NewVersioned(1), Keybase: "https://keybase.io/hello"}, false},
		{"BadKeybase3", registry.EntityMetadata{Versioned: cbor.NewVersioned(1), Keybase: "foo-bar"}, false},
		{"BadKeybase4", registry.EntityMetadata{Versioned: cbor.NewVersioned(1), Keybase: "foo:bar"}, false},
		{"ValidTwitter", registry.EntityMetadata{Versioned: cbor.NewVersioned(1), Twitter: EntityValidTwitter}, true},
		{"BadTwitter1", registry.EntityMetadata{Versioned: cbor.NewVersioned(1), Twitter: "helloworld-"}, false},
		{"BadTwitter2", registry.EntityMetadata{Versioned: cbor.NewVersioned(1), Twitter: "https://twitter.com/hello"}, false},
		{"BadTwitter3", registry.EntityMetadata{Versioned: cbor.NewVersioned(1), Twitter: "foo-bar"}, false},
		{"BadTwitter4", registry.EntityMetadata{Versioned: cbor.NewVersioned(1), Twitter: "foo:bar"}, false},
	}
)

// validBounds returns true iff all the given entity metadata fields are within
// each field's valid bounds.
func validBounds(version uint16, name, url, email, keybase, twitter string) bool {
	if version < registry.MinSupportedVersion || version > registry.MaxSupportedVersion ||
		len(name) > registry.MaxEntityNameLength ||
		len(url) > registry.MaxEntityURLLength ||
		len(email) > registry.MaxEntityEmailLength ||
		len(keybase) > registry.MaxEntityKeybaseLength ||
		len(twitter) > registry.MaxEntityTwitterLength {
		return false
	}
	return true
}

func init() {
	// Generate test cases for entity metadata by permutating through all field
	// value lists below.
	versions := []uint16{0, 1, 2}
	serials := []uint64{0, 1, 10, 42, 1000, 1_000_000, 10_000_000, math.MaxUint64}
	names := []string{EntityValidName, EntityTooLongName}
	urls := []string{EntityValidURL, EntityTooLongURL}
	emails := []string{EntityValidEmail, EntityTooLongEmail}
	keybaseHandles := []string{EntityValidKeybase, EntityTooLongKeybase}
	twitterHandles := []string{EntityValidTwitter, EntityTooLongTwitter}

	count := 0
	EntityMetadataExtendedVersionAndSize = []EntityMetadataTestCase{}
	for _, v := range versions {
		for _, s := range serials {
			for _, name := range names {
				for _, url := range urls {
					for _, email := range emails {
						for _, keybase := range keybaseHandles {
							for _, twitter := range twitterHandles {
								meta := registry.EntityMetadata{
									Versioned: cbor.Versioned{V: v},
									Serial:    s,
									Name:      name,
									URL:       url,
									Email:     email,
									Keybase:   keybase,
									Twitter:   twitter,
								}
								tc := EntityMetadataTestCase{
									Name:       "ExtendedVersionAndSizeChecks: " + strconv.Itoa(count),
									EntityMeta: meta,
									Valid:      validBounds(v, name, url, email, keybase, twitter),
								}
								EntityMetadataExtendedVersionAndSize = append(
									EntityMetadataExtendedVersionAndSize, tc,
								)
								count++
							}
						}
					}
				}
			}
		}
	}
}
