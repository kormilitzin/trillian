// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"flag"
	"fmt"

	_ "net/http/pprof"

	_ "github.com/go-sql-driver/mysql" // Load MySQL driver

	"github.com/golang/glog"
	"github.com/google/trillian"
	"github.com/google/trillian/crypto/keys"
	"github.com/google/trillian/extension"
	"github.com/google/trillian/server"
	"github.com/google/trillian/server/interceptor"
	"github.com/google/trillian/server/vmap"
	"github.com/google/trillian/storage/mysql"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
)

var (
	mySQLURI       = flag.String("mysql_uri", "test:zaphod@tcp(127.0.0.1:3306)/test", "Connection URI for MySQL database")
	serverPortFlag = flag.Int("port", 8090, "Port to serve log RPC requests on")
	httpPortFlag   = flag.Int("http_port", 8091, "Port to serve HTTP metrics and REST requests on (negative means disabled)")
)

func main() {
	flag.Parse()

	db, err := mysql.OpenDB(*mySQLURI)
	if err != nil {
		glog.Exitf("Failed to open database: %v", err)
	}
	// No defer: database ownership is delegated to server.Main

	registry := extension.Registry{
		AdminStorage:  mysql.NewAdminStorage(db),
		SignerFactory: keys.PEMSignerFactory{},
		MapStorage:    mysql.NewMapStorage(db),
	}

	ti := interceptor.TreeInterceptor{Admin: registry.AdminStorage}
	s := grpc.NewServer(grpc.UnaryInterceptor(interceptor.WrapErrors(ti.UnaryInterceptor)))
	// No defer: server ownership is delegated to server.Main

	httpEndpoint := ""
	if *httpPortFlag >= 0 {
		httpEndpoint = fmt.Sprintf("localhost:%v", *httpPortFlag)
	}

	m := server.Main{
		RPCEndpoint:  fmt.Sprintf("localhost:%v", *serverPortFlag),
		HTTPEndpoint: httpEndpoint,
		DB:           db,
		Registry:     registry,
		Server:       s,
		RegisterHandlerFn: func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error {
			return nil
		},
		RegisterServerFn: func(s *grpc.Server, registry extension.Registry) error {
			mapServer := vmap.NewTrillianMapServer(registry)
			if err := mapServer.IsHealthy(); err != nil {
				return err
			}
			trillian.RegisterTrillianMapServer(s, mapServer)
			return err
		},
	}

	ctx := context.Background()
	if err := m.Run(ctx); err != nil {
		glog.Exitf("Server exited with error: %v", err)
	}
}
