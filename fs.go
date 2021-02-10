package registry

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/oasisprotocol/oasis-core/go/common/crypto/signature"
)

const (
	registryDir       = "registry"
	registryEntityDir = "entity"

	placeholderFilename = ".placeholder"
	statementExt        = ".json"
)

// MutableProvider is a mutable registry provider interface.
type MutableProvider interface {
	Provider

	// BaseDir returns the base registry directory (when available).
	BaseDir() string

	// Init initializes a new registry in the local filesystem.
	Init() error

	// UpdateEntity updates entity metadata in the registry.
	UpdateEntity(entity *SignedEntityMetadata) error
}

type fsProvider struct {
	baseDir string
	fs      billy.Filesystem
}

// Implements Provider.
func (p *fsProvider) Verify() error {
	_, err := p.GetEntities(context.Background())
	return err
}

// Implements Provider.
func (p *fsProvider) VerifyUpdate(src Provider) error {
	ctx := context.Background()
	dstEnts, err := p.GetEntities(ctx)
	if err != nil {
		return fmt.Errorf("destination registry is corrupted: %w", err)
	}

	srcEnts, err := src.GetEntities(ctx)
	if err != nil {
		return fmt.Errorf("source registry is corrupted: %w", err)
	}

	// No entites can be removed by an update.
	for id := range srcEnts {
		if dstEnts[id] == nil {
			return fmt.Errorf("entity statement has been removed: %s", id)
		}
	}

	// Updated entities must use a higher serial number.
	for id, dst := range dstEnts {
		var src *EntityMetadata
		if src = srcEnts[id]; src == nil {
			// New entity.
			continue
		}

		if !src.Equal(dst) && dst.Serial <= src.Serial {
			return fmt.Errorf("updated entity '%s' metadata must increase serial number (existing: %d provided: %d)",
				id,
				src.Serial,
				dst.Serial,
			)
		}
	}

	return nil
}

// Implements Provider.
func (p *fsProvider) GetEntities(ctx context.Context) (map[signature.PublicKey]*EntityMetadata, error) {
	entities, err := p.fs.ReadDir(p.fs.Join(registryDir, registryEntityDir))
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read entity directory: %s", ErrCorruptedRegistry, err)
	}

	results := make(map[signature.PublicKey]*EntityMetadata)
	for _, fi := range entities {
		if filepath.Ext(fi.Name()) != statementExt {
			continue
		}

		var id signature.PublicKey
		if err = id.UnmarshalHex(strings.TrimSuffix(fi.Name(), statementExt)); err != nil {
			return nil, fmt.Errorf("%w: entity: bad statement filename '%s': %s", ErrCorruptedRegistry, fi.Name(), err)
		}

		if fi.Size() > MaxStatementSize {
			return nil, fmt.Errorf(
				"%w: entity: statement too big (size: %d max: %d): %s",
				ErrCorruptedRegistry, fi.Size(), MaxStatementSize, fi.Name(),
			)
		}

		var result *EntityMetadata
		if result, err = p.GetEntity(context.Background(), id); err != nil {
			return nil, fmt.Errorf("%w: entity: bad statement '%s': %s", ErrCorruptedRegistry, fi.Name(), err)
		}

		results[id] = result
	}
	return results, nil
}

func (p *fsProvider) getEntityPath(id signature.PublicKey) string {
	entityID := publicKeyToFilename(id)
	return p.fs.Join(registryDir, registryEntityDir, entityID+statementExt)
}

// Implements Provider.
func (p *fsProvider) GetEntity(ctx context.Context, id signature.PublicKey) (*EntityMetadata, error) {
	f, err := p.fs.Open(p.getEntityPath(id))
	switch {
	case err == nil:
	case os.IsNotExist(err):
		return nil, ErrNoSuchEntity
	default:
		return nil, fmt.Errorf("%w: failed to open entity metadata: %s", ErrCorruptedRegistry, err)
	}
	defer f.Close()

	entity := new(EntityMetadata)
	return entity, entity.Load(id, f)
}

// Implements MutableProvider.
func (p *fsProvider) BaseDir() string {
	return p.baseDir
}

// Implements MutableProvider.
func (p *fsProvider) Init() error {
	paths := []string{
		registryDir,
		p.fs.Join(registryDir, registryEntityDir),
	}

	for _, path := range paths {
		if _, err := p.fs.Stat(path); !os.IsNotExist(err) {
			return fmt.Errorf("registry already initialized (or corrupted)")
		}

		if err := p.fs.MkdirAll(path, 0o755); err != nil {
			return fmt.Errorf("failed to create path %s: %w", path, err)
		}

		placeholderPath := p.fs.Join(path, placeholderFilename)
		f, err := p.fs.OpenFile(placeholderPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
		if err != nil {
			return fmt.Errorf("failed to create placeholder %s: %w", placeholderPath, err)
		}
		f.Close()
	}

	return nil
}

// Implements MutableProvider.
func (p *fsProvider) UpdateEntity(entity *SignedEntityMetadata) error {
	// Make sure the signed entity is valid before processing it.
	var inner EntityMetadata
	if err := entity.Open(&inner); err != nil {
		return fmt.Errorf("bad signed entity metadata: %w", err)
	}
	if err := inner.ValidateBasic(); err != nil {
		return fmt.Errorf("bad signed entity metadata: %w", err)
	}

	// Check if the entity already exists. In this case, require that the serial number is bumped.
	existing, err := p.GetEntity(context.Background(), entity.Signature.PublicKey)
	switch err {
	case nil:
		if inner.Serial <= existing.Serial {
			return fmt.Errorf("updated entity metadata must increase serial number (existing: %d provided: %d)",
				existing.Serial,
				inner.Serial,
			)
		}
	case ErrNoSuchEntity:
	default:
		return fmt.Errorf("failed to query for existing entity: %w", err)
	}

	f, err := p.fs.Create(p.getEntityPath(entity.Signature.PublicKey))
	if err != nil {
		return fmt.Errorf("failed to create entity metadata file: %w", err)
	}
	defer f.Close()

	return entity.Save(f)
}

// NewFilesystemProvider creates a new filesystem-based registry interface.
func NewFilesystemProvider(fs billy.Filesystem) (MutableProvider, error) {
	return &fsProvider{fs: fs}, nil
}

// NewFilesystemPathProvider creates a new filesystem-based registry interface for the given path.
func NewFilesystemPathProvider(path string) (MutableProvider, error) {
	return &fsProvider{
		baseDir: path,
		fs:      osfs.New(path),
	}, nil
}

func publicKeyToFilename(pk signature.PublicKey) string {
	rawPk, err := pk.MarshalBinary()
	if err != nil {
		panic(fmt.Errorf("registry: failed to marshal public key: %w", err))
	}

	return hex.EncodeToString(rawPk)
}
