/**
User: cr-mao
Date: 2023/8/27 08:53
Email: crmao@qq.com
Desc: pool.go
*/
package grpc_conn_pool

import (
	"errors"
	"google.golang.org/grpc/credentials/insecure"
	"sync"
	"sync/atomic"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

var ErrConn = errors.New("grpc conn has error Shutdown or TransientFailure")

const (
	defaultPoolSize = 10
)

type ClientBuilder struct {
	opts     *Options
	dialOpts []grpc.DialOption
	pools    sync.Map
}

type Options struct {
	PoolSize uint32
	DialOpts []grpc.DialOption
}

func NewClientBuilder(opts *Options) *ClientBuilder {
	// grpc 要求必须传这个，没传则帮助带上
	if len(opts.DialOpts) == 0 {
		opts.DialOpts = []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}
	}
	b := &ClientBuilder{
		opts:     opts,
		dialOpts: opts.DialOpts,
	}
	return b
}

// Build 构建连接
func (b *ClientBuilder) GetConn(target string) (*grpc.ClientConn, error) {
	val, ok := b.pools.Load(target)
	if ok {
		return val.(*Pool).Get()
	}
	size := b.opts.PoolSize
	if size <= 0 {
		size = defaultPoolSize
	}
	pool, err := newPool(size, target, b.dialOpts...)
	if err != nil {
		return nil, err
	}
	b.pools.Store(target, pool)
	return pool.Get()
}

type Pool struct {
	target   string
	poolSize uint64
	index    uint64
	conns    []*grpc.ClientConn
	opts     []grpc.DialOption
	sync.Mutex
}

func newPool(poolSize uint32, target string, opts ...grpc.DialOption) (*Pool, error) {

	p := &Pool{
		poolSize: uint64(poolSize),
		conns:    make([]*grpc.ClientConn, poolSize),
		target:   target,
		opts:     opts,
	}
	size := int(poolSize)
	for i := 0; i < size; i++ {
		conn, err := grpc.Dial(target, opts...)
		if err != nil {
			return nil, err
		}
		p.conns[i] = conn
	}
	return p, nil
}

func (p *Pool) Get() (*grpc.ClientConn, error) {
	idx := int(atomic.AddUint64(&p.index, 1) % p.poolSize)
	conn := p.conns[idx]
	if conn != nil && p.checkState(conn) == nil {
		return conn, nil
	}
	// gc old conn
	if conn != nil {
		_ = conn.Close()
	}
	p.Lock()
	defer p.Unlock()
	// double check, already inited
	conn = p.conns[idx]
	if conn != nil && p.checkState(conn) == nil {
		return conn, nil
	}
	conn, err := grpc.Dial(p.target, p.opts...)
	if err != nil {
		return nil, err
	}
	p.conns[idx] = conn
	return conn, nil
}

// 检测连接是否正常
func (p *Pool) checkState(conn *grpc.ClientConn) error {
	state := conn.GetState()
	switch state {
	case connectivity.TransientFailure, connectivity.Shutdown:
		return ErrConn
	}
	return nil
}

// 检测连接是否正常
func CheckState(conn *grpc.ClientConn) error {
	state := conn.GetState()
	switch state {
	case connectivity.TransientFailure, connectivity.Shutdown:
		return ErrConn
	}
	return nil
}
