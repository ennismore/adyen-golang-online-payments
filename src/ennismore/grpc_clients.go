package ennismore

import (
	"context"
	"github.com/ennismore/em-domain/v2/pkg/grpc"
	"github.com/ennismore/em-domain/v2/service/api/booking"
	"github.com/ennismore/em-domain/v2/service/bookingrepository"
	"github.com/ennismore/em-domain/v2/service/config"
	"github.com/ennismore/em-domain/v2/service/id"
	"github.com/ennismore/em-domain/v2/service/userrepository"
	"time"
)

type Clients struct {
	booking.BookingServiceClient
	id.IdObfuscatorClient
	config.ConfigServiceClient
	bookingrepository.BookingRepositoryClient
	userrepository.UserAccountClient
	timeout         time.Duration
	grpcConnections []*grpc.Connection
	stopped         bool
}

func (g *Clients) ConnectAll(ctx context.Context, apiAddr, idObfuscatorAddr, configAddr, repoAddr, userRepoAddr string) {
	apiConn := &grpc.Connection{Timeout: g.timeout, Add: apiAddr}
	apiConn.Connect(ctx)
	g.BookingServiceClient = booking.NewBookingServiceClient(apiConn.Connection)
	g.grpcConnections = append(g.grpcConnections, apiConn)

	idConn := &grpc.Connection{Timeout: g.timeout, Add: idObfuscatorAddr}
	idConn.Connect(ctx)
	g.IdObfuscatorClient = id.NewIdObfuscatorClient(idConn.Connection)
	g.grpcConnections = append(g.grpcConnections, idConn)

	confConn := &grpc.Connection{Timeout: g.timeout, Add: configAddr}
	confConn.Connect(ctx)
	g.ConfigServiceClient = config.NewConfigServiceClient(confConn.Connection)
	g.grpcConnections = append(g.grpcConnections, confConn)

	repoConn := &grpc.Connection{Timeout: g.timeout, Add: repoAddr}
	repoConn.Connect(ctx)
	g.BookingRepositoryClient = bookingrepository.NewBookingRepositoryClient(repoConn.Connection)
	g.grpcConnections = append(g.grpcConnections, repoConn)

	userConn := &grpc.Connection{Timeout: g.timeout, Add: userRepoAddr}
	userConn.Connect(ctx)
	g.UserAccountClient = userrepository.NewUserAccountClient(userConn.Connection)
	g.grpcConnections = append(g.grpcConnections, userConn)
}

func (g *Clients) StopAll() {
	for _, f := range g.grpcConnections {
		f.Stop()
	}
	g.stopped = true
}
