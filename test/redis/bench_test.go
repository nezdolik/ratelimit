package redis_test

import (
	"context"
	"math/rand"
	"runtime"
	"testing"
	"time"

	pb "github.com/envoyproxy/go-control-plane/envoy/service/ratelimit/v3"
	stats "github.com/lyft/gostats"

	"github.com/envoyproxy/ratelimit/src/config"
	"github.com/envoyproxy/ratelimit/src/limiter"
	"github.com/envoyproxy/ratelimit/src/redis"
	"github.com/envoyproxy/ratelimit/src/redis/driver"
	"github.com/envoyproxy/ratelimit/src/settings"
	"github.com/envoyproxy/ratelimit/src/utils"
	"github.com/envoyproxy/ratelimit/test/common"
)

func BenchmarkParallelDoLimit(b *testing.B) {
	b.Skip("Skip benchmark")

	b.ReportAllocs()

	// See https://github.com/mediocregopher/radix/blob/v3.5.1/bench/bench_test.go#L176
	parallel := runtime.GOMAXPROCS(0)
	poolSize := parallel * runtime.GOMAXPROCS(0)

	do := func(b *testing.B, fn func() error) {
		b.ResetTimer()
		b.SetParallelism(parallel)
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if err := fn(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}

	mkDoLimitBench := func(pipelineWindow time.Duration, pipelineLimit int, rateLimitAlgorithm string) func(*testing.B) {
		return func(b *testing.B) {
			statsStore := stats.NewStore(stats.NewNullSink(), false)
			client := driver.NewClientImpl(statsStore, false, "", "single", "127.0.0.1:6379", poolSize, pipelineWindow, pipelineLimit)
			defer client.Close()

			var cache limiter.RateLimitCache
			timeSource := utils.NewTimeSourceImpl()
			if rateLimitAlgorithm == settings.FixedRateLimit {
				cache = redis.NewFixedRateLimitCacheImpl(client, nil, timeSource, rand.New(utils.NewLockedSource(time.Now().Unix())), 10, nil, 0.8, "")
			} else if rateLimitAlgorithm == settings.WindowedRateLimit {
				cache = redis.NewWindowedRateLimitCacheImpl(client, nil, timeSource, rand.New(utils.NewLockedSource(time.Now().Unix())), 10, nil, 0.8, "")
			} else {
				b.Fatalf("unknown rate limit type %s", rateLimitAlgorithm)
			}
			request := common.NewRateLimitRequest("domain", [][][2]string{{{"key", "value"}}}, 1)
			limits := []*config.RateLimit{config.NewRateLimit(1000000000, pb.RateLimitResponse_RateLimit_SECOND, "key_value", statsStore)}

			// wait for the pool to fill up
			for {
				time.Sleep(50 * time.Millisecond)
				if client.NumActiveConns() >= poolSize {
					break
				}
			}

			b.ResetTimer()

			do(b, func() error {
				cache.DoLimit(context.Background(), request, limits)
				return nil
			})
		}
	}

	// Fixed ratelimit
	b.Run("fixed ratelimit with no pipeline", mkDoLimitBench(0, 0, settings.FixedRateLimit))

	b.Run("fixed ratelimit with pipeline 35us 1", mkDoLimitBench(35*time.Microsecond, 1, settings.FixedRateLimit))
	b.Run("fixed ratelimit with pipeline 75us 1", mkDoLimitBench(75*time.Microsecond, 1, settings.FixedRateLimit))
	b.Run("fixed ratelimit with pipeline 150us 1", mkDoLimitBench(150*time.Microsecond, 1, settings.FixedRateLimit))
	b.Run("fixed ratelimit with pipeline 300us 1", mkDoLimitBench(300*time.Microsecond, 1, settings.FixedRateLimit))

	b.Run("fixed ratelimit with pipeline 35us 2", mkDoLimitBench(35*time.Microsecond, 2, settings.FixedRateLimit))
	b.Run("fixed ratelimit with pipeline 75us 2", mkDoLimitBench(75*time.Microsecond, 2, settings.FixedRateLimit))
	b.Run("fixed ratelimit with pipeline 150us 2", mkDoLimitBench(150*time.Microsecond, 2, settings.FixedRateLimit))
	b.Run("fixed ratelimit with pipeline 300us 2", mkDoLimitBench(300*time.Microsecond, 2, settings.FixedRateLimit))

	b.Run("fixed ratelimit with pipeline 35us 4", mkDoLimitBench(35*time.Microsecond, 4, settings.FixedRateLimit))
	b.Run("fixed ratelimit with pipeline 75us 4", mkDoLimitBench(75*time.Microsecond, 4, settings.FixedRateLimit))
	b.Run("fixed ratelimit with pipeline 150us 4", mkDoLimitBench(150*time.Microsecond, 4, settings.FixedRateLimit))
	b.Run("fixed ratelimit with pipeline 300us 4", mkDoLimitBench(300*time.Microsecond, 4, settings.FixedRateLimit))

	b.Run("fixed ratelimit with pipeline 35us 8", mkDoLimitBench(35*time.Microsecond, 8, settings.FixedRateLimit))
	b.Run("fixed ratelimit with pipeline 75us 8", mkDoLimitBench(75*time.Microsecond, 8, settings.FixedRateLimit))
	b.Run("fixed ratelimit with pipeline 150us 8", mkDoLimitBench(150*time.Microsecond, 8, settings.FixedRateLimit))
	b.Run("fixed ratelimit with pipeline 300us 8", mkDoLimitBench(300*time.Microsecond, 8, settings.FixedRateLimit))

	b.Run("fixed ratelimit with pipeline 35us 16", mkDoLimitBench(35*time.Microsecond, 16, settings.FixedRateLimit))
	b.Run("fixed ratelimit with pipeline 75us 16", mkDoLimitBench(75*time.Microsecond, 16, settings.FixedRateLimit))
	b.Run("fixed ratelimit with pipeline 150us 16", mkDoLimitBench(150*time.Microsecond, 16, settings.FixedRateLimit))
	b.Run("fixed ratelimit with pipeline 300us 16", mkDoLimitBench(300*time.Microsecond, 16, settings.FixedRateLimit))

	// Windowed ratelimit
	b.Run("windowed ratelimit with no pipeline", mkDoLimitBench(0, 0, settings.WindowedRateLimit))

	b.Run("windowed ratelimit with pipeline 35us 1", mkDoLimitBench(35*time.Microsecond, 1, settings.WindowedRateLimit))
	b.Run("windowed ratelimit with pipeline 75us 1", mkDoLimitBench(75*time.Microsecond, 1, settings.WindowedRateLimit))
	b.Run("windowed ratelimit with pipeline 150us 1", mkDoLimitBench(150*time.Microsecond, 1, settings.WindowedRateLimit))
	b.Run("windowed ratelimit with pipeline 300us 1", mkDoLimitBench(300*time.Microsecond, 1, settings.WindowedRateLimit))

	b.Run("windowed ratelimit with pipeline 35us 2", mkDoLimitBench(35*time.Microsecond, 2, settings.WindowedRateLimit))
	b.Run("windowed ratelimit with pipeline 75us 2", mkDoLimitBench(75*time.Microsecond, 2, settings.WindowedRateLimit))
	b.Run("windowed ratelimit with pipeline 150us 2", mkDoLimitBench(150*time.Microsecond, 2, settings.WindowedRateLimit))
	b.Run("windowed ratelimit with pipeline 300us 2", mkDoLimitBench(300*time.Microsecond, 2, settings.WindowedRateLimit))

	b.Run("windowed ratelimit with pipeline 35us 4", mkDoLimitBench(35*time.Microsecond, 4, settings.WindowedRateLimit))
	b.Run("windowed ratelimit with pipeline 75us 4", mkDoLimitBench(75*time.Microsecond, 4, settings.WindowedRateLimit))
	b.Run("windowed ratelimit with pipeline 150us 4", mkDoLimitBench(150*time.Microsecond, 4, settings.WindowedRateLimit))
	b.Run("windowed ratelimit with pipeline 300us 4", mkDoLimitBench(300*time.Microsecond, 4, settings.WindowedRateLimit))

	b.Run("windowed ratelimit with pipeline 35us 8", mkDoLimitBench(35*time.Microsecond, 8, settings.WindowedRateLimit))
	b.Run("windowed ratelimit with pipeline 75us 8", mkDoLimitBench(75*time.Microsecond, 8, settings.WindowedRateLimit))
	b.Run("windowed ratelimit with pipeline 150us 8", mkDoLimitBench(150*time.Microsecond, 8, settings.WindowedRateLimit))
	b.Run("windowed ratelimit with pipeline 300us 8", mkDoLimitBench(300*time.Microsecond, 8, settings.WindowedRateLimit))

	b.Run("windowed ratelimit with pipeline 35us 16", mkDoLimitBench(35*time.Microsecond, 16, settings.WindowedRateLimit))
	b.Run("windowed ratelimit with pipeline 75us 16", mkDoLimitBench(75*time.Microsecond, 16, settings.WindowedRateLimit))
	b.Run("windowed ratelimit with pipeline 150us 16", mkDoLimitBench(150*time.Microsecond, 16, settings.WindowedRateLimit))
	b.Run("windowed ratelimit with pipeline 300us 16", mkDoLimitBench(300*time.Microsecond, 16, settings.WindowedRateLimit))
}
