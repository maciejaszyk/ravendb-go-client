package ravendb

import (
	"fmt"
	"sync"
	"time"
)

// https://sourcegraph.com/github.com/ravendb/ravendb-jvm-client@v4.0/-/blob/src/main/java/net/ravendb/client/documents/identity/HiLoIdGenerator.java

// RangeValue represents an inclusive integer range min to max
type RangeValue struct {
	Min     int
	Max     int
	Current int // TODO: make atomic
}

// NewRangeValue creates a new RangeValue
func NewRangeValue(min int, max int) *RangeValue {
	return &RangeValue{
		Min:     min,
		Max:     max,
		Current: min - 1,
	}
}

// Next returns next id
func (r *RangeValue) Next() int {
	// TODO: make this atomic
	r.Current++
	return r.Current
}

// Curr returns current id
func (r *RangeValue) Curr() int {
	// TODO: make this atomic
	return r.Current
}

// HiLoIDGenerator generates keys server side
type HiLoIDGenerator struct {
	tag                    string
	store                  *DocumentStore
	dbName                 string
	lastRangeAt            time.Time
	lastBatchSize          int
	rangev                 *RangeValue
	prefix                 string
	serverTag              string
	convetions             *DocumentConventions
	identityPartsSeparator string
	lock                   sync.Mutex
}

// NewHiLoIDGenerator creates a HiLoKeyGenerator
func NewHiLoIDGenerator(tag string, store *DocumentStore, dbName string) *HiLoIDGenerator {
	if dbName == "" {
		dbName = store.database
	}
	res := &HiLoIDGenerator{
		tag:           tag,
		store:         store,
		dbName:        dbName,
		lastRangeAt:   time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC),
		lastBatchSize: 0,
		rangev:        NewRangeValue(1, 0),
		prefix:        "",
		serverTag:     "",
		convetions:    store.Conventions,
	}
	res.identityPartsSeparator = res.convetions.IdentityPartsSeparator
	return res
}

func (g *HiLoIDGenerator) getDocumentIDFromID(nextID int) string {
	return fmt.Sprintf("%s%d-%s", g.prefix, nextID, g.serverTag)
}

// GenerateDocumentID returns next key
func (g *HiLoIDGenerator) GenerateDocumentID() string {
	// TODO: propagate error
	id, _ := g.nextID()
	return g.getDocumentIDFromID(id)
}

func (g *HiLoIDGenerator) nextID() (int, error) {
	// TODO: make Next() atomic and reduce lock scope
	g.lock.Lock()
	defer g.lock.Unlock()
	for {
		// local range is not exhausted yet
		rangev := g.rangev
		id := rangev.Next()
		if id <= rangev.Max {
			return id, nil
		}

		// local range is exhausted , need to get a new range
		err := g.getNextRange()
		if err != nil {
			return 0, err
		}
	}
}

func (g *HiLoIDGenerator) getNextRange() error {
	exec := g.store.GetRequestExecutor("").GetCommandExecutor(false)
	cmd := NewNextHiLoCommand(g.tag, g.lastBatchSize, g.lastRangeAt,
		g.identityPartsSeparator, g.rangev.Max)
	res, err := ExecuteNewNextHiLoCommand(exec, cmd)
	if err != nil {
		return err
	}
	g.prefix = res.Prefix
	g.serverTag = res.ServerTag
	g.lastRangeAt = res.GetLastRangeAt()
	g.lastBatchSize = res.LastSize
	g.rangev = NewRangeValue(res.Low, res.High)
	return nil
}

// ReturnUnusedRange returns unused range
func (g *HiLoIDGenerator) ReturnUnusedRange() error {
	cmd := NewHiLoReturnCommand(g.tag, g.rangev.Curr(), g.rangev.Max)
	// TODO: use store.getRequestsExecutor().Exec()
	exec := g.store.getSimpleExecutor()
	return ExecuteHiLoReturnCommand(exec, cmd)
}
