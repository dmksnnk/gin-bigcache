# Gin-bigcache

An integration between [Gin](https://github.com/gin-gonic/gin) web framework and [BigCache](https://github.com/allegro/bigcache) cache.

## What

This package aims to provide an easy integration of a fast and highly-configurable cache with a quick and popular web framework
to make your web service even faster.

## Usage

```go
package main

import (
    "log"
    "time"

    "github.com/allegro/bigcache/v2"
    "github.com/dmksnnk/gin-bigcache"
    "github.com/gin-gonic/gin"
    )

func main() {
    cache, err := gbcache.New(bigcache.DefaultConfig(time.Minute))
    if err != nil {
        log.Fatalf("Can't create cache!: %s", err.Error())
    }

    r := gin.Default()
    r.GET("/ping", cache.CachePage(
        func(c *gin.Context) {
            c.JSON(200, gin.H{"message": "pong"})
        }),
    )
    r.Run()
}

```

## How

Inspired by [github.com/gin-contrib/cache](https://github.com/gin-contrib/cache) (which doesn't work propely with bigger responses),
Gin-bigcache is a wrapper around bigcache. It stores your responses in [gobs](https://golang.org/pkg/encoding/gob/) in a memory.

Package uses bigcache's config, for better tuning of your cache please refer to [bigcache itself](https://github.com/allegro/bigcache/blob/master/config.go).
