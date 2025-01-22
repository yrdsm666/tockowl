package main

import (
	"io/ioutil"
	"os"
	"path"
	"sync"

	"github.com/pkg/sftp"
	log "github.com/sirupsen/logrus"
	"github.com/yrdsm666/tockowl/cmd/remote/remoteconfig"
)

func WriteSingleFileToServers(serverList []remoteconfig.Server, localFile string, dstDirPath string, newFileName string) {
	serverNum := len(serverList)
	var wg sync.WaitGroup
	wg.Add(serverNum)
	for i := 0; i < serverNum; i++ {
		go func(server remoteconfig.Server) {
			defer wg.Done()
			WriteSingleFileToServer(server, localFile, dstDirPath, newFileName)
		}(serverList[i])
	}
	wg.Wait()
}

func WriteSingleFileToServer(server remoteconfig.Server, localFilePath string, remotePath string, newFileName string) {
	var client *sftp.Client
	if server.SSHClient != nil {
		client = CreateSftpClientBySSHClient(server.SSHClient)
	} else {
		client = CreateSftpClient(server.Host, server.Port, server.Username, server.Password)
	}

	if client == nil {
		return
	}

	defer client.Close()
	// 判断文件夹是否存在
	_, err := client.Stat(remotePath)
	if os.IsNotExist(err) {
		// 文件夹不存在，创建文件夹
		err := client.MkdirAll(remotePath)
		if err != nil {
			log.Fatalf("failed to create dir %s on server(%s): %v\n", remotePath, server.Host, err)
			return
		}
	}
	fileContent, err := os.ReadFile(localFilePath)
	if err != nil {
		log.Printf("%s\n", err)
		return
	}
	srcFile := File{
		Name:    newFileName,
		Content: fileContent,
	}

	// 打开文件进行写入
	filePath := JoinPath(remotePath, srcFile.Name)
	dstFile, err := client.Create(filePath)
	if err != nil {
		log.Fatalf("failed to create file %s on server(%s): %v", filePath, server.Host, err)
		return
	}
	defer dstFile.Close()

	// 写入字节数组到文件
	_, err = dstFile.Write(srcFile.Content)
	if err != nil {
		log.Fatalf("failed to write to %s on server(%s): %v", filePath, server.Host, err)
		return
	}
}

func WriteFilesInSameDirToServers(serverList []remoteconfig.Server, localPath string, dstDirPath string) {
	serverNum := len(serverList)
	var wg sync.WaitGroup
	wg.Add(serverNum)
	for i := 0; i < serverNum; i++ {
		go func(server remoteconfig.Server) {
			defer wg.Done()
			WriteFilesInSameDirToServer(server, localPath, dstDirPath)
		}(serverList[i])
	}
	wg.Wait()
}

func WriteFilesInSameDirToServer(server remoteconfig.Server, localPath string, remotePath string) {
	var client *sftp.Client
	if server.SSHClient != nil {
		client = CreateSftpClientBySSHClient(server.SSHClient)
	} else {
		client = CreateSftpClient(server.Host, server.Port, server.Username, server.Password)
	}

	if client == nil {
		return
	}

	defer client.Close()

	localFiles, err := ioutil.ReadDir(localPath)
	if err != nil {
		log.Fatal("read dir list fail ", err)
	}

	// client.Mkdir(remotePath)
	for _, backupDir := range localFiles {
		localFilePath := path.Join(localPath, backupDir.Name())
		remoteFilePath := path.Join(remotePath, backupDir.Name())
		if backupDir.IsDir() {
			client.Mkdir(remoteFilePath)
			WriteFilesInSameDirToServer(server, localFilePath, remoteFilePath)
		} else {
			WriteSingleFileToServer(server, path.Join(localPath, backupDir.Name()), remotePath, "keySet")
		}
	}
}
