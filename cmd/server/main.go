// Copyright 2018 The Prizem Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/nats-io/gnatsd/server"
	nats "github.com/nats-io/go-nats"
	"github.com/prizem-io/api/v1"
	"github.com/prizem-io/api/v1/proto"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/prizem-io/control-plane/pkg/app"
	"github.com/prizem-io/control-plane/pkg/log"
	replication "github.com/prizem-io/control-plane/pkg/replication/nats"
	"github.com/prizem-io/control-plane/pkg/store/caching"
	"github.com/prizem-io/control-plane/pkg/store/logging"
	"github.com/prizem-io/control-plane/pkg/store/postgres"
	grpctransport "github.com/prizem-io/control-plane/pkg/transport/grpc"
	resttransport "github.com/prizem-io/control-plane/pkg/transport/rest"
)

func main() {
	zapLogger, _ := zap.NewProduction()
	defer zapLogger.Sync() // flushes buffer, if any
	sugar := zapLogger.Sugar()
	logger := log.New(sugar)

	serverID := uuid.NewV4()

	// Load the application config
	logger.Info("Loading configuration...")
	config, err := app.LoadConfig()
	if err != nil {
		logger.Fatal(err)
	}

	// Connect to database
	logger.Info("Connecting to database...")
	var db *sqlx.DB
	err = backoff.Retry(func() (err error) {
		db, err = app.ConnectDB(&config.Database)
		return
	}, backoff.NewExponentialBackOff())
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	// Connect to NATS
	logger.Info("Connecting to NATS...")
	s := runNATSServer()
	defer s.Shutdown()
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		logger.Fatal(err)
	}

	store := postgres.New(db)

	var routes api.Routes
	routesReplicator := replication.NewRoutes(logger, serverID.String(), nc)
	err = routesReplicator.Subscribe()
	if err != nil {
		logger.Fatal(err)
	}
	defer routesReplicator.Unsubscribe()
	routes = store
	routes = logging.NewRouting(routes, logger)
	routesCache := caching.NewRoutes(routes)
	routesCache.AddCallback(routesReplicator.PublishRoutes)
	routesReplicator.AddCallback(routesCache.SetServices)

	var endpoints api.Endpoints
	endpointsReplicator := replication.NewEndpoints(logger, serverID.String(), nc)
	err = endpointsReplicator.Subscribe()
	if err != nil {
		logger.Fatal(err)
	}
	defer endpointsReplicator.Unsubscribe()
	endpoints = store
	endpoints = logging.NewEndpoints(endpoints, logger)
	endpointsCache := caching.NewEndpoints(endpoints)
	endpointsCache.AddCallback(endpointsReplicator.PublishEndpoints)
	endpointsReplicator.AddCallback(endpointsCache.SetEndpoints)

	// Mechanical domain.
	errc := make(chan error)

	// Interrupt handler.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	// HTTP transport.
	go func() {
		srv := resttransport.NewServer()
		srv.RegisterRoutes(routesCache)
		srv.RegisterEndpoints(endpointsCache)

		errc <- http.ListenAndServe(":8000", srv.Handler())
	}()

	// gRPC transport.
	go func() {
		ln, err := net.Listen("tcp", ":9000")
		if err != nil {
			errc <- err
			return
		}
		s := grpc.NewServer() //grpc.Creds(creds))

		routesServer := grpctransport.NewRoutes(logger, routesCache)
		routesCache.AddCallback(routesServer.PublishRoutes)
		proto.RegisterRouteDiscoveryServer(s, routesServer)

		endpointsServer := grpctransport.NewEndpoints(logger, endpointsCache)
		endpointsCache.AddCallback(endpointsServer.PublishEndpoints)
		proto.RegisterEndpointDiscoveryServer(s, endpointsServer)

		errc <- s.Serve(ln)
	}()

	logger.Info("Control plane started")
	<-errc
}

// runNATSServer starts a new Go routine based NATS server
func runNATSServer() *server.Server {
	s := server.New(&server.Options{
		Port:           4222,
		NoLog:          true,
		NoSigs:         true,
		MaxControlLine: 256,
		Cluster: server.ClusterOpts{
			Username: "foo",
			Password: "bar",
			Port:     4248,
		},
		HTTPPort: 8222,
	})
	if s == nil {
		panic("No NATS Server object returned.")
	}

	// Run server in Go routine.
	go s.Start()

	// Wait for accept loop(s) to be started
	if !s.ReadyForConnections(10 * time.Second) {
		panic("Unable to start NATS Server in Go Routine")
	}
	return s
}
