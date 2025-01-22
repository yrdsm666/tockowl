package remoteconfig

import (
	"log"
	"os"

	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v3"
)

type RemoteConfig struct {
	LocalKeyDirPath          string   `yaml:"local_key_dir_path"`
	ServerKeyDirPath         string   `yaml:"server_key_dir_relative_path"`
	LocalConfigPath          string   `yaml:"local_config_path"`
	ServerConfigPath         string   `yaml:"server_config_relative_path"`
	LocalExecutableFilePath  string   `yaml:"local_executable_file_path"`
	ServerExecutableFilePath string   `yaml:"server_executable_file_path"`
	ServerList               []Server `yaml:"server_list"`
}

type Server struct {
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
	SSHClient *ssh.Client
}

func InitRemoteConfig(path string) (*RemoteConfig, error) {
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("can not read config file: %s\n.", path)
		return nil, err
	}
	var config *RemoteConfig
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("can not unmarshal config file: %s\n.", path)
		return nil, err
	}
	return config, nil
}
