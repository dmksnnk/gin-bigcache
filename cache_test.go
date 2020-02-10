package gbcache

import (
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/allegro/bigcache/v2"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func doRequest(method, target string, router *gin.Engine) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, target, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	return w
}

func TestCachePage(t *testing.T) {
	cache, err := New(bigcache.DefaultConfig(time.Second))
	assert.Nil(t, err)

	router := gin.New()
	router.GET("/cache_ping", cache.CachePage(func(c *gin.Context) {
		c.String(200, "pong "+fmt.Sprint(time.Now().UnixNano()))
	}))

	w1 := doRequest("GET", "/cache_ping", router)
	w2 := doRequest("GET", "/cache_ping", router)

	assert.Equal(t, 200, w1.Code)
	assert.Equal(t, 200, w2.Code)
	assert.Equal(t, w1.Body.String(), w2.Body.String())
}

func TestCachePageBadResponse(t *testing.T) {
	cache, err := New(bigcache.DefaultConfig(time.Second))
	assert.Nil(t, err)

	router := gin.New()
	router.GET("/cache_ping", cache.CachePage(func(c *gin.Context) {
		c.String(400, "pong "+fmt.Sprint(time.Now().UnixNano()))
	}))

	w1 := doRequest("GET", "/cache_ping", router)
	w2 := doRequest("GET", "/cache_ping", router)

	assert.Equal(t, 400, w1.Code)
	assert.Equal(t, 400, w2.Code)
	assert.NotEqual(t, w1.Body.String(), w2.Body.String())
}

func TestCachePageExpired(t *testing.T) {
	cache, err := New(bigcache.DefaultConfig(time.Second))
	assert.Nil(t, err)

	router := gin.New()
	router.GET("/cache_ping", cache.CachePage(func(c *gin.Context) {
		c.String(200, "pong "+fmt.Sprint(time.Now().UnixNano()))
	}))

	w1 := doRequest("GET", "/cache_ping", router)
	time.Sleep(time.Second * 3)
	w2 := doRequest("GET", "/cache_ping", router)

	assert.Equal(t, 200, w1.Code)
	assert.Equal(t, 200, w2.Code)
	assert.NotEqual(t, w1.Body.String(), w2.Body.String())
}

func TestCachePageAborted(t *testing.T) {
	cache, err := New(bigcache.DefaultConfig(time.Second))
	assert.Nil(t, err)

	router := gin.New()
	router.GET("/cache_ping", cache.CachePage(func(c *gin.Context) {
		c.AbortWithStatusJSON(200, map[string]int64{"time": time.Now().UnixNano()})
	}))

	w1 := doRequest("GET", "/cache_ping", router)
	w2 := doRequest("GET", "/cache_ping", router)

	assert.Equal(t, 200, w1.Code)
	assert.Equal(t, 200, w2.Code)
	assert.NotEqual(t, w1.Body.String(), w2.Body.String())
}

func TestCachePageWithoutQuery(t *testing.T) {
	cache, err := New(bigcache.DefaultConfig(time.Second))
	assert.Nil(t, err)

	router := gin.New()
	router.GET("/cache_ping", cache.CachePageWithoutQuery(func(c *gin.Context) {
		c.String(200, "pong "+fmt.Sprint(time.Now().UnixNano()))
	}))

	w1 := doRequest("GET", "/cache_ping?foo=1", router)
	w2 := doRequest("GET", "/cache_ping?foo=2", router)

	assert.Equal(t, 200, w1.Code)
	assert.Equal(t, 200, w2.Code)
	assert.Equal(t, w1.Body.String(), w2.Body.String())
}
func TestCachePageWithoutHeader(t *testing.T) {
	cache, err := New(bigcache.DefaultConfig(time.Second))
	assert.Nil(t, err)

	router := gin.New()
	router.GET("/cache_ping", cache.CachePageWithoutHeader(func(c *gin.Context) {
		c.String(200, "pong "+fmt.Sprint(time.Now().UnixNano()))
	}))

	w1 := doRequest("GET", "/cache_ping", router)
	w2 := doRequest("GET", "/cache_ping", router)

	assert.Equal(t, 200, w1.Code)
	assert.Equal(t, 200, w2.Code)
	// has content-type
	assert.NotNil(t, w1.Header()["Content-Type"])
	// don't have content-type
	assert.Nil(t, w2.Header()["Content-Type"])
	assert.Equal(t, w1.Body.String(), w2.Body.String())
}
