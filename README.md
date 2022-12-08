# stcache

A simple cache server showing how to use hashicorp/raft

# doc

https://zhuanlan.zhihu.com/p/58048906

# build

```bash
make build
```

# start

## start leader node1

```bash
sh run.sh 1 1
```

## start follower node2

```bash
sh run.sh 2 0
```

## start follower node3

```bash
sh run.sh 3 0
```