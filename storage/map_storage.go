package storage

import (
	"github.com/google/trillian"
)

// ReadOnlyMapTX provides a read-only view into the Map data.
type ReadOnlyMapTX interface {
	ReadOnlyTreeTX
	MapRootReader
	Getter
}

// MapTX is the transactional interface for reading/modifying a Map.
// It extends the basic TreeTX interface with Map specific methods.
type MapTX interface {
	TreeTX
	MapRootReader
	MapRootWriter
	Getter
	Setter
}

// ReadOnlyMapStorage provides a narrow read-only view into a MapStorage.
type ReadOnlyMapStorage interface {
	// Snapshot starts a new read-only transaction.
	// Commit must be called when the caller is finished with the returned object,
	// and values read through it should only be propagated if Commit returns
	// without error.
	Snapshot() (ReadOnlyMapTX, error)
}

// MapStorage should be implemented by concrete storage mechanisms which want to support Maps
type MapStorage interface {
	ReadOnlyMapStorage
	// Begin starts a new Map transaction.
	// Either Commit or Rollback must be called when the caller is finished with
	// the returned object, and values read through it should only be propagated
	// if Commit returns without error.
	Begin() (MapTX, error)
}

// Setter allows the setting of key->value pairs on the map.
type Setter interface {
	// Set sets key to leaf
	Set(key []byte, value trillian.MapLeaf) error
}

// Getter allows access to the values stored in the map.
type Getter interface {
	// Get retrieves the value associates with key, if any, at the specified revision.
	// Setting revision to -1 will fetch the latest revision.
	Get(revision int64, key []byte) (trillian.MapLeaf, error)
}

// MapRootReader provides access to the map roots.
type MapRootReader interface {
	// LatestSignedMapRoot returns the most recently created SignedMapRoot.
	LatestSignedMapRoot() (trillian.SignedMapRoot, error)
}

// MapRootWriter allows the storage of new SignedMapRoots
type MapRootWriter interface {
	// StoreSignedMapRoot stores root.
	StoreSignedMapRoot(root trillian.SignedMapRoot) error
}