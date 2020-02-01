package gbcache

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/allegro/bigcache/v2"
)

// storage in a wrapper around *bigcache.BigCache to allow to store cachedResponse
type storage struct {
	s *bigcache.BigCache
}

func newStorage(cfg bigcache.Config) (*storage, error) {
	s, err := bigcache.NewBigCache(cfg)
	if err != nil {
		return nil, fmt.Errorf("Can't create BigCache: %w", err)
	}
	return &storage{
		s: s,
	}, nil

}

func (s *storage) set(key string, r *cachedResponse) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(r); err != nil {
		return fmt.Errorf("Can't encode cache entry to gob: %w", err)
	}

	if err := s.s.Set(key, buf.Bytes()); err != nil {
		return fmt.Errorf("Can't save cache entry: %w", err)
	}

	return nil
}

func (s *storage) get(key string) (*cachedResponse, error) {
	entry, err := s.s.Get(key)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(entry)
	dec := gob.NewDecoder(buf)
	var r cachedResponse
	if err = dec.Decode(&r); err != nil {
		return nil, fmt.Errorf("Can't decore cache entry: %w", err)
	}

	return &r, nil
}

func (s *storage) append(key string, r *cachedResponse) error {
	cached, err := s.get(key)
	if err == bigcache.ErrEntryNotFound {
		return s.set(key, r)
	} else if err != nil {
		return err
	}
	// append data
	cached.Data = append(cached.Data, r.Data...)

	return s.set(key, cached)
}

func (s *storage) delete(key string) error {
	return s.s.Delete(key)
}
