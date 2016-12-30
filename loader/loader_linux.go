package loader

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/lyft/goruntime/snapshot"
	"github.com/lyft/goruntime/snapshot/entry"
	"github.com/lyft/goruntime/stats"
	"golang.org/x/sys/unix"

	logger "github.com/Sirupsen/logrus"
)

type loaderStats struct {
	loadAttempts stats.Counter
	loadFailures stats.Counter
	numValues    stats.Gauge
}

func newLoaderStats(scope stats.Scope) loaderStats {
	ret := loaderStats{}
	ret.loadAttempts = scope.NewCounter("load_attempts")
	ret.loadFailures = scope.NewCounter("load_failures")
	ret.numValues = scope.NewGauge("num_values")
	return ret
}

// Implementation of Loader that watches a symlink and reads from the filesystem.
type Loader struct {
	watcher         *fsnotify.Watcher
	watchPath       string
	subdirectory    string
	currentSnapshot snapshot.IFace
	nextSnapshot    snapshot.IFace
	updateLock      sync.RWMutex
	callbacks       []chan<- int
	stats           loaderStats
}

func (l *Loader) Snapshot() snapshot.IFace {
	// This could probably be done with an atomic pointer but the unsafe pointers the atomics
	// take scared me so skipping for now.
	l.updateLock.RLock()
	defer l.updateLock.RUnlock()
	return l.currentSnapshot
}

func (l *Loader) AddUpdateCallback(callback chan<- int) {
	l.callbacks = append(l.callbacks, callback)
}

func (l *Loader) onSymLinkSwap() {
	targetDir := filepath.Join(l.watchPath, l.subdirectory)
	logger.Debugf("runtime symlink swap. loading new snapshot at %s",
		targetDir)

	l.nextSnapshot = snapshot.New()
	filepath.Walk(targetDir, l.walkDirectoryCallback)

	// This could probably be done with an atomic pointer but the unsafe pointers the atomics
	// take scared me so skipping for now.
	l.stats.loadAttempts.Inc()
	l.stats.numValues.Set(uint64(len(l.nextSnapshot.Entries())))
	l.updateLock.Lock()
	l.currentSnapshot = l.nextSnapshot
	l.updateLock.Unlock()

	l.nextSnapshot = nil
	for _, callback := range l.callbacks {
		// Arbitrary integer just to wake up channel.
		callback <- 1
	}
}

type walkError struct {
	err error
}

func checkWalkError(err error) {
	if err != nil {
		panic(walkError{err})
	}
}

func (l *Loader) walkDirectoryCallback(path string, info os.FileInfo, err error) error {
	defer func() {
		if e := recover(); e != nil {
			if localError, ok := e.(walkError); ok {
				l.stats.loadFailures.Inc()
				logger.Warnf("runtime: error processing %s: %s", path,
					localError.err.Error())
			} else {
				panic(e)
			}
		}
	}()

	logger.Debugf("runtime: processing %s", path)
	checkWalkError(err)
	if !info.IsDir() {
		contents, err := ioutil.ReadFile(path)
		checkWalkError(err)

		key, err := filepath.Rel(filepath.Join(l.watchPath, l.subdirectory), path)
		checkWalkError(err)

		key = strings.Replace(key, "/", ".", -1)
		stringValue := string(contents)
		entry := entry.New(stringValue, 0, false)

		uint64Value, err := strconv.ParseUint(strings.TrimSpace(stringValue), 10, 64)
		if err == nil {
			entry.Uint64Value = uint64Value
			entry.Uint64Valid = true
		}

		logger.Debugf("runtime: adding key=%s value=%s uint=%t", key,
			stringValue, entry.Uint64Valid)
		l.nextSnapshot.SetEntry(key, entry)
	}

	return nil
}

func New(runtimePath string, runtimeSubdirectory string, scope stats.Scope) IFace {
	if runtimePath == "" || runtimeSubdirectory == "" {
		logger.Warnf("no runtime configuration. using nil loader.")
		return &Nil{}
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Fatalf("unable to create runtime watcher")
	}

	// We need to watch the directory that the symlink is in vs. the symlink itself.
	err = watcher.Add(filepath.Dir(runtimePath))

	if err != nil {
		logger.Fatalf("unable to create runtime watcher")
	}

	newLoader := Loader{
		watcher, runtimePath, runtimeSubdirectory, nil, nil, sync.RWMutex{}, nil,
		newLoaderStats(scope)}
	newLoader.onSymLinkSwap()

	go func() {
		for {
			select {
			case ev := <-watcher.Events:
				if ev.Name == runtimePath && (ev.Op&unix.IN_MOVED_TO) != 0 {
					newLoader.onSymLinkSwap()
				}
			case err := <-watcher.Errors:
				logger.Warnf("runtime watch error:", err)
			}
		}
	}()

	return &newLoader
}
