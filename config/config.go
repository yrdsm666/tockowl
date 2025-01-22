package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/yrdsm666/tockowl/crypto"
	"gopkg.in/yaml.v3"
)

type NetworkType = int64

const (
	NetworkSync  NetworkType = iota
	NetworkAsync NetworkType = iota
)

type BkrConfig struct {
	Type string `yaml:"type"`
}

type FinConfig struct {
}

type TockCatConfig struct {
	Shortcut bool `yaml:"shortcut"`
}

type ParbftConfig struct {
	Type           string `yaml:"type"`
	TimeOut        int    `yaml:"timeout"`
	Shortcut       bool   `yaml:"shortcut"`
	LeaderRotation string `yaml:"rotation"`
}

type SdumboConfig struct {
	Type string `yaml:"type"`
	// LightLoad      bool   `yaml:"light_load"`
	// LightLoadNodes []int  `yaml:"light_load_nodes"`
	Shortcut bool `yaml:"shortcut"`
}

type TockowlConfig struct {
	Type string `yaml:"type"`
	// LightLoad      bool   `yaml:"light_load"`
	// LightLoadNodes []int  `yaml:"light_load_nodes"`
}

type Tockowl2Config struct {
	TimeOut        int    `yaml:"timeout"`
	LeaderRotation string `yaml:"rotation"`
}

type DumboNgConfig struct {
	ConsensusType  string `yaml:"consensus_type"`
	TimeOut        int    `yaml:"timeout"`
	Shortcut       bool   `yaml:"shortcut"`
	LeaderRotation string `yaml:"rotation"`
}

// type TxpoolConfig struct {
// 	Type  string `yaml:"type"`
// 	Epoch int    `yaml:"epoch"`
// }

type FaultConfig struct {
	Type       string `yaml:"type"`
	Open       bool   `yaml:"open"`
	FaultNodes []int  `yaml:"fault_nodes"`
}

type ConsensusConfig struct {
	Type    string `yaml:"type"`
	Testing bool   `yaml:"testing,omitempty"`
	// ConsensusID int64           `yaml:"consensus_id"`
	Bkr      *BkrConfig      `yaml:"bkr,omitempty"`
	Fin      *FinConfig      `yaml:"fin,omitempty"`
	TockCat  *TockCatConfig  `yaml:"tockcat,omitempty"`
	Parbft   *ParbftConfig   `yaml:"parbft,omitempty"`
	Sdumbo   *SdumboConfig   `yaml:"sdumbo,omitempty"`
	Tockowl  *TockowlConfig  `yaml:"tockowl,omitempty"`
	Tockowl2 *Tockowl2Config `yaml:"tockowl+,omitempty"`
	DumboNg  *DumboNgConfig  `yaml:"dumbo_ng,omitempty"`
	// Txpool      *TxpoolConfig      `yaml:"txpool,omitempty"`
	Byzantine   *FaultConfig
	Nodes       []*ReplicaInfo
	Keys        *crypto.KeySet
	F           int // Initialize in NewConsensus if neeeded
	BatchSize   int `yaml:"batch_size"`
	PayloadSize int `yaml:"payload_size"`
	Epoch       int `yaml:"epoch"`
	Topic       string
}

type LogConfig struct {
	Level    string `yaml:"level"`
	ToFile   bool   `yaml:"to_file"`
	Filename string `yaml:"filename"`
}

type ExecutorConfig struct {
	Type    string `yaml:"type"`
	Address string `yaml:"address"`
}

type P2PConfig struct {
	Type       string `yaml:"type"`
	ExtraDelay int    `yaml:"extra_delay"`
}

type RemoteExperiment struct {
	Open bool `yaml:"open"`
	Time int  `yaml:"time"`
}

type Config struct {
	Total               int64             `yaml:"total"`
	AddressStartPort    string            `yaml:"address_start_port"`
	RpcAddressStartPort string            `yaml:"rpc_address_start_port"`
	AddressIp           string            `yaml:"address_ip"`
	Log                 *LogConfig        `yaml:"log"`
	Executor            *ExecutorConfig   `yaml:"executor"`
	P2P                 *P2PConfig        `yaml:"p2p"`
	Consensus           *ConsensusConfig  `yaml:"consensus"`
	NetworkType         int               `yaml:"network_type"`
	KeyGen              bool              `yaml:"key_gen"`
	Fault               *FaultConfig      `yaml:"fault"`
	Crypto              string            `yaml:"crypto"`
	RemoteExperiment    *RemoteExperiment `yaml:"remote_experiment"`
	Keys                *crypto.KeySet
	Nodes               []*ReplicaInfo `yaml:"nodes"`
	Topic               string         `yaml:"topic"`
}

type ReplicaInfo struct {
	ID         int64  `yaml:"id"`
	Address    string `yaml:"address"`
	RpcAddress string `yaml:"rpc_address"` // RpcAddress is a external address for client
	IsLocal    bool   `yaml:"is_local"`
}

func NewConfig(path string, id int64, partition string) (*Config, error) {
	cfg := new(Config)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	if !cfg.RemoteExperiment.Open {
		cfg.Nodes = make([]*ReplicaInfo, cfg.Total)
		for i := int64(0); i < cfg.Total; i++ {
			node := &ReplicaInfo{
				ID:         i,
				Address:    cfg.GetAddress(i, cfg.AddressStartPort, cfg.AddressIp),
				RpcAddress: cfg.GetAddress(i, cfg.RpcAddressStartPort, cfg.AddressIp),
			}
			cfg.Nodes[i] = node
		}
	}

	cfg.Consensus.Nodes = cfg.Nodes
	cfg.Consensus.F = (len(cfg.Nodes) - 1) / 3
	cfg.Consensus.Byzantine = cfg.Fault
	cfg.Consensus.Topic = cfg.Topic
	return cfg, nil
}

func (c *ConsensusConfig) GetNodeInfo(id int64) *ReplicaInfo {
	for _, info := range c.Nodes {
		if info.ID == id {
			return info
		}
	}
	panic(fmt.Sprintf("node %d does not exist", id))
}

func (c *Config) GetNodeInfo(id int64) *ReplicaInfo {
	for _, info := range c.Nodes {
		if info.ID == id {
			return info
		}
	}
	panic(fmt.Sprintf("node %d does not exist", id))
}

func (c *Config) GetAddress(id int64, initPortStr string, ip string) string {
	initPort, err := strconv.Atoi(initPortStr)
	if err != nil {
		panic(fmt.Sprintf("GetAddress error: %s", err))
	}
	address := ip + ":" + strconv.Itoa(initPort+int(id))
	return address
}
