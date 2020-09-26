package lockservice

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/rpc"
	"os"
	"sync"
	"time"
)

type LockServer struct {
	mu	  sync.Mutex
	l	  net.Listener
	dead  bool // for test_test.go
	dying bool // for test_test.go

	// for each lock name, is it locked?
	locks map[uint64]bool

	// each CID's last sequence #
	lastSeq   map[uint64]uint64
	lastReply map[uint64]bool
}

//
// server Lock RPC handler.
//
func (ls *LockServer) TryLock(args *LockArgs, reply *LockReply) error {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	// Check if seqno has been seen, and reply from the cache if so
	last, ok := ls.lastSeq[args.CID]
	if ok && args.Seq <= last {
		reply.OK = ls.lastReply[args.CID]
		return nil
	}
	ls.lastSeq[args.CID] = args.Seq

	locked, _ := ls.locks[args.Lockname]

	if locked {
		reply.OK = false
	} else {
		reply.OK = true
		ls.locks[args.Lockname] = true
	}
	ls.lastReply[args.CID] = reply.OK
	return nil
}

//
// server Unlock RPC handler.
//
func (ls *LockServer) Unlock(args *UnlockArgs, reply *UnlockReply) error {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	last, ok := ls.lastSeq[args.CID]
	if ok && args.Seq <= last {
		reply.OK = ls.lastReply[args.CID]
		return nil
	}
	ls.lastSeq[args.CID] = args.Seq

	locked, _ := ls.locks[args.Lockname]

	if locked {
		ls.locks[args.Lockname] = false
		reply.OK = true
	} else {
		reply.OK = false
	}
	ls.lastReply[args.CID] = reply.OK
	return nil
}

//
// tell the server to shut itself down.
// for testing.
// please don't change this.
//
func (ls *LockServer) kill() {
	ls.dead = true
	ls.l.Close()
}

//
// hack to allow test_test.go to have primary process
// an RPC but not send a reply. can't use the shutdown()
// trick b/c that causes client to immediately get an
// error and send to backup before primary does.
// please don't change anything to do with DeafConn.
//
type DeafConn struct {
	c io.ReadWriteCloser
}

func (dc DeafConn) Write(p []byte) (n int, err error) {
	return len(p), nil
}
func (dc DeafConn) Close() error {
	return dc.c.Close()
}
func (dc DeafConn) Read(p []byte) (n int, err error) {
	return dc.c.Read(p)
}

func StartServer(primary string) *LockServer {
	ls := new(LockServer)
	ls.locks = map[uint64]bool{}

	ls.lastSeq = map[uint64]uint64{}
	ls.lastReply = map[uint64]bool{}

	me := primary

	// tell net/rpc about our RPC server and handlers.
	rpcs := rpc.NewServer()
	rpcs.Register(ls)

	// prepare to receive connections from clients.
	// change "unix" to "tcp" to use over a network.
	os.Remove(me) // only needed for "unix"
	l, e := net.Listen("unix", me)
	if e != nil {
		log.Fatal("listen error: ", e)
	}
	ls.l = l

	// please don't change any of the following code,
	// or do anything to subvert it.

	// create a thread to accept RPC connections from clients.
	go func() {
		for ls.dead == false {
			conn, err := ls.l.Accept()
			if err == nil && ls.dead == false {
				if ls.dying {
					// process the request but force discard of reply.

					// without this the connection is never closed,
					// b/c ServeConn() is waiting for more requests.
					// test_test.go depends on this two seconds.
					go func() {
						time.Sleep(2 * time.Second)
						conn.Close()
					}()
					ls.l.Close()

					// this object has the type ServeConn expects,
					// but discards writes (i.e. discards the RPC reply).
					deaf_conn := DeafConn{c: conn}

					rpcs.ServeConn(deaf_conn)

					ls.dead = true
				} else {
					go rpcs.ServeConn(conn)
				}
			} else if err == nil {
				conn.Close()
			}
			if err != nil && ls.dead == false {
				fmt.Printf("LockServer(%v) accept: %v\n", me, err.Error())
				ls.kill()
			}
		}
	}()

	return ls
}
