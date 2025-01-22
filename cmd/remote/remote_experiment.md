remote experiment

This document details how to run experiments on multiple remote servers.

First, you should have several servers with the same system and a host to start the experiment. Servers are used to run replicas and clients. Note that the host can also act as a server. It is recommended to use Linux systems for both servers and host, because there may be some subtle problems when running WAN experiments under Windows systems.

## Servers

### Install golang

Check out the official Go version list in https://golang.org/dl/.

Download the Go installation package using wget and extract it to the /usr/local directory.

Pay attention to downloading the installation package corresponding to the server architecture.

```
sudo wget -c https://dl.google.com/go/go1.20.3.linux-amd64.tar.gz -O - | sudo tar -xz -C /usr/local
sudo wget -c https://dl.google.com/go/go1.20.3.linux-arm64.tar.gz -O - | sudo tar -xz -C /usr/local
```

Make sure all servers have the same version of golang installed.

Edit bashrc file.

```
vim ~/.bashrc
```

Add the following content at the end of the file:

```
export PATH=$PATH:/usr/local/go/bin
```

Execute the following command to make the environment variables take effect.

```
source ~/.bashrc
go version
```

### Enable SFTP

Make sure all servers enable sftp. Edit the sshd_config file.

```
sudo vi /etc/ssh/sshd_config
```

Comment out the original line and add a new one.

```
# Subsystem sftp /usr/libexec/openssh/sftp-server
Subsystem sftp internal-sftp
```

Then restart the sshd service.

```
sudo systemctl restart sshd.service
```

### Install bn

```
sudo apt update

sudo apt install llvm g++ libgmp-dev libssl-dev

git clone https://github.com/dfinity/bn
cd bn
sudo apt install make
make
sudo make install

export LD_LIBRARY_PATH=/lib:/usr/lib:/usr/local/lib
```

Reference https://github.com/dfinity-side-projects/bn

### Download code

Clone code from github.

```
git clone https://github.com/yrdsm666/tockowl.git
```

Enter the code directory and compile the code.

```
cd upgradeable-consensus
go mod tidy
make
```

## Hosts

Enter the cmd/remote/remoteconfig directory and create the remote_config.yaml file

Add server information to the configuration file.

```
server_list:
  - host: 192.168.198.138
    port: 22
    username: "lmh"
    password: "123456"
  - host: 192.168.198.139
    port: 22
    username: "lmh2"
    password: "123456"
```

Enter the code directory and compile the code.

```
make
```

Run remote experiments.

```
./remote
```

