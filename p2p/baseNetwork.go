package p2p

import (
	"context"
	"math"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yrdsm666/tockowl/config"
	"github.com/yrdsm666/tockowl/pb"
	"github.com/yrdsm666/tockowl/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
)

type BaseNetwork struct {
	nodes       []*config.ReplicaInfo
	log         *logrus.Entry
	id          int64
	output      chan<- []byte
	rpcServer   *grpc.Server
	partition   string
	networkType int64
	extraDelay  int
	context     context.Context
	cancel      context.CancelFunc
	sendFailed  bool
	pb.UnimplementedP2PServer
}

func NewBaseNetwork(
	log *logrus.Entry,
	id int64,
	networkType int64,
	partition string,
) (*BaseNetwork, string, error) {
	cfg, err := config.NewConfig("config/config.yaml", id, partition)
	utils.PanicOnError(err)
	var port string
	if id == -1 {
		// undernode for leopard
		port = ":1121"
	} else {
		info := cfg.GetNodeInfo(id)
		port = info.Address[strings.Index(info.Address, ":"):]
	}
	listen, err := net.Listen("tcp", port)
	utils.PanicOnError(err)
	rpcServer := grpc.NewServer(grpc.MaxRecvMsgSize(1024 * 1024 * 1000)) // MAX: 1000 MB
	utils.PanicOnError(err)
	ctx, cancel := context.WithCancel(context.Background())
	bp := &BaseNetwork{
		nodes:                  cfg.Nodes,
		log:                    log,
		id:                     id,
		output:                 nil,
		rpcServer:              rpcServer,
		UnimplementedP2PServer: pb.UnimplementedP2PServer{},
		networkType:            networkType,
		extraDelay:             cfg.P2P.ExtraDelay,
		partition:              partition,
		context:                ctx,
		cancel:                 cancel,
		sendFailed:             false,
	}
	pb.RegisterP2PServer(rpcServer, bp)
	log.Infof("[BaseNetwork] Server start at %s", listen.Addr().String())
	go rpcServer.Serve(listen)
	var nid string
	if id == -1 {
		nid = "127.0.0.1:1121"
	} else {
		nid = cfg.Nodes[id].Address
	}
	// nid := cfg.Nodes[id].Address
	return bp, nid, nil
}

func (bp *BaseNetwork) SetReceiver(receiver MsgReceiver) {
	bp.output = receiver.GetMsgByteEntrance()
}

func (bp *BaseNetwork) Subscribe(topic []byte) error {
	// do nothing
	return nil
}

func (bp *BaseNetwork) UnSubscribe(topic []byte) error {
	// do nothing
	return nil
}

func (bp *BaseNetwork) GetPeerID() string {
	// do nothing
	return ""
}

func (bp *BaseNetwork) GetP2PType() string {
	// do nothing
	return "p2p"
}

func (bp *BaseNetwork) Broadcast(msgByte []byte, consensusID int64, topic []byte) error {
	nodes := bp.nodes
	// bp.log.Errorf("broadcast packet to %d nodes", len(nodes))

	for _, node := range nodes {
		if node.ID == bp.id {
			continue
		}
		err := bp.Unicast(node.Address, msgByte, consensusID, topic)
		if err != nil {
			bp.log.WithError(err).Warn("send msg failed")
		}
		// bp.log.Debugf("unicast to %s done", node.Address)
	}
	return nil
}

func (bp *BaseNetwork) Unicast(address string, msgByte []byte, consensusID int64, topic []byte) error {
	// conn, err := grpc.Dial(address, grpc.WithInsecure())
	// bp.log.Trace("[BaseNetwork] unicast")

	go func() {
		delay := bp.messageDelay()
		time.Sleep(time.Duration(delay) * time.Millisecond)
		conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			bp.log.WithError(err).Warn("dial ", address, " failed")
			return
		}
		client := pb.NewP2PClient(conn)
		packet := new(pb.Packet)
		err = proto.Unmarshal(msgByte, packet)
		if err != nil {
			bp.log.WithField("address", address).WithError(err).Warn("Unmarshal msgByte to packet failed")
		}
		utils.PanicOnError(err)
		// bp.log.Debug("unicast packet to", address)
		_, err = client.Send(context.Background(), packet)
		if err != nil {
			if !bp.sendFailed && bp.id == 1 {
				// Make sure the error log is only printed once
				bp.log.WithError(err).Warn("send to ", address, " failed")
				bp.sendFailed = true
			}
			return
		}
		conn.Close()
	}()
	return nil
}

// BaseNetwork implements P2PServer
func (bp *BaseNetwork) Send(ctx context.Context, in *pb.Packet) (*pb.Empty, error) {
	select {
	case <-bp.context.Done():
		return &pb.Empty{}, nil
	default:
		bytePacket, err := proto.Marshal(in)
		// bp.log.Warn("[BaseNetwork] received msg:", in)
		if err != nil {
			bp.log.Warn("marshal packet failed")
			return nil, err
		}
		if bp.output != nil {
			// if bp.partition == in.Partition {
			bp.output <- bytePacket
			// } else {
			// 	bp.log.WithFields(logrus.Fields{
			// 		"bp.partition": bp.partition,
			// 		"in.Partition": in.Partition,
			// 		"in":           in,
			// 	}).Warn("[BaseNetwork] error partition")
			// }
		} else {
			bp.log.Warn("[BaseNetwork] output nil")
		}
		return &pb.Empty{}, nil
	}
}

func (bp *BaseNetwork) Stop() {
	bp.cancel()
	bp.rpcServer.Stop()
}

func (bp *BaseNetwork) Request(ctx context.Context, in *pb.HeaderRequest) (*pb.Header, error) {
	id := in.Address
	address := bp.nodes[id].RpcAddress
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		bp.log.Warn("dial", address, "failed")
		return nil, err
	}
	client := pb.NewP2PClient(conn)
	header, err := client.Request(context.Background(), in)
	if err != nil {
		bp.log.Warn("[BaseNetwork]\t sent request to ", address, "fail: ", err)
		return nil, err
	}
	conn.Close()
	return header, nil
}

func (bp *BaseNetwork) timeDelay() (delay int) {
	if bp.networkType == -1 {
		return
	} else if bp.networkType == 0 {
		delay = rand.Intn(100)
	} else if bp.networkType == 1 {
		delay = rand.Intn(120) + int(math.Abs(rand.NormFloat64()*60))
	} else {
		delay = int(rand.ExpFloat64() * 200)
	}
	return
}

func (bp *BaseNetwork) messageDelay() (delay int) {
	if bp.id == 0 {
		return bp.extraDelay
	} else {
		return 0
	}
}
