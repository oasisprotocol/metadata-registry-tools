// Package registry provides an interface to the Oasis off-chain registry of signed statements.
package registry

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/mail"
	"net/url"
	"regexp"

	"github.com/oasisprotocol/oasis-core/go/common/cbor"
	"github.com/oasisprotocol/oasis-core/go/common/crypto/signature"
	"github.com/oasisprotocol/oasis-core/go/common/prettyprint"
)

var (
	// ErrNoSuchEntity is the error returned where the requested entity cannot be found.
	ErrNoSuchEntity = errors.New("registry: no such entity")

	// ErrCorruptedRegistry is the error returned where the registry is corrupted (does not conform
	// to the specifications or contains data that fails signature verification).
	ErrCorruptedRegistry = errors.New("registry: corrupted registry")
)

const (
	// MaxStatementSize is the maximum encoded signed statement size in bytes.
	MaxStatementSize = 16 * 1024

	// MaxEntityNameLength is the maximum length of the entity metadata's Name field.
	MaxEntityNameLength = 50
	// MaxEntityURLLength is the maximum length of the entity metadata's URL field.
	MaxEntityURLLength = 64
	// MaxEntityEmailLength is the maximum length of the entity metadata's Email field.
	MaxEntityEmailLength = 32
	// MaxEntityKeybaseLength is the maximum length of the entity metadata's Keybase field.
	MaxEntityKeybaseLength = 32
	// MaxEntityTwitterLength is the maximum length of the entity metadata's Twitter field.
	MaxEntityTwitterLength = 32

	// MinSupportedVersion is the minimum supported entity metadata version.
	MinSupportedVersion = 1
	// MaxSupportedVersion is the maximum supported entity metadata version.
	MaxSupportedVersion = 1
)

var (
	// TwitterHandleRegexp is the regular expression used for validating the Twitter field.
	TwitterHandleRegexp = regexp.MustCompile(`^[A-Za-z0-9_]+$`)
	// KeybaseHandleRegexp is the regular expression used for validating the Keybase field.
	KeybaseHandleRegexp = regexp.MustCompile(`^[A-Za-z0-9_]+$`)
)

// Provider is the read-only registry provider interface.
type Provider interface {
	// Verify verifies the integrity of the whole registry.
	Verify() error

	// VerifyUpdate verifies the integrity of a registry update from src.
	VerifyUpdate(src Provider) error

	// GetEntities returns a list of all entities in the registry.
	GetEntities(ctx context.Context) (map[signature.PublicKey]*EntityMetadata, error)

	// GetEntity returns metadata for a specific entity.
	GetEntity(ctx context.Context, id signature.PublicKey) (*EntityMetadata, error)
}

// EntityMetadataSignatureContext is the domain separation context used for entity metadata.
var EntityMetadataSignatureContext = signature.NewContext("oasis-metadata-registry: entity")

var _ prettyprint.PrettyPrinter = (*EntityMetadata)(nil)

// EntityMetadata contains metadata about an entity.
type EntityMetadata struct {
	cbor.Versioned

	// Serial is the serial number of the entity metadata statement.
	Serial uint64 `json:"serial"`

	// Name is the entity name.
	Name string `json:"name,omitempty"`

	// URL is an URL associated with an entity.
	URL string `json:"url,omitempty"`

	// Email is the entity's contact e-mail address.
	Email string `json:"email,omitempty"`

	// Keybase is the keybase.io handle.
	Keybase string `json:"keybase,omitempty"`

	// Twitter is the Twitter handle.
	Twitter string `json:"twitter,omitempty"`
}

// Equal compares vs another entity metadata for equality.
func (e *EntityMetadata) Equal(other *EntityMetadata) bool {
	return bytes.Equal(cbor.Marshal(e), cbor.Marshal(other))
}

// ValidateBasic performs basic validity checks on the entity metadata.
func (e *EntityMetadata) ValidateBasic() error {
	if e.Versioned.V < MinSupportedVersion || e.Versioned.V > MaxSupportedVersion {
		return fmt.Errorf("unsupported entity metadata version: %d", e.Versioned.V)
	}

	// Name.
	if len(e.Name) > MaxEntityNameLength {
		return fmt.Errorf("entity name too long (length: %d max: %d)", len(e.Name), MaxEntityNameLength)
	}

	// URL.
	if len(e.URL) > MaxEntityURLLength {
		return fmt.Errorf("entity URL too long (length: %d max: %d)", len(e.URL), MaxEntityURLLength)
	}
	if len(e.URL) > 0 {
		parsedURL, err := url.Parse(e.URL)
		if err != nil {
			return fmt.Errorf("entity URL is malformed: %w", err)
		}
		if parsedURL.Scheme != "https" {
			return fmt.Errorf("entity URL must use the https scheme (scheme: %s)", parsedURL.Scheme)
		}
		if port := parsedURL.Port(); port != "" {
			return fmt.Errorf("entity URL must use the default port (port: %s)", port)
		}
		if len(parsedURL.RawQuery) != 0 || len(parsedURL.Fragment) != 0 {
			return fmt.Errorf("entity URL must not contain query values or fragments")
		}
	}

	// Email.
	if len(e.Email) > MaxEntityEmailLength {
		return fmt.Errorf("entity e-mail too long (length: %d max: %d)", len(e.Email), MaxEntityEmailLength)
	}
	if len(e.Email) > 0 {
		parsedEmail, err := mail.ParseAddress(e.Email)
		if err != nil {
			return fmt.Errorf("entity e-mail is malformed: %w", err)
		}
		if len(parsedEmail.Name) != 0 {
			return fmt.Errorf("entity e-mail must not contain a name")
		}
	}

	// Keybase.
	if len(e.Keybase) > MaxEntityKeybaseLength {
		return fmt.Errorf("entity keybase handle too long (length: %d max: %d)", len(e.Keybase), MaxEntityKeybaseLength)
	}
	if len(e.Keybase) > 0 {
		if !KeybaseHandleRegexp.MatchString(e.Keybase) {
			return fmt.Errorf("entity keybase handle is malformed")
		}
	}

	// Twitter.
	if len(e.Twitter) > MaxEntityTwitterLength {
		return fmt.Errorf("entity twitter handle too long (length: %d max: %d)", len(e.Twitter), MaxEntityTwitterLength)
	}
	if len(e.Twitter) > 0 {
		if !TwitterHandleRegexp.MatchString(e.Twitter) {
			return fmt.Errorf("entity twitter handle is malformed")
		}
	}

	return nil
}

// Load loads and verifies entity metadata from a given reader containing signed entity metadata.
func (e *EntityMetadata) Load(id signature.PublicKey, r io.Reader) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return fmt.Errorf("%w: failed to read metadata: %s", ErrCorruptedRegistry, err)
	}
	if len(b) > MaxStatementSize {
		return fmt.Errorf("%w: statement too big (size: %d max: %d)", ErrCorruptedRegistry, len(b), MaxStatementSize)
	}

	var sigEntity SignedEntityMetadata
	if err = json.Unmarshal(b, &sigEntity); err != nil {
		return fmt.Errorf("%w: failed to unmarshal signed entity metadata: %s", ErrCorruptedRegistry, err)
	}
	if !sigEntity.Signature.PublicKey.Equal(id) {
		return fmt.Errorf("%w: entity metadata signer does not match expected entity (expected: %s got: %s)",
			ErrCorruptedRegistry,
			id,
			sigEntity.Signature.PublicKey,
		)
	}

	if err = sigEntity.Open(e); err != nil {
		return fmt.Errorf("%w: failed to verify signed entity metadata: %s", ErrCorruptedRegistry, err)
	}
	if err = e.ValidateBasic(); err != nil {
		return fmt.Errorf("%w: failed to validate entity metadata: %s", ErrCorruptedRegistry, err)
	}
	return nil
}

// PrettyPrint writes a pretty-printed representation of EntityMetadata to the
// given writer.
func (e *EntityMetadata) PrettyPrint(ctx context.Context, prefix string, w io.Writer) {
	fmt.Fprintf(w, "%sVersion: %d\n", prefix, e.V)
	fmt.Fprintf(w, "%sSerial:  %d\n", prefix, e.Serial)
	fmt.Fprintf(w, "%sName:    %s\n", prefix, e.Name)
	fmt.Fprintf(w, "%sURL:     %s\n", prefix, e.URL)
	fmt.Fprintf(w, "%sEmail:   %s\n", prefix, e.Email)
	fmt.Fprintf(w, "%sKeybase: %s\n", prefix, e.Keybase)
	fmt.Fprintf(w, "%sTwitter: %s\n", prefix, e.Twitter)
}

// PrettyType returns a representation of EntityMetadata that can be used for
// pretty printing.
func (e EntityMetadata) PrettyType() (interface{}, error) {
	return e, nil
}

// SignedEntityMetadata is a signed entity metadata statement.
type SignedEntityMetadata struct {
	signature.Signed
}

// Open first verifies the blob signature and then unmarshals the blob.
func (s *SignedEntityMetadata) Open(meta *EntityMetadata) error { // nolint: interfacer
	return s.Signed.Open(EntityMetadataSignatureContext, meta)
}

// Save serializes and writes entity metadata to the given writer.
func (s *SignedEntityMetadata) Save(w io.Writer) error {
	b, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if _, err = w.Write(b); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}
	return nil
}

// SignEntityMetadata serializes the EntityMetadata and signs the result.
func SignEntityMetadata(signer signature.Signer, meta *EntityMetadata) (*SignedEntityMetadata, error) {
	signed, err := signature.SignSigned(signer, EntityMetadataSignatureContext, meta)
	if err != nil {
		return nil, err
	}

	return &SignedEntityMetadata{
		Signed: *signed,
	}, nil
}
