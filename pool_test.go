/**
User: cr-mao
Date: 2023/8/27 09:30
Email: crmao@qq.com
Desc: pool_test.go
*/
package grpc_conn_pool_test

import (
	"net"
	"testing"
	"time"

	"github.com/cr-mao/grpc-conn-pool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Test_GrpcConnPool(t *testing.T) {

	target := "127.0.0.1:13688"
	go func() {
		clientBuilder := grpc_conn_pool.NewClientBuilder(&grpc_conn_pool.Options{
			PoolSize: 20,
			DialOpts: []grpc.DialOption{
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			},
		})

		i := 0
		begin := time.Now()
		for {
			conn, err := clientBuilder.GetConn(target)
			if err != nil {
				t.Errorf("get conn err %v", err)
				return
			}
			err = grpc_conn_pool.CheckState(conn)
			if err != nil {
				t.Log(conn.GetState())
			}
			//time.Sleep(time.Millisecond * 2)
			//t.Log("ok")
			i++
			if i > 1000000 {
				break
			}
		}
		cost := time.Now().Sub(begin).Nanoseconds()
		t.Log(cost)
		t.Logf("%d ns/op", cost/1000000)
	}()

	addr := "0.0.0.0:13688"

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatalf("Listen err :%v", err)
	}
	t.Logf("grpc serve at: %s", addr)
	server := grpc.NewServer()
	_ = server.Serve(lis)

}
