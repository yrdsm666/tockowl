package p2p

import (
	"github.com/sirupsen/logrus"
	"github.com/yrdsm666/tockowl/config"
)

type P2PAdaptor interface {
	Broadcast(msgByte []byte, consensusID int64, topic []byte) error
	Unicast(address string, msgByte []byte, consensusID int64, topic []byte) error
	SetReceiver(receiver MsgReceiver)
	Subscribe(topic []byte) error
	UnSubscribe(topic []byte) error
	GetPeerID() string
	GetP2PType() string
	Stop()
}

type MsgReceiver interface {
	GetMsgByteEntrance() chan<- []byte
}

func BuildP2P(
	cfg *config.P2PConfig,
	log *logrus.Entry,
	id int64,
	networkType int64,
	partition string) (P2PAdaptor, string, error) {
	if cfg.Type == "grpc" {
		return NewBaseNetwork(log, id, networkType, partition)
	}
	log.WithField("type", cfg.Type).Warn("network type error")
	return nil, "", nil
}
