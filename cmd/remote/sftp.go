package main

import (
	"github.com/pkg/sftp"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

// CreateSftpClient 创建sftp会话
func CreateSftpClient(host string, port int, account, password string) *sftp.Client {
	// 连接Linux服务器
	conn, err := ConnectByPassword(host, port, account, password)
	if err != nil {
		log.Fatalf("\"连接服务器失败 ip: %s:%d, error: %s\\n\"", host, port, err)
		return nil
	}

	// 创建sftp会话
	client, err := sftp.NewClient(conn)
	if err != nil {
		conn.Close()
		log.Fatalf("\"创建sftp会话失败 ip: %s:%d, error: %s\\n\"", host, port, err)
		return nil
	}
	return client
}

// CreateSftpClientBySSHClient 在已建立ssh会话的基础上创建sftp会话
func CreateSftpClientBySSHClient(sshClient *ssh.Client) *sftp.Client {
	// 创建sftp会话
	client, err := sftp.NewClient(sshClient)
	if err != nil {
		return nil
	}
	return client
}
