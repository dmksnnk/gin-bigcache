package gbcache

import (
	"github.com/allegro/bigcache/v2"
	"github.com/gin-gonic/gin"
)

// cachedWriter is a proxy that writes data into a cache and response
type cachedWriter struct {
	gin.ResponseWriter
	status  int
	written bool
	storage *storage
	key     string
	log     bigcache.Logger
}

func newCachedWriter(storage *storage, writer gin.ResponseWriter, key string, log bigcache.Logger) *cachedWriter {
	return &cachedWriter{
		ResponseWriter: writer,
		status:         0,
		written:        false,
		storage:        storage,
		key:            key,
		log:            log,
	}
}

func (w *cachedWriter) WriteHeader(code int) {
	w.status = code
	w.written = true
	w.ResponseWriter.WriteHeader(code)
}

func (w *cachedWriter) Status() int {
	return w.ResponseWriter.Status()
}

func (w *cachedWriter) Written() bool {
	return w.ResponseWriter.Written()
}

func (w *cachedWriter) Write(data []byte) (int, error) {
	ret, err := w.ResponseWriter.Write(data)
	// don't save cache for statuses >= 300
	if err != nil || w.Status() >= 300 {
		return ret, err
	}

	r := cachedResponse{
		Status: w.Status(),
		Header: w.Header(),
		Data:   data,
	}

	if err := w.storage.append(w.key, &r); err != nil {
		w.log.Printf("Can't append cache: %s", err.Error())
	}

	return ret, nil
}

func (w *cachedWriter) WriteString(data string) (n int, err error) {
	ret, err := w.ResponseWriter.WriteString(data)
	// don't save cache for statuses >= 300
	if err != nil || w.Status() >= 300 {
		return ret, err
	}

	r := cachedResponse{
		Status: w.Status(),
		Header: w.Header(),
		Data:   []byte(data),
	}

	if err := w.storage.append(w.key, &r); err != nil {
		w.log.Printf("Can't set cache: %s", err.Error())
	}
	return ret, err
}
