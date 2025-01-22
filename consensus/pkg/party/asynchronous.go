package party

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/yrdsm666/tockowl/config"
	"github.com/yrdsm666/tockowl/consensus/pkg/protobuf"
	"github.com/yrdsm666/tockowl/executor"
	"github.com/yrdsm666/tockowl/p2p"
	pb "github.com/yrdsm666/tockowl/pb"
	"github.com/yrdsm666/tockowl/types"
	"github.com/yrdsm666/tockowl/utils"
	"google.golang.org/protobuf/proto"
)

type Asynchronous interface {
	//Msg(msgType pb.MsgType, id int, round int, sid int, proposal []byte, signatureByte []byte) *pb.Msg
	GetMsgByteEntrance() chan<- []byte
	GetRequestEntrance() chan<- *pb.Request
	GetNetworkInfo() map[uint32]string
	GetSelfInfo() *config.ReplicaInfo
	SafeExit()
	Broadcast(msg *pb.Msg) error
	Unicast(address string, msg *pb.Msg) error
	ProcessProposal(cmds []string) error
}

type AsynchronousImpl struct {
	ID              int64
	ConsensusID     int64
	Config          *config.ConsensusConfig
	MemPool         *types.MemPool
	MsgByteEntrance chan []byte // receive msg
	RequestEntrance chan *pb.Request
	p2pAdaptor      p2p.P2PAdaptor
	Log             *logrus.Entry
	Executor        executor.Executor
	Metric          utils.Metric
	cancel          context.CancelFunc
	partitionName   string
}

func (a *AsynchronousImpl) Init(
	id int64,
	cid int64,
	cfg *config.ConsensusConfig,
	exec executor.Executor,
	p2pAdaptor p2p.P2PAdaptor,
	log *logrus.Entry,
	cancel context.CancelFunc,
	partition string,
) {
	a.ID = id
	a.ConsensusID = cid
	a.Config = cfg
	a.Executor = exec
	a.p2pAdaptor = p2pAdaptor
	a.Log = log.WithField("consensus id", cid)
	a.cancel = cancel

	a.partitionName = partition
	a.MsgByteEntrance = make(chan []byte, 10)
	a.RequestEntrance = make(chan *pb.Request, 10)

	a.MemPool = types.NewMemPool()
	a.Metric = *utils.NewMetric(cfg.BatchSize)

	if p2pAdaptor != nil {
		p2pAdaptor.SetReceiver(a)
		err := p2pAdaptor.Subscribe([]byte(partition))
		if err != nil {
			a.Log.Error("Subscribe error: ", err.Error())
			return
		}
	} else {
		a.Log.Warn("p2p is nil")
	}

	a.Log.Info("Joined to topic: ", partition)
}

func (a *AsynchronousImpl) GetConsensusID() int64 {
	return a.ConsensusID
}

func (a *AsynchronousImpl) DecodeMsgByte(msgByte []byte) (*protobuf.Message, error) {
	msg := new(protobuf.Message)
	err := proto.Unmarshal(msgByte, msg)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (a *AsynchronousImpl) SafeExit() {
	// close(a.MsgByteEntrance)
	close(a.RequestEntrance)
}

func (a *AsynchronousImpl) GetMsgByteEntrance() chan<- []byte {
	return a.MsgByteEntrance
}

func (a *AsynchronousImpl) GetRequestEntrance() chan<- *pb.Request {
	return a.RequestEntrance
}

func (a *AsynchronousImpl) GetSelfInfo() *config.ReplicaInfo {
	self := &config.ReplicaInfo{}
	for _, info := range a.Config.Nodes {
		if info.ID == a.ID {
			self = info
			break
		}
	}
	return self
}

func (a *AsynchronousImpl) GetNetworkInfo() map[int64]string {
	networkInfo := make(map[int64]string)
	for _, info := range a.Config.Nodes {
		if info.ID == a.ID {
			continue
		}
		networkInfo[info.ID] = info.Address
	}
	return networkInfo
}

// Broadcast objects include the sender itself
func (a *AsynchronousImpl) Broadcast(msg *protobuf.Message) error {
	if a.p2pAdaptor == nil {
		a.Log.Warn("p2pAdaptor nil")
		return nil
	}

	msgByte, err := proto.Marshal(msg)
	utils.PanicOnError(err)

	// Send a message to the node itself
	a.GetMsgByteEntrance() <- msgByte

	return a.p2pAdaptor.Broadcast(msgByte, a.ConsensusID, []byte("consensus"))
}

func (h *AsynchronousImpl) Unicast(address string, msg *protobuf.Message) error {
	if h.p2pAdaptor == nil {
		h.Log.Warn("p2pAdaptor nil")
		return nil
	}

	// if address == "" {
	// 	h.Log.Warn("msg: ", msg)
	// 	return errors.New("invalid address")
	// }

	msgByte, err := proto.Marshal(msg)
	utils.PanicOnError(err)
	return h.p2pAdaptor.Unicast(address, msgByte, h.ConsensusID, []byte("consensus"))
}

// func (h *AsynchronousImpl) ProcessProposal(b *pb.Block, p []byte) {
// 	h.Executor.CommitBlock(b, p, h.ConsensusID)
// 	h.MemPool.Remove(types.RawTxArrayFromBytes(b.Txs))
// }

func (a *AsynchronousImpl) Stop() {
	a.Log.Info("stopping consensus")
	if a.ID == 1 {
		a.Metric.PrintMetric()
	}
	a.SafeExit()
	a.cancel()
	// a.Log.Info("stopping consensus end")
}

func (a *AsynchronousImpl) GetWeight(nid int64) float64 {
	if nid < int64(len(a.Config.Nodes)) {
		return 1.0 / float64(len(a.Config.Nodes))
	}
	return 0.0
}

func (a *AsynchronousImpl) GetMaxAdversaryWeight() float64 {
	return 1.0 / 3.0
}

func (a *AsynchronousImpl) VerifyBlock(block []byte, proof []byte) bool {
	return true
}
