package cafs

import (
	"bytes"
	"context"
	"fmt"
	"hash/crc32"
	"io"
	"sync/atomic"

	"github.com/oneconcern/datamon/pkg/storage"

	"github.com/minio/blake2b-simd"
)

// Writer interface for a content addressable FS
type Writer interface {
	io.WriteCloser
	Flush() (Key, []byte, error)
}

type fsWriter struct {
	store               storage.Store       // CAFS backing store
	prefix              string              // Prefix for store paths
	leafSize            uint32              // Size of chunks
	leafs               []Key               // List of keys backing a file
	buf                 []byte              // Buffer stage a chunk == leafsize
	offset              int                 // till where buffer is used
	flushed             uint32              // writer has been flushed to store
	pather              func(string) string // pathing logic
	count               uint64              // total number of parallel writes
	flushChan           chan blobFlush      // channel for parallel writes
	errC                chan error          // channel for errors during parallel writes
	flushThreadDoneChan chan struct{}
	maxGoRoutines       chan struct{} // Max number of concurrent writes
	blobFlushes         []blobFlush
	errors              []error
}

func (w *fsWriter) Write(p []byte) (n int, err error) {
	written := 0
	for {
		if written == len(p) {
			return len(p), nil
		}
		// Copy p to w.buf
		writable := len(w.buf) - w.offset
		if len(p) < writable {
			writable = len(p)
		}
		c := copy(w.buf[w.offset:], p[written:writable])
		w.offset += c
		written += c
		if w.offset == len(w.buf) { // sizes line up, flush and continue
			w.count++ // next leaf
			w.maxGoRoutines <- struct{}{}
			go pFlush(
				false,
				w.buf,
				w.prefix,
				w.leafSize,
				w.count,
				w.flushChan,
				w.errC,
				w.maxGoRoutines,
				w.pather,
				w.store,
			)
			w.buf = make([]byte, w.leafSize) // new buffer
			w.offset = 0                     // new offset for new buffer
			continue
		}
	}
}

type blobFlush struct {
	count uint64
	key   Key
}

func pFlush(
	isLastNode bool,
	buffer []byte,
	prefix string,
	leafSize uint32,
	count uint64,
	flushChan chan blobFlush,
	errC chan error,
	maxGoRoutines chan struct{},
	pather func(string) string,
	destination storage.Store,
) {
	defer func() {
		<-maxGoRoutines
	}()
	// Calculate hash value
	hasher, err := blake2b.New(&blake2b.Config{
		Size: blake2b.Size,
		Tree: &blake2b.Tree{
			Fanout:        0,
			MaxDepth:      2,
			LeafSize:      leafSize,
			NodeOffset:    count,
			NodeDepth:     0,
			InnerHashSize: blake2b.Size,
			IsLastNode:    isLastNode,
		},
	})
	if err != nil {
		errC <- err
		return
	}
	_, err = hasher.Write(buffer)
	if err != nil {
		errC <- fmt.Errorf("flush segment hash: %v", err)
		return
	}

	leafKey, err := NewKey(hasher.Sum(nil))
	if err != nil {
		errC <- fmt.Errorf("flush key segment: %v", err)
		return
	}

	// Write the blob
	if pather == nil {
		// w.pather = func(lks string) string { return filepath.Join(lks[:3], lks[3:6], lks[6:]) }
		pather = func(lks string) string { return prefix + lks }
	}
	found, _ := destination.Has(context.TODO(), pather(leafKey.String()))
	if !found {
		d, ok := destination.(storage.StoreCRC)
		if ok {
			crc := crc32.Checksum(buffer, crc32.MakeTable(crc32.Castagnoli))
			err = d.PutCRC(context.TODO(), pather(leafKey.String()), bytes.NewReader(buffer), storage.OverWrite, crc)
		} else {
			err = destination.Put(context.TODO(), pather(leafKey.String()), bytes.NewReader(buffer), storage.OverWrite)
		}
		if err != nil {
			errC <- fmt.Errorf("write segment file: %s, err: %w", pather(leafKey.String()), err)
			return
		}
		fmt.Printf("Uploading blob:%s\n", leafKey.String())
	} else {
		fmt.Printf("Duplicate blob:%s\n", leafKey.String())
	}
	flushChan <- blobFlush{
		count: count,
		key:   leafKey,
	}
}

func (w *fsWriter) flush(isLastNode bool) (int, error) {
	if w.offset == 0 {
		return 0, nil
	}
	hasher, err := blake2b.New(&blake2b.Config{
		Size: blake2b.Size,
		Tree: &blake2b.Tree{
			Fanout:        0,
			MaxDepth:      2,
			LeafSize:      w.leafSize,
			NodeOffset:    uint64(len(w.leafs)),
			NodeDepth:     0,
			InnerHashSize: blake2b.Size,
			IsLastNode:    isLastNode,
		},
	})
	if err != nil {
		return 0, err
	}

	_, err = hasher.Write(w.buf[:w.offset])
	if err != nil {
		return 0, fmt.Errorf("flush segment hash: %v", err)
	}

	leafKey, err := NewKey(hasher.Sum(nil))
	if err != nil {
		return 0, fmt.Errorf("flush key segment: %v", err)
	}

	if w.pather == nil {
		// w.pather = func(lks string) string { return filepath.Join(lks[:3], lks[3:6], lks[6:]) }
		w.pather = func(lks string) string { return w.prefix + lks }
	}
	found, _ := w.store.Has(context.TODO(), w.pather(leafKey.String()))
	if !found {
		d, ok := w.store.(storage.StoreCRC)
		if ok {
			crc := crc32.Checksum(w.buf[:w.offset], crc32.MakeTable(crc32.Castagnoli))
			err = d.PutCRC(context.TODO(), w.pather(leafKey.String()), bytes.NewReader(w.buf[:w.offset]), storage.OverWrite, crc)
		} else {
			err = w.store.Put(context.TODO(), w.pather(leafKey.String()), bytes.NewReader(w.buf[:w.offset]), storage.OverWrite)
		}
		if err != nil {
			return 0, fmt.Errorf("write segment file: %s err:%w", w.pather(leafKey.String()), err)
		}
		fmt.Printf("Uploading blob:%s, bytes:%d\n", leafKey.String(), w.offset)
	} else {
		fmt.Printf("Duplicate blob:%s, bytes:%d\n", leafKey.String(), w.offset)
	}

	n := w.offset
	w.offset = 0
	w.leafs = append(w.leafs, leafKey)
	return n, nil
}

func (w *fsWriter) flushThread() {
	var err error
	var bf blobFlush
	notDone := true
	for notDone {
		select {
		case bf, notDone = <-w.flushChan:
			if notDone {
				w.blobFlushes = append(w.blobFlushes, bf)
			}
		case err = <-w.errC:
			w.errors = append(w.errors, err)
		}
	}
	w.flushThreadDoneChan <- struct{}{}
}

// don't Write() during Flush()
func (w *fsWriter) Flush() (Key, []byte, error) {
	for i := 0; i < cap(w.maxGoRoutines); i++ {
		w.maxGoRoutines <- struct{}{}
	}
	close(w.flushChan)
	<-w.flushThreadDoneChan
	if len(w.errors) != 0 {
		return Key{}, nil, w.errors[0]
	}
	w.leafs = make([]Key, len(w.blobFlushes))
	for _, bf := range w.blobFlushes {
		w.leafs[bf.count-1] = bf.key
	}
	atomic.StoreUint32(&w.flushed, 1)

	_, err := w.flush(true)
	if err != nil {
		return Key{}, nil, err
	}

	rhash, err := RootHash(w.leafs, w.leafSize)
	if err != nil {
		return Key{}, nil, fmt.Errorf("flush make root hash: %v", err)
	}

	leafHashes := make([]byte, len(w.leafs)*KeySize)
	for i, leaf := range w.leafs {
		offset := KeySize * i
		copy(leafHashes[offset:offset+KeySize], leaf[:])
	}
	return rhash, leafHashes, nil
}

func (w *fsWriter) Close() error {
	if !atomic.CompareAndSwapUint32(&w.flushed, 1, 0) {
		return fmt.Errorf("stream closed without being flushed")
	}
	return nil
}
