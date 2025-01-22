package consensus

import (
	"github.com/sirupsen/logrus"
	"github.com/yrdsm666/tockowl/config"
	"github.com/yrdsm666/tockowl/consensus/bkr"
	"github.com/yrdsm666/tockowl/consensus/dumbo_ng"
	"github.com/yrdsm666/tockowl/consensus/fin"
	"github.com/yrdsm666/tockowl/consensus/model"
	"github.com/yrdsm666/tockowl/consensus/parbft"
	"github.com/yrdsm666/tockowl/consensus/sdumbo"
	"github.com/yrdsm666/tockowl/consensus/tockcat"
	"github.com/yrdsm666/tockowl/consensus/tockowl"
	"github.com/yrdsm666/tockowl/executor"
	"github.com/yrdsm666/tockowl/p2p"
)

// a mux method to start consensus
func BuildConsensus(
	nid int64,
	cid int64,
	cfg *config.ConsensusConfig,
	exec executor.Executor,
	p2pAdaptor p2p.P2PAdaptor,
	log *logrus.Entry,
	partition string,
	upd func(int64),
) model.Consensus {
	var c model.Consensus = nil
	switch cfg.Type {
	case "sdumbo":
		switch cfg.Sdumbo.Type {
		case "acs":
			c = sdumbo.NewSdumboAab(nid, cid, cfg, exec, p2pAdaptor, log, partition)
		case "mvba":
			c = sdumbo.NewSdumboMvba(nid, cid, cfg, exec, p2pAdaptor, log, partition)
		default:
			log.Warnf("Sdumbo type not supported: %s", cfg.Sdumbo.Type)
		}
	case "tockowl":
		switch cfg.Tockowl.Type {
		case "acs":
			c = tockowl.NewTockowlAab(nid, cid, cfg, exec, p2pAdaptor, log, partition)
		case "mvba":
			c = tockowl.NewTockowlMvba(nid, cid, cfg, exec, p2pAdaptor, log, partition)
		default:
			log.Warnf("Tockowl type not supported: %s", cfg.Tockowl.Type)
		}
	case "tockowl+":
		c = tockowl.NewTockowlMvba(nid, cid, cfg, exec, p2pAdaptor, log, partition)
	case "bkr":
		switch cfg.Bkr.Type {
		case "ba":
			c = bkr.NewBAConsensus(nid, cid, cfg, exec, p2pAdaptor, log, partition)
		case "bkr":
			c = bkr.NewBKRConsensus(nid, cid, cfg, exec, p2pAdaptor, log, partition)
		default:
			log.Warnf("Bkr type not supported: %s", cfg.Tockowl.Type)
		}
	case "fin":
		c = fin.NewFinConsensus(nid, cid, cfg, exec, p2pAdaptor, log, partition)
	case "tockcat":
		c = tockcat.NewTockCatConsensus(nid, cid, cfg, exec, p2pAdaptor, log, partition)
	case "parbft":
		c = parbft.NewParbftConsensus(nid, cid, cfg, exec, p2pAdaptor, log, partition)
	case "dumbo_ng":
		c = dumbo_ng.NewDumboNgConsensus(nid, cid, cfg, exec, p2pAdaptor, log, partition)
	default:
		log.Warnf("init consensus type not supported: %s", cfg.Type)
	}
	return c
}
