package gbcache

import (
	"encoding/gob"
	"fmt"
	"net/http"
	"net/url"

	"github.com/allegro/bigcache/v2"
	"github.com/gin-gonic/gin"
)

type Cache struct {
	storage *storage
	log     bigcache.Logger
}

type cachedResponse struct {
	Status int
	Header http.Header
	Data   []byte
}

// New creates new cache with given bigcache config
func New(cfg bigcache.Config) (*Cache, error) {
	storage, err := newStorage(cfg)
	if err != nil {
		return nil, fmt.Errorf("Can't create cache storage: %w", err)
	}

	gob.Register(&cachedResponse{})

	var log bigcache.Logger
	// if logger logger is provided - use it, othervise use default
	if cfg.Logger != nil {
		log = cfg.Logger
	} else {
		log = bigcache.DefaultLogger()
	}

	return &Cache{
		storage: storage,
		log:     log,
	}, nil
}

// fallback to regular handling if can't process request from cache
func (cache *Cache) fallback(handle gin.HandlerFunc, c *gin.Context, err error, msg string) {
	cache.log.Printf("%s: %s", msg, err.Error())
	handle(c)
}

func (cache *Cache) do(key string, c *gin.Context, handle gin.HandlerFunc) (*cachedResponse, error) {
	cached, err := cache.storage.get(key)
	if err == nil {
		return cached, nil
	}

	if err != bigcache.ErrEntryNotFound {
		return nil, err
	}
	// If no cache entry found - get new one from respose
	// replace writer
	writer := newCachedWriter(cache.storage, c.Writer, key, cache.log)
	c.Writer = writer
	handle(c)

	// delete if was aborted
	if c.IsAborted() {
		if err := cache.storage.delete(key); err != nil {
			cache.log.Printf("Can't delete key from storage: %s", err.Error())
		}
	}
	return nil, nil

}

// CachePage caches response from gin handler function
func (cache *Cache) CachePage(handle gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := url.QueryEscape(c.Request.URL.RequestURI())

		cached, err := cache.do(key, c, handle)
		if err != nil {
			cache.fallback(handle, c, err, "Can't get entry from cache")
			return
		}
		if _, err := c.Writer.Write(cached.Data); err != nil {
			cache.fallback(handle, c, err, "Can't write to response writer")
			return
		}
		c.Writer.WriteHeader(cached.Status)
		for k, vals := range cached.Header {
			for _, v := range vals {
				c.Writer.Header().Set(k, v)
			}
		}
	}

}

// CachePageWithoutQuery adds ability to ignore GET query parameters
func (cache *Cache) CachePageWithoutQuery(handle gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := url.QueryEscape(c.Request.URL.Path)

		cached, err := cache.do(key, c, handle)
		if err != nil {
			cache.fallback(handle, c, err, "Can't get entry from cache")
			return
		}
		if _, err := c.Writer.Write(cached.Data); err != nil {
			cache.fallback(handle, c, err, "Can't write to response writer")
			return
		}
		c.Writer.WriteHeader(cached.Status)
		for k, vals := range cached.Header {
			for _, v := range vals {
				c.Writer.Header().Set(k, v)
			}
		}
	}
}

// CachePageWithoutQuery returns a page from cache without headers
func (cache *Cache) CachePageWithoutHeader(handle gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := url.QueryEscape(c.Request.URL.RequestURI())

		cached, err := cache.do(key, c, handle)
		if err != nil {
			cache.fallback(handle, c, err, "Can't get entry from cache")
			return
		}
		if _, err := c.Writer.Write(cached.Data); err != nil {
			cache.fallback(handle, c, err, "Can't write to response writer")
			return
		}
		c.Writer.WriteHeader(cached.Status)
	}

}
