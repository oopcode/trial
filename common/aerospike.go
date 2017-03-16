package common

import (
	"log"
	"sync"
	"time"

	as "github.com/aerospike/aerospike-client-go"
)

// ASConn safely wraps and maintains an Aerospike connection.
type ASConn struct {
	sync.Mutex
	host   string
	port   int
	conn   *as.Client
	policy *as.ClientPolicy
}

// NewASConn is a constructor for ASConn.
func NewASConn(host string, port int) *ASConn {
	out := &ASConn{
		host: host,
		port: port,
	}
	out.policy = as.NewClientPolicy()
	out.policy.Timeout = time.Millisecond * 10
	return out
}

// Close wraps as.Client.Close().
func (c *ASConn) Close() {
	c.Lock()
	defer c.Unlock()
	c.conn.Close()
}

// PutBins wraps as.Client.PutBins(). If Aerospike is unreachable, PutBins
// will try to reconnect; it will return an error if something goes wrong with
// that.
func (c *ASConn) PutBins(policy *as.WritePolicy, key *as.Key,
	bins ...*as.Bin) error {
	c.Lock()
	defer c.Unlock()
	if !c.isConnected() {
		// Try to reconnect.
		conn, err := as.NewClientWithPolicy(c.policy, c.host, c.port)
		if err != nil {
			log.Printf("Failed to connect to aerospike; %v", err)
			return err
		}
		c.conn = conn
	}
	return c.conn.PutBins(policy, key, bins...)
}

func (c *ASConn) isConnected() bool {
	return c.conn != nil && c.conn.IsConnected()
}
