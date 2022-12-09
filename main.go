package main

import (
	"fmt"
	"gframework/log"
	"net"
	"net/http"
)

type stCached struct {
	hs   *HttpServer
	opts *Options
	cm   *cacheManager
	raft *raftNodeInfo
}

type stCachedContext struct {
	st *stCached
}

func main() {
	st := &stCached{
		opts: NewOptions(),
		cm:   NewCacheManager(),
	}
	ctx := &stCachedContext{st}

	logConf := log.Config{
		Path:    "./",
		Level:   "debug",
		MaxSize: 1,
	}
	log.Init(logConf)

	var l net.Listener
	var err error
	l, err = net.Listen("tcp", st.opts.httpAddress)
	if err != nil {
		log.Panic(fmt.Sprintf("listen %s failed: %s", st.opts.httpAddress, err))
	}
	log.Infof("http server listen:%s", l.Addr())

	httpServer := NewHttpServer(ctx)
	st.hs = httpServer
	go func() {
		http.Serve(l, httpServer.mux)
	}()

	raft, err := newRaftNode(st.opts, ctx)
	if err != nil {
		log.Panic(fmt.Sprintf("new raft node failed:%v", err))
	}
	st.raft = raft

	if st.opts.joinAddress != "" {
		err = joinRaftCluster(st.opts)
		if err != nil {
			log.Panic(fmt.Sprintf("join raft cluster failed:%v", err))
		}
	}

	// monitor leadership
	for {
		select {
		case leader := <-st.raft.leaderNotifyCh:
			if leader {
				log.Info("become leader, enable write api")
				st.hs.setWriteFlag(true)
			} else {
				log.Info("become follower, close write api")
				st.hs.setWriteFlag(false)
			}
		}
	}
}
