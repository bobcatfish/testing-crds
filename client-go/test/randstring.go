package test

import (
	"math/rand"
	"strings"
	"sync"
	"time"
)

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyz"
	randSuffixLen = 8
)

// r is used by AppendRandomString to generate a random string. It is seeded with the time
// at import so the strings will be different between test runs.
var (
	r        *rand.Rand
	rndMutex *sync.Mutex
)

// once is used to initialize r
var once sync.Once

func initSeed() {
	seed := time.Now().UTC().UnixNano()
	r = rand.New(rand.NewSource(seed))
	rndMutex = &sync.Mutex{}
}

// appendRandomString will generate a random string that begins with prefix. This is useful
// if you want to make sure that your tests can run at the same time against the same
// environment without conflicting. This method will seed rand with the current time when
// called for the first time.
func appendRandomString(prefix string) string {
	once.Do(initSeed)
	suffix := make([]byte, randSuffixLen)
	rndMutex.Lock()
	for i := range suffix {
		suffix[i] = letterBytes[r.Intn(len(letterBytes))]
	}
	rndMutex.Unlock()
	return strings.Join([]string{prefix, string(suffix)}, "-")
}
