package ipfscluster

import (
	"sync"

	cid "github.com/ipfs/go-cid"
)

// MapState is a very simple database to store the state of the system
// using a Go map. It is thread safe. It implements the State interface.
type MapState struct {
	mux    sync.RWMutex
	PinMap map[string]struct{}
}

// NewMapState initializes the internal map and returns a new MapState object.
func NewMapState() *MapState {
	return &MapState{
		PinMap: make(map[string]struct{}),
	}
}

// AddPin adds a Cid to the internal map.
func (st *MapState) AddPin(c *cid.Cid) error {
	st.mux.Lock()
	defer st.mux.Unlock()
	var a struct{}
	st.PinMap[c.String()] = a
	return nil
}

// RmPin removes a Cid from the internal map.
func (st *MapState) RmPin(c *cid.Cid) error {
	st.mux.Lock()
	defer st.mux.Unlock()
	delete(st.PinMap, c.String())
	return nil
}

// HasPin returns true if the Cid belongs to the State.
func (st *MapState) HasPin(c *cid.Cid) bool {
	st.mux.RLock()
	defer st.mux.RUnlock()
	_, ok := st.PinMap[c.String()]
	return ok
}

// ListPins provides a list of Cids in the State.
func (st *MapState) ListPins() []*cid.Cid {
	st.mux.RLock()
	defer st.mux.RUnlock()
	cids := make([]*cid.Cid, 0, len(st.PinMap))
	for k := range st.PinMap {
		c, _ := cid.Decode(k)
		cids = append(cids, c)
	}
	return cids
}
