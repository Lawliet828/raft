package main

import (
	"encoding/json"
	"gframework/log"
	"io"

	"github.com/hashicorp/raft"
)

type FSM struct {
	ctx *stCachedContext
}

type logEntryData struct {
	Key   string
	Value string
}

// Apply applies a Raft log entry to the key-value store.
func (f *FSM) Apply(logEntry *raft.Log) interface{} {
	e := logEntryData{}
	if err := json.Unmarshal(logEntry.Data, &e); err != nil {
		panic("Failed unmarshalling Raft log entry. This is a bug.")
	}
	ret := f.ctx.st.cm.Set(e.Key, e.Value)
	log.Infof("fms.Apply(), logEntry:%s, ret:%v\n", logEntry.Data, ret)
	return ret
}

// Snapshot returns the latest snapshot
func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	return &snapshot{cm: f.ctx.st.cm}, nil
}

// Restore stores the key-value store to a previous state.
func (f *FSM) Restore(serialized io.ReadCloser) error {
	return f.ctx.st.cm.UnMarshal(serialized)
}
