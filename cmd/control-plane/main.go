// Copyright 2018 The Prizem Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/nats-io/gnatsd/server"
	"github.com/nats-io/go-nats"
	"github.com/oklog/run"
	"github.com/prizem-io/api/v1"
	"github.com/prizem-io/api/v1/proto"
	"github.com/satori/go.uuid"
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

	httpListenPort := readEnvInt("HTTP_PORT", 8000)
	grpcListenPort := readEnvInt("GRPC_PORT", 9000)
	disableEmbeddedNATS := readEnvBool("DISABLE_EMBEDDED_NATS", false)

	flag.IntVar(&httpListenPort, "httpPort", httpListenPort, "The HTTP listening port")
	flag.IntVar(&grpcListenPort, "grpcPort", grpcListenPort, "The gRPC listening port")
	flag.BoolVar(&disableEmbeddedNATS, "disableEmbeddedNATS", disableEmbeddedNATS, "Disable running NATS as an embedded server")
	flag.Parse()

	// Load the application config
	logger.Info("Loading configuration...")
	config, err := app.LoadConfig()
	if err != nil {
		logger.Fatal(err)
	}

	eb := backoff.NewExponentialBackOff()
	notify := func(err error, d time.Duration) {
		logger.Infof("Failed attempt: %v -> will retry in %s", err, d)
	}

	// Connect to database
	logger.Info("Connecting to database...")
	var db *sqlx.DB
	err = backoff.RetryNotify(func() (err error) {
		db, err = app.ConnectDB(&config.Database)
		return
	}, eb, notify)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	// Connect to NATS
	logger.Info("Connecting to NATS...")
	if !disableEmbeddedNATS {
		s := runNATSServer()
		defer s.Shutdown()
	}

	var nc *nats.Conn
	eb.Reset()
	err = backoff.RetryNotify(func() (err error) {
		nc, err = nats.Connect(nats.DefaultURL)
		return
	}, eb, notify)
	if err != nil {
		logger.Fatal(err)
	}

	// Using Postgres as the backend store
	store := postgres.New(db)

	// Initialize routes service
	eb.Reset()
	var routes api.Routes
	routesReplicator := replication.NewRoutes(logger, serverID.String(), nc)
	err = backoff.RetryNotify(routesReplicator.Subscribe, eb, notify)
	if err != nil {
		logger.Fatal(err)
	}
	defer routesReplicator.Unsubscribe()
	routes = store
	routes = logging.NewRouting(routes, logger)
	routesCache := caching.NewRoutes(routes)
	routesCache.AddCallback(routesReplicator.PublishRoutes)
	routesReplicator.AddCallback(routesCache.SetServices)
	routes = routesCache

	// Initialize endpoints service
	eb.Reset()
	var endpoints api.Endpoints
	endpointsReplicator := replication.NewEndpoints(logger, serverID.String(), nc)
	err = backoff.RetryNotify(endpointsReplicator.Subscribe, eb, notify)
	if err != nil {
		logger.Fatal(err)
	}
	defer endpointsReplicator.Unsubscribe()
	endpoints = store
	endpoints = logging.NewEndpoints(endpoints, logger)
	endpointsCache := caching.NewEndpoints(endpoints)
	endpointsCache.AddCallback(endpointsReplicator.PublishEndpoints)
	endpointsReplicator.AddCallback(endpointsCache.SetEndpoints)
	endpoints = endpointsCache

	var g run.Group

	// HTTP transport.
	{
		var listener net.Listener
		g.Add(func() error {
			var err error
			logger.Infof("HTTP listener starting on :%d", httpListenPort)
			listener, err = net.Listen("tcp", fmt.Sprintf(":%d", httpListenPort))
			if err != nil {
				return err
			}
			srv := resttransport.NewServer()
			srv.RegisterRoutes(routes)
			srv.RegisterEndpoints(endpoints)

			return http.Serve(listener, srv.Handler())
		}, func(error) {
			if listener != nil {
				listener.Close()
			}
		})
	}
	// gRPC transport.
	{
		var listener net.Listener
		g.Add(func() error {
			var err error
			logger.Infof("gRPC listener starting on :%d", grpcListenPort)
			listener, err = net.Listen("tcp", fmt.Sprintf(":%d", grpcListenPort))
			if err != nil {
				return err
			}
			s := grpc.NewServer()

			routesServer := grpctransport.NewRoutes(logger, routes)
			routesCache.AddCallback(routesServer.PublishRoutes)
			proto.RegisterRouteDiscoveryServer(s, routesServer)

			endpointsServer := grpctransport.NewEndpoints(logger, endpoints)
			endpointsCache.AddCallback(endpointsServer.PublishEndpoints)
			proto.RegisterEndpointDiscoveryServer(s, endpointsServer)

			return s.Serve(listener)
		}, func(error) {
			if listener != nil {
				listener.Close()
			}
		})
	}
	// This function just sits and waits for ctrl-C.
	{
		cancelInterrupt := make(chan struct{})
		g.Add(func() error {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
			select {
			case sig := <-c:
				return fmt.Errorf("received signal %s", sig)
			case <-cancelInterrupt:
				return nil
			}
		}, func(error) {
			close(cancelInterrupt)
		})
	}

	logger.Info("Control plane started")
	logger.Infof("exit %v", g.Run())
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

func readEnvInt(key string, defaultValue int) int {
	if i, err := strconv.Atoi(os.Getenv(key)); err == nil {
		return i
	}
	return defaultValue
}

func readEnvBool(key string, defaultValue bool) bool {
	if i, err := strconv.ParseBool(os.Getenv(key)); err == nil {
		return i
	}
	return defaultValue
}
