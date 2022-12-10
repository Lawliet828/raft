package main

import (
	"errors"
	"fmt"
	"gframework/log"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type raftNodeInfo struct {
	raft           *raft.Raft
	fsm            *FSM // finite state machine
	leaderNotifyCh chan bool
}

func newRaftNode(opts *Options) (*raftNodeInfo, error) {
	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = raft.ServerID(opts.raftTCPAddress)
	// https://github.com/hashicorp/go-hclog/issues/45
	raftConfig.Logger = hclog.FromStandardLogger(log.GetRaftStdLog(), &hclog.LoggerOptions{Level: hclog.Debug})
	raftConfig.SnapshotInterval = 20 * time.Second
	raftConfig.SnapshotThreshold = 2
	leaderNotifyCh := make(chan bool, 1)
	raftConfig.NotifyCh = leaderNotifyCh

	address, err := net.ResolveTCPAddr("tcp", opts.raftTCPAddress)
	if err != nil {
		return nil, err
	}
	transport, err := raft.NewTCPTransport(address.String(), address, 3, 5*time.Second, os.Stderr)
	if err != nil {
		return nil, err
	}

	if err = os.MkdirAll(opts.dataDir, 0700); err != nil {
		return nil, err
	}
	snapshotStore, err := raft.NewFileSnapshotStore(opts.dataDir, 1, os.Stderr)
	if err != nil {
		return nil, err
	}

	logStore, err := raftboltdb.NewBoltStore(filepath.Join(opts.dataDir, "raft-log.db"))
	if err != nil {
		return nil, err
	}

	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(opts.dataDir, "raft-stable.db"))
	if err != nil {
		return nil, err
	}

	fsm := NewFSM()

	raftNode, err := raft.NewRaft(raftConfig, fsm, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		return nil, err
	}

	// 启动raft
	if opts.bootstrap {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      raftConfig.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		raftNode.BootstrapCluster(configuration)
	}

	return &raftNodeInfo{raft: raftNode, fsm: fsm, leaderNotifyCh: leaderNotifyCh}, nil
}

// joinRaftCluster joins a node to raft cluster
func joinRaftCluster(opts *Options) error {
	url := fmt.Sprintf("http://%s/join?peerAddress=%s", opts.joinAddress, opts.raftTCPAddress)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if string(body) != "ok" {
		return errors.New(fmt.Sprintf("Error joining cluster: %s", body))
	}

	return nil
}
