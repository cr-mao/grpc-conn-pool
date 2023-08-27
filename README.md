## go grpc 连接池

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Reference](https://pkg.go.dev/badge/github.com/cr-mao/grpc-conn-pool.svg)](https://pkg.go.dev/github.com/cr-mao/grpc-conn-pool)


golang grpc client pool 


```shell
go get  github.com/cr-mao/grpc-conn-pool@v1.0.0
```

### 特点
- 支持多target，pool. 
- 连接检测，失效，才进行替换连接
- 获得一次连接耗时 90 op/ns 左右（15年mac pro)


### 使用


```go
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
		// 10000次 只要7ms.
		cost := time.Now().Sub(begin).Nanoseconds()
		t.Log(cost)
		t.Logf("%d ns/op", cost/1000000) // 730 op/ns
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
```
