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
	"goravel/tests"
)

type contextKey string

type CacheTestSuite struct {
	suite.Suite
	tests.TestCase
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

// TestTagsNamespacing covers Driver.Tags (framework#1579, redis#170): a tagged
// view keeps its items apart from the untagged store and from other tag lists.
func (s *CacheTestSuite) TestTagsNamespacing() {
	for _, store := range []string{"memory", "redis"} {
		s.Run(store, func() {
			driver := facades.Cache().Store(store)
			tagged := driver.Tags("people", "authors")
			key := s.storeKey(store, "john")

			s.NoError(tagged.Put(key, "John", time.Minute))
			s.Equal("John", tagged.GetString(key))
			s.True(tagged.Has(key))

			// Tagged items are invisible to the untagged store and to any other tag list, in any order.
			s.Nil(driver.Get(key))
			s.False(driver.Has(key))
			s.Nil(driver.Tags("people").Get(key))
			s.Nil(driver.Tags("authors", "people").Get(key))
			s.Equal("default", driver.Tags("people").Get(key, "default"))

			// The same key lives separately under each tag list and under the untagged store.
			s.NoError(driver.Put(key, "plain", time.Minute))
			s.NoError(driver.Tags("people").Put(key, "people", time.Minute))
			s.NoError(driver.Tags("authors", "people").Put(key, "reversed", time.Minute))

			s.Equal("plain", driver.GetString(key))
			s.Equal("people", driver.Tags("people").GetString(key))
			s.Equal("reversed", driver.Tags("authors", "people").GetString(key))
			s.Equal("John", tagged.GetString(key))
		})
	}
}

func (s *CacheTestSuite) TestTagsFlushRemovesOnlyTaggedItems() {
	for _, store := range []string{"memory", "redis"} {
		s.Run(store, func() {
			driver := facades.Cache().Store(store)

			s.NoError(driver.Tags("people", "authors").Put(s.storeKey(store, "john"), "John", time.Minute))
			s.True(driver.Tags("people").Forever(s.storeKey(store, "anne"), "Anne"))
			s.True(driver.Tags("authors").Forever(s.storeKey(store, "bob"), "Bob"))
			s.True(driver.Tags("other").Forever(s.storeKey(store, "tom"), "Tom"))
			s.True(driver.Forever(s.storeKey(store, "plain"), "Plain"))

			s.True(driver.Tags("people").Flush())

			s.Nil(driver.Tags("people", "authors").Get(s.storeKey(store, "john")))
			s.Nil(driver.Tags("people").Get(s.storeKey(store, "anne")))
			s.Equal("Bob", driver.Tags("authors").GetString(s.storeKey(store, "bob")))
			s.Equal("Tom", driver.Tags("other").GetString(s.storeKey(store, "tom")))
			s.Equal("Plain", driver.GetString(s.storeKey(store, "plain")))

			// Only the second name of the view indexes the item, so a flush that
			// stops at the first name leaves it behind.
			multiKey := s.storeKey(store, "multi")
			s.True(driver.Tags("nobody").Forever(multiKey, "multi"))
			s.Equal("multi", driver.Tags("nobody").GetString(multiKey))
			s.True(driver.Tags("people", "nobody").Flush())
			s.Nil(driver.Tags("nobody").Get(multiKey))

			// A flushed tag can be reused.
			s.True(driver.Tags("people").Forever(s.storeKey(store, "anne"), "Again"))
			s.Equal("Again", driver.Tags("people").GetString(s.storeKey(store, "anne")))
		})
	}
}

func (s *CacheTestSuite) TestTagsStoreFlush() {
	for _, store := range []string{"memory", "redis"} {
		s.Run(store, func() {
			driver := facades.Cache().Store(store)
			key := s.storeKey(store, "flush")

			s.True(driver.Tags("a", "b").Forever(key, "tagged"))
			s.True(driver.Forever(key, "plain"))

			// Flushing the whole store drops the tagged items too.
			s.True(driver.Flush())
			s.False(driver.Tags("a", "b").Has(key))
			s.False(driver.Has(key))

			// Flushing a tag that holds nothing is a no-op, and the tag still works.
			s.True(driver.Tags("b").Forever(key, "keep"))
			s.True(driver.Tags("a").Flush())
			s.Equal("keep", driver.Tags("b").GetString(key))
			s.True(driver.Tags("a", "b").Forever(key, "after-flush"))
			s.Equal("after-flush", driver.Tags("a", "b").GetString(key))
		})
	}
}

func (s *CacheTestSuite) TestTagsExpiration() {
	for _, store := range []string{"memory", "redis"} {
		s.Run(store, func() {
			driver := facades.Cache().Store(store)
			key := s.storeKey(store, "expiring")

			s.NoError(driver.Tags("ttl").Put(key, "value", time.Minute))
			s.True(driver.Tags("ttl").Has(key))
			s.NoError(driver.Tags("ttl").Put(key, "value", 100*time.Millisecond))
			time.Sleep(200 * time.Millisecond)
			s.False(driver.Tags("ttl").Has(key))
			s.Nil(driver.Tags("ttl").Get(key))

			// Flushing a tag whose items have expired still succeeds.
			s.True(driver.Tags("ttl").Flush())
		})
	}
}

func (s *CacheTestSuite) TestTagsOperations() {
	for _, store := range []string{"memory", "redis"} {
		s.Run(store, func() {
			driver := facades.Cache().Store(store)
			tagged := driver.Tags("ops")

			addKey := s.storeKey(store, "add")
			s.True(tagged.Add(addKey, "first", time.Minute))
			s.False(tagged.Add(addKey, "second", time.Minute))
			s.Equal("first", tagged.GetString(addKey))

			counterKey := s.storeKey(store, "counter")
			incremented, err := tagged.Increment(counterKey)
			s.NoError(err)
			s.Equal(int64(1), incremented)
			incremented, err = tagged.Increment(counterKey, 4)
			s.NoError(err)
			s.Equal(int64(5), incremented)
			decremented, err := tagged.Decrement(counterKey, 2)
			s.NoError(err)
			s.Equal(int64(3), decremented)
			s.Equal(int64(3), tagged.GetInt64(counterKey))

			s.NoError(tagged.Put(s.storeKey(store, "not-a-number"), "goravel", time.Minute))
			_, err = tagged.Increment(s.storeKey(store, "not-a-number"))
			s.Error(err)
			_, err = tagged.Decrement(s.storeKey(store, "not-a-number"))
			s.Error(err)

			pullKey := s.storeKey(store, "pull")
			s.True(tagged.Forever(pullKey, "pull-value"))
			s.Equal("pull-value", tagged.Pull(pullKey))
			s.False(tagged.Has(pullKey))
			s.Equal("default", tagged.Pull(pullKey, "default"))

			forgetKey := s.storeKey(store, "forget")
			s.True(tagged.Forever(forgetKey, "value"))
			s.True(tagged.Forget(forgetKey))
			s.False(tagged.Has(forgetKey))

			var calls atomic.Int32
			rememberKey := s.storeKey(store, "remember")
			value, err := tagged.Remember(rememberKey, time.Minute, func() (any, error) {
				calls.Add(1)
				return "remember-value", nil
			})
			s.NoError(err)
			s.Equal("remember-value", value)
			value, err = tagged.Remember(rememberKey, time.Minute, func() (any, error) {
				calls.Add(1)
				return "new-value", nil
			})
			s.NoError(err)
			s.Equal("remember-value", value)
			s.EqualValues(1, calls.Load())

			foreverKey := s.storeKey(store, "remember-forever")
			foreverValue, err := tagged.RememberForever(foreverKey, func() (any, error) {
				return "remember-forever-value", nil
			})
			s.NoError(err)
			s.Equal("remember-forever-value", foreverValue)

			errorKey := s.storeKey(store, "remember-error")
			_, err = tagged.Remember(errorKey, time.Minute, func() (any, error) {
				return nil, errors.New("remember error")
			})
			s.EqualError(err, "remember error")
			s.False(tagged.Has(errorKey))

			// Every write above went through the tagged view, so the tag flush takes them all.
			s.True(tagged.Flush())
			s.False(tagged.Has(addKey))
			s.False(tagged.Has(counterKey))
			s.False(tagged.Has(rememberKey))
			s.False(tagged.Has(foreverKey))
		})
	}
}

func (s *CacheTestSuite) TestTagsTypedGetters() {
	for _, store := range []string{"memory", "redis"} {
		s.Run(store, func() {
			driver := facades.Cache().Store(store)
			tagged := driver.Tags("typed")

			s.Equal("default", tagged.GetString(s.storeKey(store, "missing-string"), "default"))
			s.Equal(10, tagged.GetInt(s.storeKey(store, "missing-int"), 10))
			s.Equal(int64(11), tagged.GetInt64(s.storeKey(store, "missing-int64"), 11))
			s.True(tagged.GetBool(s.storeKey(store, "missing-bool"), true))
			s.Equal("fallback", tagged.Get(s.storeKey(store, "missing-any"), "fallback").(string))
			s.Equal("callback", tagged.Get(s.storeKey(store, "missing-callback"), func() any { return "callback" }).(string))

			s.NoError(tagged.Put(s.storeKey(store, "string"), "value", time.Minute))
			s.NoError(tagged.Put(s.storeKey(store, "int"), 3, time.Minute))
			s.NoError(tagged.Put(s.storeKey(store, "int64"), int64(4), time.Minute))
			s.NoError(tagged.Put(s.storeKey(store, "bool"), true, time.Minute))

			s.Equal("value", tagged.GetString(s.storeKey(store, "string")))
			s.Equal(3, tagged.GetInt(s.storeKey(store, "int")))
			s.Equal(int64(4), tagged.GetInt64(s.storeKey(store, "int64")))
			s.True(tagged.GetBool(s.storeKey(store, "bool")))

			// The typed values are not visible outside the tag list.
			s.Nil(driver.Get(s.storeKey(store, "string")))
		})
	}
}

func (s *CacheTestSuite) TestTagsWithoutNames() {
	for _, store := range []string{"memory", "redis"} {
		s.Run(store, func() {
			driver := facades.Cache().Store(store)
			key := s.storeKey(store, "empty-tags")

			// An empty tag list still namespaces the item away from the untagged store...
			s.True(driver.Tags().Forever(key, "value"))
			s.Equal("value", driver.Tags().GetString(key))
			s.Nil(driver.Get(key))

			// ...but no tag indexes it, so flushing the empty set removes nothing.
			s.True(driver.Tags().Flush())
			s.Equal("value", driver.Tags().GetString(key))
		})
	}
}

func (s *CacheTestSuite) TestTagsNamesAreLengthPrefixed() {
	for _, store := range []string{"memory", "redis"} {
		s.Run(store, func() {
			driver := facades.Cache().Store(store)
			key := s.storeKey(store, "names")

			s.True(driver.Tags("ab").Forever(key, "joined"))
			s.True(driver.Tags("a", "b").Forever(key, "split"))
			s.True(driver.Tags("a|b").Forever(key, "piped"))

			s.Equal("joined", driver.Tags("ab").GetString(key))
			s.Equal("split", driver.Tags("a", "b").GetString(key))
			s.Equal("piped", driver.Tags("a|b").GetString(key))

			// "ab" is one tag named "ab", not the two tags "a" and "b".
			s.True(driver.Tags("a").Flush())
			s.Equal("joined", driver.Tags("ab").GetString(key))
			s.Nil(driver.Tags("a", "b").Get(key))
			s.Equal("piped", driver.Tags("a|b").GetString(key))

			s.True(driver.Tags("ab").Flush())
			s.Nil(driver.Tags("ab").Get(key))
			s.Equal("piped", driver.Tags("a|b").GetString(key))
		})
	}
}
