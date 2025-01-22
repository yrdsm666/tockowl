package main

import (
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

func ConnectByPassword(host string, port int, account string, password string) (*ssh.Client, error) {
	// 创建ssh登录配置
	config := &ssh.ClientConfig{
		Timeout:         5 * time.Second,                          // 超时时间
		User:            account,                                  // 登录账号
		Auth:            []ssh.AuthMethod{ssh.Password(password)}, // 密码
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),              // 这个不够安全，生产环境不建议使用
		// HostKeyCallback: ssh.FixedHostKey(), // 建议使用这种，目前还没研究出怎么使用[todo]
	}

	// dial连接服务器
	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		log.Fatal("连接服务器失败, ip:", host, " | ", err)
		return nil, err
	}
	return client, nil
}

// ConnectByKey 使用秘钥连接Linux服务器
func ConnectByKey(host string, sshPort int, account string, keyFile string) (*ssh.Client, error) {
	// 读取秘钥
	key, err := os.ReadFile(keyFile)
	if err != nil {
		log.Fatal("秘钥读取失败", " | ", err)
		return nil, err
	}

	// 创建秘钥签名
	// 会拿着秘钥去登录服务器
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatal("秘钥签名失败", " | ", err)
		return nil, err
	}

	// 创建ssh登录配置
	config := &ssh.ClientConfig{
		Timeout: 5 * time.Second, // 超时时间
		User:    account,
		Auth: []ssh.AuthMethod{
			// 使用秘钥登录远程服务器
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 这个不够安全，生产环境不建议使用
		// var hostKey ssh.PublicKey
		// HostKeyCallback: ssh.FixedHostKey(hostKey),
	}

	// 连接远程服务器
	addr := fmt.Sprintf("%s:%d", host, sshPort)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		log.Fatal("连接服务器失败, ip:", host, " | ", err)
	}

	return client, nil
}
