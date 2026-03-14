package feature

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
)

type contextKey string

type CacheTestSuite struct {
	suite.Suite
}

func TestCacheTestSuite(t *testing.T) {
	suite.Run(t, new(CacheTestSuite))
}

func (s *CacheTestSuite) SetupTest() {
	s.flushStores()
}

func (s *CacheTestSuite) TearDownTest() {
	s.flushStores()
}

func (s *CacheTestSuite) flushStores() {
	s.True(facades.Cache().Store("memory").Flush())
	s.True(facades.Cache().Store("redis").Flush())
}

func (s *CacheTestSuite) storeKey(store, key string) string {
	return fmt.Sprintf("cache:%s:%s", store, key)
}

func (s *CacheTestSuite) TestBasicOperations() {
	for _, store := range []string{"memory", "redis"} {
		s.Run(store, func() {
			driver := facades.Cache().Store(store)
			key := s.storeKey(store, "basic")

			s.False(driver.Has(key))
			s.NoError(driver.Put(key, "goravel", time.Second))
			s.True(driver.Has(key))
			s.Equal("goravel", driver.GetString(key))
			s.True(driver.Forget(key))
			s.False(driver.Has(key))

			s.True(driver.Add(key, "framework", time.Second))
			s.False(driver.Add(key, "duplicate", time.Second))
			s.Equal("framework", driver.GetString(key))

			pullKey := s.storeKey(store, "pull")
			s.NoError(driver.Put(pullKey, "pull-value", time.Second))
			s.Equal("pull-value", driver.Pull(pullKey).(string))
			s.False(driver.Has(pullKey))

			foreverKey := s.storeKey(store, "forever")
			s.True(driver.Forever(foreverKey, "forever-value"))
			s.Equal("forever-value", driver.GetString(foreverKey))

			expiresKey := s.storeKey(store, "expires")
			s.NoError(driver.Put(expiresKey, "expires", 100*time.Millisecond))
			time.Sleep(200 * time.Millisecond)
			s.False(driver.Has(expiresKey))

			s.True(driver.Flush())
		})
	}
}

func (s *CacheTestSuite) TestTypedGetAndDefaultValues() {
	for _, store := range []string{"memory", "redis"} {
		s.Run(store, func() {
			driver := facades.Cache().Store(store)

			s.Equal("default", driver.GetString(s.storeKey(store, "missing-string"), "default"))
			s.Equal(10, driver.GetInt(s.storeKey(store, "missing-int"), 10))
			s.Equal(int64(11), driver.GetInt64(s.storeKey(store, "missing-int64"), 11))
			s.True(driver.GetBool(s.storeKey(store, "missing-bool"), true))
			s.Equal("fallback", driver.Get(s.storeKey(store, "missing-any"), "fallback").(string))
			s.Equal("callback", driver.Get(s.storeKey(store, "missing-any-callback"), func() any { return "callback" }).(string))

			stringKey := s.storeKey(store, "string")
			intKey := s.storeKey(store, "int")
			int64Key := s.storeKey(store, "int64")
			boolKey := s.storeKey(store, "bool")

			s.NoError(driver.Put(stringKey, "value", time.Second))
			s.NoError(driver.Put(intKey, 3, time.Second))
			s.NoError(driver.Put(int64Key, int64(4), time.Second))
			s.NoError(driver.Put(boolKey, true, time.Second))

			s.Equal("value", driver.GetString(stringKey))
			s.Equal(3, driver.GetInt(intKey))
			s.Equal(int64(4), driver.GetInt64(int64Key))
			s.True(driver.GetBool(boolKey))
		})
	}
}

func (s *CacheTestSuite) TestIncrementAndDecrement() {
	for _, store := range []string{"memory", "redis"} {
		s.Run(store, func() {
			driver := facades.Cache().Store(store)
			key := s.storeKey(store, "counter")

			incremented, err := driver.Increment(key)
			s.NoError(err)
			s.Equal(int64(1), incremented)

			incremented, err = driver.Increment(key, 2)
			s.NoError(err)
			s.Equal(int64(3), incremented)

			decremented, err := driver.Decrement(key)
			s.NoError(err)
			s.Equal(int64(2), decremented)

			decremented, err = driver.Decrement(key, 2)
			s.NoError(err)
			s.Equal(int64(0), decremented)
		})
	}
}

func (s *CacheTestSuite) TestRememberAndRememberForever() {
	for _, store := range []string{"memory", "redis"} {
		s.Run(store, func() {
			driver := facades.Cache().Store(store)

			rememberKey := s.storeKey(store, "remember")
			var rememberCallbackCount atomic.Int32

			first, err := driver.Remember(rememberKey, time.Second, func() (any, error) {
				rememberCallbackCount.Add(1)
				return "remember-value", nil
			})
			s.NoError(err)
			s.Equal("remember-value", first)

			second, err := driver.Remember(rememberKey, time.Second, func() (any, error) {
				rememberCallbackCount.Add(1)
				return "new-value", nil
			})
			s.NoError(err)
			s.Equal("remember-value", second)
			s.EqualValues(1, rememberCallbackCount.Load())

			errorValue, err := driver.Remember(s.storeKey(store, "remember-error"), time.Second, func() (any, error) {
				return nil, errors.New("remember error")
			})
			s.EqualError(err, "remember error")
			s.Nil(errorValue)

			rememberForeverKey := s.storeKey(store, "remember-forever")
			var rememberForeverCallbackCount atomic.Int32

			foreverFirst, err := driver.RememberForever(rememberForeverKey, func() (any, error) {
				rememberForeverCallbackCount.Add(1)
				return "remember-forever-value", nil
			})
			s.NoError(err)
			s.Equal("remember-forever-value", foreverFirst)

			foreverSecond, err := driver.RememberForever(rememberForeverKey, func() (any, error) {
				rememberForeverCallbackCount.Add(1)
				return "new-forever-value", nil
			})
			s.NoError(err)
			s.Equal("remember-forever-value", foreverSecond)
			s.EqualValues(1, rememberForeverCallbackCount.Load())

			errorForeverValue, err := driver.RememberForever(s.storeKey(store, "remember-forever-error"), func() (any, error) {
				return nil, errors.New("remember forever error")
			})
			s.EqualError(err, "remember forever error")
			s.Nil(errorForeverValue)
		})
	}
}

func (s *CacheTestSuite) TestLockAndContext() {
	for _, store := range []string{"memory", "redis"} {
		s.Run(store, func() {
			driver := facades.Cache().Store(store)
			lockKey := s.storeKey(store, "lock")

			firstLock := driver.Lock(lockKey, time.Second)
			s.True(firstLock.Get())

			secondLock := driver.Lock(lockKey, time.Second)
			s.False(secondLock.Get())
			s.False(secondLock.Release())
			s.True(secondLock.ForceRelease())

			thirdLock := driver.Lock(lockKey, time.Second)
			s.True(thirdLock.Get())
			s.True(thirdLock.Release())

			callbackCalled := make(chan struct{}, 1)
			blockingLock := driver.Lock(lockKey, time.Second)
			s.True(blockingLock.Get())

			go func() {
				waiterLock := driver.Lock(lockKey, time.Second)
				if waiterLock.BlockWithTicker(500*time.Millisecond, 10*time.Millisecond, func() {
					callbackCalled <- struct{}{}
				}) {
					_ = waiterLock.Release()
				}
			}()

			time.Sleep(50 * time.Millisecond)
			s.True(blockingLock.Release())

			select {
			case <-callbackCalled:
			case <-time.After(time.Second):
				s.Fail("expected lock callback to be called")
			}

			ctxDriver := driver.WithContext(context.WithValue(context.Background(), contextKey("trace"), "cache-test"))
			ctxKey := s.storeKey(store, "with-context")
			s.NoError(ctxDriver.Put(ctxKey, "ctx-value", time.Second))
			s.Equal("ctx-value", driver.GetString(ctxKey))
		})
	}
}

func (s *CacheTestSuite) TestDockerAndStore() {
	redisDriver := facades.Cache().Store("redis")
	dockerDriver, err := redisDriver.Docker()
	s.NoError(err)
	s.NotNil(dockerDriver)

	s.NoError(redisDriver.Put("cache:redis:docker", "ok", time.Second))
	s.Equal("ok", redisDriver.GetString("cache:redis:docker"))
}
