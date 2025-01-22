package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/yrdsm666/tockowl"
	"github.com/yrdsm666/tockowl/config"
	"github.com/yrdsm666/tockowl/crypto"
	"github.com/yrdsm666/tockowl/logging"
)

var (
	logger  = logging.GetLogger()
	sigChan = make(chan os.Signal)
)

func main() {
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM,
		syscall.SIGQUIT)

	cfg, err := config.NewConfig("config/config.yaml", 0, "")
	if err != nil {
		logger.Error("read config.yaml failed: ", err)
		return
	}

	total := len(cfg.Nodes)
	fault := (total - 1) / 3
	partitionName := ""

	crypto.CryptoInit(cfg.Crypto, total, total-fault)

	if cfg.KeyGen {
		logger.Infof("[Consensus] KenGen for %d nodes with %d fault...", total, fault)
		keys := crypto.KeyGen(total, total-fault)
		logger.Info("[Consensus] Store all keys to files...")
		crypto.StoreAllKeys(keys, partitionName)
		return
	}

	if cfg.RemoteExperiment.Open {
		logger.Infof("[Consensus] Remote Experiment")
		satrtRemoteExperiment(cfg)
		return
	}

	// total := 4
	nodes := make([]*tockowl.Node, total)

	active := total
	var wg sync.WaitGroup
	wg.Add(active)
	for i := int64(0); i < int64(active); i++ {
		go func(index int64) {
			defer wg.Done()
			keySet := crypto.LoadKeyFromFiles(index, int64(total), partitionName)
			nodes[index] = tockowl.NewNode(index, keySet, partitionName, func(int64) {})
		}(i)
	}
	wg.Wait()

	<-sigChan
	logger.Info("[Consensus] Exit...")
	for i := 0; i < total; i++ {
		nodes[i].Stop()
	}
}

func satrtRemoteExperiment(cfg *config.Config) {
	localNodesNum := 0
	for _, node := range cfg.Nodes {
		if node.IsLocal {
			localNodesNum++
		}
	}

	nodes := make([]*tockowl.Node, len(cfg.Nodes))
	partitionName := ""

	active := localNodesNum
	var wg1 sync.WaitGroup
	wg1.Add(active)
	for i := int64(0); i < int64(len(cfg.Nodes)); i++ {
		if cfg.Nodes[i].IsLocal {
			go func(index int64) {
				defer wg1.Done()
				keySet := crypto.LoadKeyFromFiles(cfg.Nodes[index].ID, int64(len(cfg.Nodes)), partitionName)
				nodes[index] = tockowl.NewNodeForRemoteExperiment(cfg.Nodes[index].ID, keySet, partitionName, func(int64) {})
			}(i)
		}
	}
	wg1.Wait()

	time.Sleep(time.Duration(cfg.RemoteExperiment.Time) * time.Second)

	logger.Info("[Consensus] Exit...")

	var wg2 sync.WaitGroup
	wg2.Add(active)
	for i := 0; i < len(nodes); i++ {
		if cfg.Nodes[i].IsLocal {
			go func(index int) {
				defer wg2.Done()
				nodes[index].Stop()
			}(i)
		}
	}
	wg2.Wait()
}
