package main

import (
	redigo "github.com/garyburd/redigo/redis"
	"github.com/pierrre/imageserver"
	imageserver_cache_chain "github.com/pierrre/imageserver/cache/chain"
	imageserver_cache_memory "github.com/pierrre/imageserver/cache/memory"
	imageserver_cache_prefix "github.com/pierrre/imageserver/cache/prefix"
	imageserver_cache_redis "github.com/pierrre/imageserver/cache/redis"
	imageserver_http "github.com/pierrre/imageserver/http"
	imageserver_http_parser_graphicsmagick "github.com/pierrre/imageserver/http/parser/graphicsmagick"
	imageserver_http_parser_merge "github.com/pierrre/imageserver/http/parser/merge"
	imageserver_http_parser_source "github.com/pierrre/imageserver/http/parser/source"
	imageserver_processor_graphicsmagick "github.com/pierrre/imageserver/processor/graphicsmagick"
	imageserver_provider_cache "github.com/pierrre/imageserver/provider/cache"
	imageserver_provider_http "github.com/pierrre/imageserver/provider/http"
	"net/http"
	"time"
)

func main() {
	cache := imageserver_cache_chain.ChainCache{
		imageserver_cache_memory.New(10 * 1024 * 1024),
		&imageserver_cache_redis.RedisCache{
			Pool: &redigo.Pool{
				Dial: func() (redigo.Conn, error) {
					return redigo.Dial("tcp", "localhost:6379")
				},
				MaxIdle: 50,
			},
			Expire: time.Duration(7 * 24 * time.Hour),
		},
	}

	server := &imageserver_http.Server{
		HttpServer: &http.Server{
			Addr: ":8080",
		},
		Parser: &imageserver_http_parser_merge.MergeParser{
			&imageserver_http_parser_source.SourceParser{},
			&imageserver_http_parser_graphicsmagick.GraphicsMagickParser{},
		},
		ImageServer: &imageserver.Server{
			Cache: &imageserver_cache_prefix.PrefixCache{
				Prefix: "processed:",
				Cache:  cache,
			},
			Provider: &imageserver_provider_cache.CacheProvider{
				Cache: &imageserver_cache_prefix.PrefixCache{
					Prefix: "source:",
					Cache:  cache,
				},
				Provider: &imageserver_provider_http.HttpProvider{},
			},
			Processor: &imageserver_processor_graphicsmagick.GraphicsMagickProcessor{
				Executable: "/usr/local/bin/gm",
				AllowedFormats: []string{
					"jpeg",
					"png",
					"bmp",
					"gif",
				},
				DefaultQualities: map[string]string{
					"jpeg": "85",
				},
			},
		},
		Expire: time.Duration(7 * 24 * time.Hour),
	}
	server.Serve()
}
