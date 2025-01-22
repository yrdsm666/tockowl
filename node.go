package tockowl

import (
	"context"
	"net"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/yrdsm666/tockowl/config"
	"github.com/yrdsm666/tockowl/consensus"
	"github.com/yrdsm666/tockowl/consensus/model"
	"github.com/yrdsm666/tockowl/crypto"
	"github.com/yrdsm666/tockowl/executor"
	"github.com/yrdsm666/tockowl/logging"
	"github.com/yrdsm666/tockowl/p2p"
	"github.com/yrdsm666/tockowl/pb"
	"github.com/yrdsm666/tockowl/utils"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type Node struct {
	id         int64
	consensus  model.Consensus
	rpcServer  *grpc.Server
	log        *logrus.Entry
	p2pAdaptor p2p.P2PAdaptor
	pb.UnimplementedP2PServer
}

func NewNode(id int64, keySet *crypto.KeySet, partition string, upd func(int64)) *Node {
	cfg, err := config.NewConfig("config/config.yaml", id, partition)
	cfg.Consensus.Keys = keySet
	utils.PanicOnError(err)
	info := cfg.GetNodeInfo(id)
	log := logging.GetLogger().WithField("id", id)
	port := info.RpcAddress[strings.Index(info.RpcAddress, ":"):]
	listen, err := net.Listen("tcp", port)
	utils.PanicOnError(err)
	rpcServer := grpc.NewServer()
	utils.PanicOnError(err)

	// p, nid, err := p2p.NewBaseP2p(log, id, int64(cfg.NetworkType))
	p, nid, err := p2p.BuildP2P(cfg.P2P, log, id, int64(cfg.NetworkType), partition)
	if p == nil {
		panic("build p2p failed")
	}
	log.WithFields(logrus.Fields{
		"nid":  nid,
		"type": p.GetP2PType(),
	}).Info("get network id")
	utils.PanicOnError(err)
	e := executor.BuildExecutor(cfg.Executor, log)
	c := consensus.BuildConsensus(id, -1, cfg.Consensus, e, p, log, partition, upd)
	if c == nil {
		panic("Initialize consensus failed")
	}
	node := &Node{
		id:                     id,
		consensus:              c,
		rpcServer:              rpcServer,
		p2pAdaptor:             p,
		log:                    log,
		UnimplementedP2PServer: pb.UnimplementedP2PServer{},
	}
	pb.RegisterP2PServer(rpcServer, node)
	log.Infof("[Node] Server start at %s", listen.Addr().String())
	go rpcServer.Serve(listen)
	return node
}

func NewNodeForRemoteExperiment(id int64, keySet *crypto.KeySet, partition string, upd func(int64)) *Node {
	cfg, err := config.NewConfig("config/config.yaml", id, partition)
	cfg.Consensus.Keys = keySet
	utils.PanicOnError(err)

	log := logging.GetLogger().WithField("id", id)

	// p, nid, err := p2p.NewBaseP2p(log, id, int64(cfg.NetworkType))
	p, nid, err := p2p.BuildP2P(cfg.P2P, log, id, int64(cfg.NetworkType), partition)
	if p == nil {
		panic("build p2p failed")
	}
	log.WithFields(logrus.Fields{
		"nid":  nid,
		"type": p.GetP2PType(),
	}).Info("get network id")
	utils.PanicOnError(err)
	e := executor.BuildExecutor(cfg.Executor, log)
	c := consensus.BuildConsensus(id, -1, cfg.Consensus, e, p, log, partition, upd)
	if c == nil {
		panic("Initialize consensus failed")
	}
	node := &Node{
		id:                     id,
		consensus:              c,
		rpcServer:              nil,
		p2pAdaptor:             p,
		log:                    log,
		UnimplementedP2PServer: pb.UnimplementedP2PServer{},
	}
	return node
}

func (node *Node) Stop() {
	if node.rpcServer != nil {
		node.rpcServer.Stop()
	}

	node.consensus.Stop()
	node.p2pAdaptor.Stop()
}

// Node implements P2PServer
func (node *Node) Send(ctx context.Context, in *pb.Packet) (*pb.Empty, error) {
	if in.Type != pb.PacketType_CLIENTPACKET {
		node.log.Warn("[Node] packet type error")
		return &pb.Empty{}, nil
	}
	request := new(pb.Request)
	if err := proto.Unmarshal(in.Msg, request); err != nil {
		node.log.WithError(err).Warn("[Node] decode request error")
		return &pb.Empty{}, nil
	}
	node.consensus.GetRequestEntrance() <- request
	return &pb.Empty{}, nil
}
