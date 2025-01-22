package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/yrdsm666/tockowl/cmd/remote/remoteconfig"
	"github.com/yrdsm666/tockowl/config"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
)

func main() {
	rcfg, err := remoteconfig.InitRemoteConfig("./cmd/remote/remoteconfig/remote_config.yaml")
	if err != nil {
		log.Fatalf("Init remote config failed: %v", err)
		return
	}
	fmt.Println(rcfg.ServerList)

	err = ConnectServers(rcfg)
	if err != nil {
		log.Fatalf("Connect servers failed: %v", err)
		return
	}

	ccfg, err := remoteconfig.InitConsensusConfig("./cmd/remote/remoteconfig/consensus_config.yaml")
	if err != nil {
		log.Fatalf("Init consensus config failed: %v", err)
	}

	fmt.Println("AssignReplicasToServers...")
	nodes := AssignReplicasToServers(int(ccfg.Total), rcfg.ServerList)

	fmt.Println("CreateConfigFileForServers...")
	CreateConfigFileForServers(nodes, rcfg.ServerList, ccfg)

	fmt.Println("UploadConfigFileToServers...")
	UploadConfigFileToServers(rcfg.ServerList, rcfg)

	// Both executable files and key files can be uploaded in advance
	// fmt.Println("UploadExecutableFileToServers...")
	// UploadExecutableFileToServers(rcfg.ServerList, rcfg)

	fmt.Println("UploadKeysToServers...")
	localKeyDir := fmt.Sprintf("%s/keys_%s_%s%d", rcfg.LocalKeyDirPath, ccfg.Crypto, "", len(nodes))
	serverKeyDir := fmt.Sprintf("%s/keys_%s_%s%d", rcfg.ServerKeyDirPath, ccfg.Crypto, "", len(nodes))
	UploadKeysToServers(rcfg.ServerList, localKeyDir, serverKeyDir)

	executableFileName := "server"
	fmt.Println("StartConsensusInServers...")
	cmd := fmt.Sprintf("cd ~/tockowl/; chmod 777 %s; ./%s", executableFileName, executableFileName)
	StartConsensusInServers(rcfg.ServerList, cmd)

	fmt.Println("CloseSSHConnect...")
	err = CloseSSHConnect(rcfg)
	if err != nil {
		log.Fatalf("Close SSH Connect falied: %v", err)
		return
	}
}

func GetAddress(id int, initPortStr string, ip string) string {
	initPort, err := strconv.Atoi(initPortStr)
	if err != nil {
		panic(fmt.Sprintf("GetAddress error: %s", err))
	}
	address := ip + ":" + strconv.Itoa(initPort+id)
	return address
}

func ConnectServers(rcfg *remoteconfig.RemoteConfig) error {
	connServers := make([]remoteconfig.Server, len(rcfg.ServerList))
	for i := 0; i < len(rcfg.ServerList); i++ {
		connServers[i] = rcfg.ServerList[i]

		fmt.Printf("Connect Server ip: %s:%d\n", rcfg.ServerList[i].Host, rcfg.ServerList[i].Port)
		conn, err := ConnectByPassword(rcfg.ServerList[i].Host, rcfg.ServerList[i].Port, rcfg.ServerList[i].Username, rcfg.ServerList[i].Password)
		if err != nil {
			log.Fatalf("Connect Server ip: %s:%d, error: %s\n", rcfg.ServerList[i].Host, rcfg.ServerList[i].Port, err)
			return nil
		}
		connServers[i].SSHClient = conn
	}
	rcfg.ServerList = connServers
	return nil
}

func CloseSSHConnect(rcfg *remoteconfig.RemoteConfig) error {
	for _, s := range rcfg.ServerList {
		if s.SSHClient != nil {
			s.SSHClient.Close()
		}
	}
	return nil
}

func AssignReplicasToServers(totalNode int, serverList []remoteconfig.Server) []*config.ReplicaInfo {
	totalServer := len(serverList)
	avgNode := float64(totalNode) / float64(totalServer)

	nodes := make([]*config.ReplicaInfo, 0)

	serverId := 0
	numberInServer := 0

	// Determine server based on replica id
	for id := 0; id < totalNode; id++ {
		if float64(id) >= float64(serverId+1)*float64(avgNode) && (serverId+1) < totalServer {
			serverId += 1
			numberInServer = 0
		}
		newNodeInfo := &config.ReplicaInfo{
			ID:      int64(id),
			Address: GetAddress(numberInServer, "8080", serverList[serverId].Host),
			IsLocal: false,
		}
		numberInServer += 1

		nodes = append(nodes, newNodeInfo)
	}

	return nodes
}

func CreateConfigFileForServers(nodes []*config.ReplicaInfo, serverList []remoteconfig.Server, ccfg *remoteconfig.ConsensusConfig) {
	totalNode := len(nodes)
	totalServer := len(serverList)
	avgNode := float64(totalNode) / float64(totalServer)
	fmt.Println("avg: ", avgNode)

	// Determine replica based on serverId
	for id := 0; id < totalServer; id++ {

		serverNodes := make([]*config.ReplicaInfo, len(nodes))
		// DeepCopy
		for j := 0; j < len(nodes); j++ {
			serverNodes[j] = &config.ReplicaInfo{
				ID:      nodes[j].ID,
				Address: nodes[j].Address,
				IsLocal: nodes[j].IsLocal,
			}
		}

		// Replicas numbering starts from 0
		for j := 0; j < totalNode; j++ {
			if (float64(j) >= float64(id)*avgNode && float64(j) < float64(id+1)*avgNode) || (float64(j) >= float64(id)*avgNode && id == totalServer-1) {
				fmt.Printf("replica %d address is %s\n", j, serverNodes[j].Address)
				serverNodes[j].IsLocal = true
			}
			// fmt.Println("serverId: ", id, "  nodeId: ", j, "  islocal: ", serverNodes[j].IsLocal)
		}
		scfg := ccfg
		scfg.Nodes = serverNodes

		yamlBytes, err := yaml.Marshal(scfg)
		if err != nil {
			log.Fatalf("Failed to encode Yaml: %v", err)
		}
		fileName := fmt.Sprintf("./cmd/remote/test_%s.yaml", serverList[id].Host)

		err = os.WriteFile(fileName, yamlBytes, 0777)
		if err != nil {
			log.Fatalf("Failed to encode write file: %v", err)
		}

	}
}

func UploadConfigFileToServers(serverList []remoteconfig.Server, rcfg *remoteconfig.RemoteConfig) {
	totalServer := len(serverList)

	for id := 0; id < totalServer; id++ {
		fileName := fmt.Sprintf("./cmd/remote/test_%s.yaml", serverList[id].Host)
		WriteSingleFileToServer(rcfg.ServerList[id], fileName, GetDir(rcfg.ServerConfigPath), "config.yaml")
	}
}

func UploadExecutableFileToServers(serverList []remoteconfig.Server, rcfg *remoteconfig.RemoteConfig) {
	totalServer := len(serverList)

	for id := 0; id < totalServer; id++ {
		WriteSingleFileToServer(rcfg.ServerList[id], rcfg.LocalExecutableFilePath, rcfg.ServerExecutableFilePath, "server")
	}
}

func UploadKeysToServers(serverList []remoteconfig.Server, localKeyDir string, serverKeyDir string) {
	fmt.Println("localKeyDir: ", localKeyDir)
	fmt.Println("ServerKeyDir: ", serverKeyDir)

	WriteFilesInSameDirToServers(serverList, localKeyDir, serverKeyDir)

}

func StartConsensusInServers(serverList []remoteconfig.Server, cmd string) error {
	serverNum := len(serverList)
	var wg sync.WaitGroup
	wg.Add(serverNum)
	for i := 0; i < len(serverList); i++ {
		go func(index int) {
			defer wg.Done()
			if serverList[index].SSHClient != nil {
				ExecuteCommandInServer(serverList[index].SSHClient, cmd, index)
			} else {
				log.Fatal("error: SSH client is nil")
			}
		}(i)
	}

	// for _, s := range serverList {
	// 	go func(server remoteconfig.Server) {
	// 		defer wg.Done()
	// 		if server.SSHClient != nil {
	// 			ExecuteCommandInServer(server.SSHClient, cmd, server.Host)
	// 		} else {
	// 			log.Fatal("error: SSH client is nil")
	// 		}
	// 	}(s)
	// }
	wg.Wait()
	return nil
}

func ExecuteCommandInServer(sshClient *ssh.Client, cmd string, index int) {
	// 创建 ssh session 会话
	session, err := sshClient.NewSession()
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// 执行远程命令
	cmdInfo, err := session.CombinedOutput(cmd)
	if err != nil {
		fmt.Println(string(cmdInfo))
		panic(err)
	}

	if index == 0 {
		fmt.Println(string(cmdInfo))
	}
}
