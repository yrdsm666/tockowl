package remoteconfig

import (
	"log"
	"os"

	"github.com/yrdsm666/tockowl/config"
	"gopkg.in/yaml.v3"
)

type ConsensusConfig struct {
	Total int64 `yaml:"total"`
	// PublicKeyPath string           `yaml:"public_key_path"`
	Log         *config.LogConfig       `yaml:"log"`
	Executor    *config.ExecutorConfig  `yaml:"executor"`
	P2P         *config.P2PConfig       `yaml:"p2p"`
	Consensus   *config.ConsensusConfig `yaml:"consensus"`
	NetworkType int                     `yaml:"network_type"`
	// Keys          *KeySet
	KeyGen           bool                     `yaml:"key_gen"`
	Byzantine        *config.FaultConfig      `yaml:"fault"`
	Crypto           string                   `yaml:"crypto"`
	Nodes            []*config.ReplicaInfo    `yaml:"nodes"`
	RemoteExperiment *config.RemoteExperiment `yaml:"remote_experiment"`
	Topic            string                   `yaml:"topic"`
}

func InitConsensusConfig(path string) (*ConsensusConfig, error) {
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("can not read config file: %s\n.", path)
		return nil, err
	}
	var config *ConsensusConfig
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("can not unmarshal config file: %s\n.", path)
		return nil, err
	}
	return config, nil
}
