package ipfscluster

import (
	"errors"
	"testing"
	"time"

	rpc "github.com/hsanjuan/go-libp2p-rpc"
	cid "github.com/ipfs/go-cid"
)

type mockComponent struct {
	rpcClient   *rpc.Client
	returnError bool
}

func (c *mockComponent) Shutdown() error {
	return nil
}

func (c *mockComponent) SetClient(client *rpc.Client) {
	c.rpcClient = client
	return
}

type mockAPI struct {
	mockComponent
}

type mockConnector struct {
	mockComponent
}

func (ipfs *mockConnector) Pin(c *cid.Cid) error {
	if ipfs.returnError {
		return errors.New("")
	}
	return nil
}

func (ipfs *mockConnector) Unpin(c *cid.Cid) error {
	if ipfs.returnError {
		return errors.New("")
	}
	return nil
}

func (ipfs *mockConnector) IsPinned(c *cid.Cid) (bool, error) {
	if ipfs.returnError {
		return false, errors.New("")
	}
	return true, nil
}

func testingCluster(t *testing.T) (*Cluster, *mockAPI, *mockConnector, *MapState, *MapPinTracker) {
	api := &mockAPI{}
	ipfs := &mockConnector{}
	cfg := testingConfig()
	st := NewMapState()
	tracker := NewMapPinTracker(cfg)

	cl, err := NewCluster(
		cfg,
		api,
		ipfs,
		st,
		tracker,
	)
	if err != nil {
		t.Fatal("cannot create cluster:", err)
	}
	time.Sleep(3 * time.Second) // make sure a leader is elected
	return cl, api, ipfs, st, tracker
}

func testClusterShutdown(t *testing.T) {
	cl, _, _, _, _ := testingCluster(t)
	err := cl.Shutdown()
	if err != nil {
		t.Error("cluster shutdown failed:", err)
	}
	cl.Shutdown()
	cl, _, _, _, _ = testingCluster(t)
	err = cl.Shutdown()
	if err != nil {
		t.Error("cluster shutdown failed:", err)
	}
}

func TestClusterStateSync(t *testing.T) {
	cl, _, _, st, _ := testingCluster(t)
	defer cleanRaft()
	defer cl.Shutdown()
	_, err := cl.StateSync()
	if err == nil {
		t.Error("expected an error as there is no state to sync")
	}

	c, _ := cid.Decode(testCid)
	err = cl.Pin(c)
	if err != nil {
		t.Fatal("pin should have worked:", err)
	}

	_, err = cl.StateSync()
	if err != nil {
		t.Fatal("sync after pinning should have worked:", err)
	}

	// Modify state on the side so the sync does not
	// happen on an empty slide
	st.RmPin(c)
	_, err = cl.StateSync()
	if err != nil {
		t.Fatal("sync with recover should have worked:", err)
	}
}

func TestClusterPin(t *testing.T) {
	cl, _, _, _, _ := testingCluster(t)
	defer cleanRaft()
	defer cl.Shutdown()

	c, _ := cid.Decode(testCid)
	err := cl.Pin(c)
	if err != nil {
		t.Fatal("pin should have worked:", err)
	}

	// test an error case
	cl.consensus.Shutdown()
	err = cl.Pin(c)
	if err == nil {
		t.Error("expected an error but things worked")
	}
}

func TestClusterUnpin(t *testing.T) {
	cl, _, _, _, _ := testingCluster(t)
	defer cleanRaft()
	defer cl.Shutdown()

	c, _ := cid.Decode(testCid)
	err := cl.Unpin(c)
	if err != nil {
		t.Fatal("pin should have worked:", err)
	}

	// test an error case
	cl.consensus.Shutdown()
	err = cl.Unpin(c)
	if err == nil {
		t.Error("expected an error but things worked")
	}
}

func TestClusterMembers(t *testing.T) {
	cl, _, _, _, _ := testingCluster(t)
	defer cleanRaft()
	defer cl.Shutdown()
	m := cl.Members()
	id := testingConfig().ID
	if len(m) != 1 || m[0].Pretty() != id {
		t.Error("bad Members()")
	}
}

func TestVersion(t *testing.T) {
	cl, _, _, _, _ := testingCluster(t)
	defer cleanRaft()
	defer cl.Shutdown()
	if cl.Version() != Version {
		t.Error("bad Version()")
	}
}
