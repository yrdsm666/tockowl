package party

import (
	"context"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/yrdsm666/tockowl/config"
	"github.com/yrdsm666/tockowl/consensus/pkg/protobuf"
	"github.com/yrdsm666/tockowl/executor"
	"github.com/yrdsm666/tockowl/p2p"
	"google.golang.org/protobuf/proto"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/share"
)

// Party is a interface of consensus parties
type Party interface {
	send(m *protobuf.Message, des uint32) error
	broadcast(m *protobuf.Message) error
	getMessageWithType(messageType string) (*protobuf.Message, error)
}

// HonestParty is a struct of honest consensus parties
type HonestParty struct {
	AsynchronousImpl
	N   uint32
	F   uint32
	PID uint32

	FastTrack      bool
	Shortcut       bool
	TimeOut        int
	LeaderRotation string

	LightLoad      bool
	LightLoadNodes []int

	dispatcheChannels *sync.Map
}

// NewHonestParty return a new honest party object
func NewHonestParty(N uint32, F uint32, pid uint32, ipList []string, portList []string, sigPK *share.PubPoly, sigSK *share.PriShare, encPK kyber.Point, encVK []*share.PubShare, encSK *share.PriShare) *HonestParty {
	p := HonestParty{
		N:              N,
		F:              F,
		PID:            pid,
		FastTrack:      false,
		Shortcut:       false,
		TimeOut:        0,
		LeaderRotation: "fixed",
		LightLoad:      false,
	}

	return &p
}

// InitHonestParty return a new honest party object
func InitHonestParty(
	id int64,
	cid int64,
	cfg *config.ConsensusConfig,
	exec executor.Executor,
	p2pAdaptor p2p.P2PAdaptor,
	log *logrus.Entry,
	partition string,
) *HonestParty {
	log.WithField("consensus id", cid).Info("[TOCKOWL] starting")
	var p *HonestParty

	switch cfg.Type {
	case "tockowl":
		p = &HonestParty{
			N:   uint32(len(cfg.Nodes)),
			F:   uint32(cfg.F),
			PID: uint32(id),
			// LightLoad: cfg.Tockowl.LightLoad,
		}
	case "tockowl+":
		p = &HonestParty{
			N:              uint32(len(cfg.Nodes)),
			F:              uint32(cfg.F),
			PID:            uint32(id),
			TimeOut:        cfg.Tockowl2.TimeOut,
			LeaderRotation: cfg.Tockowl2.LeaderRotation,
			FastTrack:      true,
		}
	case "tockcat":
		p = &HonestParty{
			N:        uint32(len(cfg.Nodes)),
			F:        uint32(cfg.F),
			PID:      uint32(id),
			Shortcut: cfg.TockCat.Shortcut,
		}
	case "sdumbo":
		p = &HonestParty{
			N:        uint32(len(cfg.Nodes)),
			F:        uint32(cfg.F),
			PID:      uint32(id),
			Shortcut: cfg.Sdumbo.Shortcut,
			// LightLoad: cfg.Sdumbo.LightLoad,
		}
	case "parbft":
		p = &HonestParty{
			N:              uint32(len(cfg.Nodes)),
			F:              uint32(cfg.F),
			PID:            uint32(id),
			Shortcut:       cfg.Parbft.Shortcut,
			TimeOut:        cfg.Parbft.TimeOut,
			LeaderRotation: cfg.Parbft.LeaderRotation,
		}
	case "dumbo_ng":
		p = &HonestParty{
			N:   uint32(len(cfg.Nodes)),
			F:   uint32(cfg.F),
			PID: uint32(id),
		}
		if cfg.DumboNg.ConsensusType == "tockowl+" {
			p.FastTrack = true
		}
		if cfg.DumboNg.ConsensusType == "tockcat" || cfg.DumboNg.ConsensusType == "sdumbo" || cfg.DumboNg.ConsensusType == "parbft1" || cfg.DumboNg.ConsensusType == "parbft2" {
			p.Shortcut = cfg.DumboNg.Shortcut
		}
		if cfg.DumboNg.ConsensusType == "tockowl+" || cfg.DumboNg.ConsensusType == "parbft2" {
			p.TimeOut = cfg.DumboNg.TimeOut
			p.LeaderRotation = cfg.DumboNg.LeaderRotation
		}
		if cfg.DumboNg.ConsensusType == "parbft1" {
			p.LeaderRotation = cfg.DumboNg.LeaderRotation
		}

	default:
		p = &HonestParty{
			N:   uint32(len(cfg.Nodes)),
			F:   uint32(cfg.F),
			PID: uint32(id),
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	p.Init(id, cid, cfg, exec, p2pAdaptor, log, cancel, partition)

	p.dispatcheChannels = new(sync.Map)

	go p.receiveMsg(ctx)

	return p
}

func (p *HonestParty) receiveMsg(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msgByte := <-p.MsgByteEntrance:
			msg, err := p.DecodeMsgByte(msgByte)
			if err != nil {
				p.Log.WithError(err).Warn("decode message failed")
				continue
			}
			go p.handleMsg(msg)
		}
	}
}

var Mu = new(sync.Mutex)
var Traffic = 0

func (p *HonestParty) handleMsg(m *protobuf.Message) {
	value1, _ := p.dispatcheChannels.LoadOrStore(m.Type, new(sync.Map))

	var value2 any
	N := p.N
	if m.Type == "Dec" {
		value2, _ = value1.(*sync.Map).LoadOrStore(string(m.Id), make(chan *protobuf.Message, N*N))
	} else {
		value2, _ = value1.(*sync.Map).LoadOrStore(string(m.Id), make(chan *protobuf.Message, N))
	}

	value2.(chan *protobuf.Message) <- m

	Mu.Lock()
	Traffic += proto.Size(m)
	Mu.Unlock()
}

// GetMessage Try to get a message according to messageType, ID
func (p *HonestParty) GetMessage(messageType string, ID []byte) chan *protobuf.Message {
	value1, _ := p.dispatcheChannels.LoadOrStore(messageType, new(sync.Map))

	var value2 any
	if messageType == "Dec" {
		value2, _ = value1.(*sync.Map).LoadOrStore(string(ID), make(chan *protobuf.Message, p.N*p.N))
	} else {
		value2, _ = value1.(*sync.Map).LoadOrStore(string(ID), make(chan *protobuf.Message, p.N))
	}

	return value2.(chan *protobuf.Message)
}

// GetMessage Try to get a message according to messageType, ID
func (p *HonestParty) GetLeaderByEpoch(epoch uint32) uint32 {
	if p.LeaderRotation == "fixed" {
		return 0
	} else if p.LeaderRotation == "roundrobin" {
		return epoch % p.N
	}
	p.Log.Error("unsupport type of the leader rotation: ", p.LeaderRotation)
	return 0
}

func (p *HonestParty) IsCrashNode(id int64) bool {
	if p.Config.Byzantine.Open && p.Config.Byzantine.Type == "crash" {
		for _, c := range p.Config.Byzantine.FaultNodes {
			if id == int64(c) {
				return true
			}
		}
	}
	return false
}

func (p *HonestParty) IsByzantineNode(id int64) bool {
	if p.Config.Byzantine.Open && p.Config.Byzantine.Type == "byz" {
		for _, c := range p.Config.Byzantine.FaultNodes {
			if id == int64(c) {
				return true
			}
		}
	}
	return false
}

func (p *HonestParty) IsLightLoad(id int64) bool {
	if p.Config.Tockowl == nil {
		return false
	}

	if p.LightLoad {
		for _, c := range p.LightLoadNodes {
			if id == int64(c) {
				return true
			}
		}
	}
	return false
}
