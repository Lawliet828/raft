package main

import (
	"encoding/json"
	"fmt"
	"gframework/log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/hashicorp/raft"
)

const (
	EnableWriteTrue  = int32(1)
	EnableWriteFalse = int32(0)
)

type HttpServer struct {
	raft        *raftNodeInfo
	enableWrite int32
}

func NewHttpServer(ctx *CachedContext) *HttpServer {
	server := &HttpServer{
		raft:        ctx.raft,
		enableWrite: EnableWriteFalse,
	}

	http.HandleFunc("/set", server.doSet)
	http.HandleFunc("/get", server.doGet)
	http.HandleFunc("/join", server.doJoin)
	return server
}

func (h *HttpServer) checkPermission() bool {
	return atomic.LoadInt32(&h.enableWrite) == EnableWriteTrue
}

func (h *HttpServer) setLeaderFlag(flag bool) {
	if flag {
		atomic.StoreInt32(&h.enableWrite, EnableWriteTrue)
	} else {
		atomic.StoreInt32(&h.enableWrite, EnableWriteFalse)
	}
}

func (h *HttpServer) doGet(w http.ResponseWriter, r *http.Request) {
	vars := r.URL.Query()

	key := vars.Get("key")
	if key == "" {
		log.Error("doGet() error, get nil key")
		fmt.Fprint(w, "")
		return
	}

	ret := h.raft.fsm.cm.Get(key)
	fmt.Fprintf(w, "%s\n", ret)
}

// doSet saves data to cache, only raft master node provides this api
func (h *HttpServer) doSet(w http.ResponseWriter, r *http.Request) {
	if !h.checkPermission() {
		fmt.Fprint(w, "write method not allowed\n")
		return
	}
	vars := r.URL.Query()

	key := vars.Get("key")
	value := vars.Get("value")
	if key == "" || value == "" {
		log.Error("doSet() error, get nil key or nil value")
		fmt.Fprint(w, "param error\n")
		return
	}

	event := logEntryData{Key: key, Value: value}
	eventBytes, err := json.Marshal(event)
	if err != nil {
		log.Errorf("json.Marshal failed, err:%v", err)
		fmt.Fprint(w, "internal error\n")
		return
	}

	applyFuture := h.raft.raft.Apply(eventBytes, 5*time.Second)
	if err := applyFuture.Error(); err != nil {
		log.Errorf("raft.Apply failed:%v", err)
		fmt.Fprint(w, "internal error\n")
		return
	}

	fmt.Fprintf(w, "ok\n")
}

// doJoin handles joining cluster request
func (h *HttpServer) doJoin(w http.ResponseWriter, r *http.Request) {
	vars := r.URL.Query()

	peerAddress := vars.Get("peerAddress")
	if peerAddress == "" {
		log.Error("invalid PeerAddress")
		fmt.Fprint(w, "invalid peerAddress\n")
		return
	}
	addPeerFuture := h.raft.raft.AddVoter(raft.ServerID(peerAddress), raft.ServerAddress(peerAddress), 0, 0)
	if err := addPeerFuture.Error(); err != nil {
		log.Errorf("Error joining peer to raft, peer address:%s, err:%v, code:%d", peerAddress, err, http.StatusInternalServerError)
		fmt.Fprint(w, "internal error\n")
		return
	}
	fmt.Fprint(w, "ok")
}
