package main

import (
	"fmt"
	"gframework/log"
	"net/http"
)

type CachedContext struct {
	hs   *HttpServer
	opts *Options
	raft *raftNodeInfo
}

func main() {
	ctx := &CachedContext{
		opts: NewOptions(),
	}

	logConf := log.Config{
		Path:    "./",
		Level:   "debug",
		MaxSize: 1,
	}
	log.Init(logConf)

	raft, err := newRaftNode(ctx.opts, ctx)
	if err != nil {
		log.Panic(fmt.Sprintf("new raft node failed:%v", err))
	}
	ctx.raft = raft

	if ctx.opts.joinAddress != "" {
		err = joinRaftCluster(ctx.opts)
		if err != nil {
			log.Panic(fmt.Sprintf("join raft cluster failed:%v", err))
		}
	}

	httpServer := NewHttpServer(ctx)
	ctx.hs = httpServer
	go func() {
		http.ListenAndServe(ctx.opts.httpAddress, nil)
	}()

	// monitor leadership
	for {
		select {
		case leader := <-ctx.raft.leaderNotifyCh:
			if leader {
				log.Info("become leader, enable write api")
				ctx.hs.setWriteFlag(true)
			} else {
				log.Info("become follower, close write api")
				ctx.hs.setWriteFlag(false)
			}
		}
	}
}
