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

type httpServer struct {
	ctx         *stCachedContext
	mux         *http.ServeMux
	enableWrite int32
}

func NewHttpServer(ctx *stCachedContext) *httpServer {
	mux := http.NewServeMux()
	s := &httpServer{
		ctx:         ctx,
		mux:         mux,
		enableWrite: EnableWriteFalse,
	}

	mux.HandleFunc("/set", s.doSet)
	mux.HandleFunc("/get", s.doGet)
	mux.HandleFunc("/join", s.doJoin)
	return s
}

func (h *httpServer) checkWritePermission() bool {
	return atomic.LoadInt32(&h.enableWrite) == EnableWriteTrue
}

func (h *httpServer) setWriteFlag(flag bool) {
	if flag {
		atomic.StoreInt32(&h.enableWrite, EnableWriteTrue)
	} else {
		atomic.StoreInt32(&h.enableWrite, EnableWriteFalse)
	}
}

func (h *httpServer) doGet(w http.ResponseWriter, r *http.Request) {
	vars := r.URL.Query()

	key := vars.Get("key")
	if key == "" {
		log.Error("doGet() error, get nil key")
		fmt.Fprint(w, "")
		return
	}

	ret := h.ctx.st.cm.Get(key)
	fmt.Fprintf(w, "%s\n", ret)
}

// doSet saves data to cache, only raft master node provides this api
func (h *httpServer) doSet(w http.ResponseWriter, r *http.Request) {
	if !h.checkWritePermission() {
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

	applyFuture := h.ctx.st.raft.raft.Apply(eventBytes, 5*time.Second)
	if err := applyFuture.Error(); err != nil {
		log.Errorf("raft.Apply failed:%v", err)
		fmt.Fprint(w, "internal error\n")
		return
	}

	fmt.Fprintf(w, "ok\n")
}

// doJoin handles joining cluster request
func (h *httpServer) doJoin(w http.ResponseWriter, r *http.Request) {
	vars := r.URL.Query()

	peerAddress := vars.Get("peerAddress")
	if peerAddress == "" {
		log.Error("invalid PeerAddress")
		fmt.Fprint(w, "invalid peerAddress\n")
		return
	}
	addPeerFuture := h.ctx.st.raft.raft.AddVoter(raft.ServerID(peerAddress), raft.ServerAddress(peerAddress), 0, 0)
	if err := addPeerFuture.Error(); err != nil {
		log.Errorf("Error joining peer to raft, peer address:%s, err:%v, code:%d", peerAddress, err, http.StatusInternalServerError)
		fmt.Fprint(w, "internal error\n")
		return
	}
	fmt.Fprint(w, "ok")
}
