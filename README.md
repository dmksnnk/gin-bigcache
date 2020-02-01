# Gin-bigcache

An integration between [Gin](https://github.com/gin-gonic/gin) web framework and [BigCache](https://github.com/allegro/bigcache).

## Usage

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/dmksnnk/gin-bigcache"
    "github.com/allegro/bigcache"
    "time"
)

func main(){
    cache := cache.New(bigcache.DefaultConfig(time.Minute))

    r := gin.Default()
    r.GET("/ping", cache.CachePage(
        func(c *gin.Context) {
            c.JSON(200, gin.H{"message": "pong"})
        })
    )
    r.Run()
}

```
