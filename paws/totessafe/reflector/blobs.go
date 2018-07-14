package reflector

import (
	"sync"
)

// WARNING:
//
// This buffer does not have a limit, and could potentially grow uncontrollably and put the system in deadlock.
// TODO @kris-nova add a limit j

var (

	// blobs is a variable size buffer
	blobs []*PawsBlob

	// blobMutex locks the buffer for every transaction
	blobMutex = sync.Mutex{}
)

// addBlob will add a blob to the heap
func addBlob(blob *PawsBlob) {
	blobMutex.Lock()
	blobs = append(blobs, blob)
	defer blobMutex.Unlock()
}

// popBlob will pop a blob from the bottom of the heap
func popBlob() *PawsBlob {
	blobMutex.Lock()
	defer blobMutex.Unlock()
	if len(blobs) > 0 {
		blob := blobs[len(blobs)-1]
		blobs = blobs[:len(blobs)-1]
		return blob
	}
	return nil
}
