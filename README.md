# TockOwl

This repository contains the artifacts of the paper _TockOwl: Asynchronous Consensus with Fault and Network Adaptability_. It is based on a fork of the open-source implementation of the [Dory](https://github.com/xygdys/Dory-BFT-Consensus) protocol.

This repository contains implementations of seven consensus protocols: TockOwl, TockOwl+, TockCat, [BKR](https://eprint.iacr.org/2016/199.pdf), [Speeding-Dumbo](https://eprint.iacr.org/2022/027.pdf), [ParBFT](https://eprint.iacr.org/2023/679.pdf), and [Dumbo-NG](https://arxiv.org/pdf/2209.00750).


## Structure

The core modules of this repository are as follows:

- `cmd`: The entry point of the program.  
- `config`: Configuration files.  
- `consensus`: The core logic of consensus.  
  - `TockOwl`: Section 4 of the paper.  
  - `TockOwl+`: Section 5 of the paper.  
  - `TockCat`: Appendix B of the paper.  
  - `BKR`: The core of the Honeybadger consensus [1, Figure 2].  
  - `Speeding-Dumbo`: A low-latency MVBA protocol [2].  
  - `ParBFT`: An efficient dual-track protocol based on hedging delays [3].  
  - `Dumbo-NG`: A high-throughput asynchronous BFT consensus [4].  
- `crypto`: An implementation of threshold signatures. Boldyreva’s pairing-based threshold scheme on the BN256 curve, implemented in Kyber, is used for threshold signatures and coin tossing.  
- `executor`: The module for executing transactions, which currently includes two methods: local commitment and remote commitment.  
- `p2p`: The network communication module, which is implemented by using gRPC.


[1] Miller A, Xia Y, Croman K, et al. The honey badger of BFT protocols. CCS 2016.

[2] Guo B, Lu Y, Lu Z, et al. Speeding dumbo: Pushing asynchronous bft closer to practice. NDSS, 2022.

[3] Dai X, Zhang B, Jin H, et al. Parbft: Faster asynchronous bft consensus with a parallel optimistic path. CCS 2023.

[4] Gao Y, Lu Y, Lu Z, et al. Dumbo-ng: Fast asynchronous bft consensus with throughput-oblivious latency. CCS 2022.

## Setup

The operating system and software versions of the machines used in the experiments are as follows. These versions are used to produce the results presented in the paper and are provided for reference purposes only.

- Ubuntu 22.04.5 LTS
- make, version 4.3
- Golang, version 1.23.4


## Running locally

This section explains how to build and execute the program on a local machine.

1. Download the code:

```shell
git clone https://github.com/yrdsm666/tockowl.git
```

2. Compile the program:

```shell
cd tockowl && make
```

3. Create the config file by running `cp config/config.yaml.example config/config.yaml`, and update `config/config.yaml` as needed to suit your requirements (see details below).

### Key Generation

Before running the consensus program, we need to generate keys for the consensus replicas.

1. Update the `key_gen` field (in `config/config.yaml`) to `true`, and specify the total number of consensus replicas. For example:

```yaml
key_gen: true
total: 4
```

2. Run the consensus program:

```shell
./server
```

The program will then generate keys for 10 replicas and store them in the `crypto/keys/keys_tbls_4` directory.

Keys only need to be generated once. However, if the total number of replicas changes—for example, from 4 to 10—the keys must be **regenerated** to accommodate the updated number of replicas.

### Basic configuration

In the program, each replica occupies a port for communication. The `address_start_port` field  (in `config/config.yaml`) specifies the initial port, and the replicas will occupy different ports in sequence from the initial port. For example, in the following configuration, the four replicas will occupy ports 8080-8083. Therefore, make sure these ports are not occupied before running the program.

```yaml
address_start_port: 8080

total: 4
```



### Consensus configuration

The `consensus` field  (in `config/config.yaml`) defines the parameters of the consensus protocol used in the experiment. The relevant fields are described below:

- `batch_size`: The number of transactions included in a block.
- `payload_size`: The size of each transaction in bytes.
- `epoch`: The number of epochs during which the protocol runs; each epoch produces one output.
- `type`: The consensus protocol being executed. Currently, this field supports the following seven protocols: `tockowl`, `tockowl+`, `tockcat`, `bkr`, `sdumbo`, `parbft`, and `dumbo_ng`. The sub-parameters for each protocol are detailed below.

Example configuration:

```yaml
consensus:
  batch_size: 50
  payload_size: 250
  epoch: 20
  type: tockowl
```

#### TockOwl Protocol

The parameters for the TockOwl protocol are defined in the `consensus.tockowl` domain and include the following field:

- `type`: Currently supports two types: `acs` and `mvba`, where `acs` outputs a set of values and `mvba` outputs a single value.

Example configuration:

```yaml
consensus:
  type: tockowl
  tockowl:
    type: mvba
```

#### TockOwl+ Protocol

The parameters for the TockOwl+ protocol are defined in the `consensus.tockowl+` domain and include the following fields:

- `timeout`: The delay time (in milliseconds) before the slow track starts in TockOwl+.

- `rotation`: The method for selecting the leader of the fast track. Currently, `fixed` and `roundrobin` are supported, representing fixed leader (replica 0) and Round-Robin, respectively.

Example configuration:

```yaml
consensus:
  type: tockowl+
  tockowl+:
    timeout: 1000
    rotation: roundrobin
```

#### TockCat Protocol

The parameters for the TockCat protocol are defined in the `consensus.tockcat` domain and include the following field:

- `shortcut`: Specifies whether to enable the output shortcut in the protocol. When enabled, the protocol may produce an output immediately after the common coin step.

Example configuration:

```yaml
consensus:
  type: tockcat
  tockcat:
    shortcut: false
```

#### BKR Protocol

The parameters for the BKR protocol are defined in the `consensus.bkr` domain and include the following:

- `type`: Specifies the protocol type. Currently, two types are supported:
  - `bkr`: Executes the full BKR protocol.
  - `ba`: Executes only the binary consensus (BA) protocol.


Example configuration:

```yaml
consensus:
  type: bkr
  bkr:
    type: bkr
```

#### Speeding-Dumbo Protocol

The parameters for the Speeding-Dumbo protocol are defined in the `consensus.sdumbo` domain and include the following:

- `type`: Specifies the protocol type. Currently, two types are supported:
  - `acs`: Outputs a set of values, corresponding to Algorithm 1 (ACS) of [Speeding-Dumbo](https://eprint.iacr.org/2022/027.pdf).
  - `mvba`: Outputs a single value, corresponding to Algorithm 3 (sMVBA) of [Speeding-Dumbo](https://eprint.iacr.org/2022/027.pdf).

- `shortcut`: Specifies whether to enable the output shortcut. When enabled, the protocol may produce an output immediately after the common coin step.

Example configuration:

```yaml
consensus:
  type: sdumbo
  sdumbo:
    type: mvba
    shortcut: true
```

#### ParBFT protocol

The parameters for the ParBFT protocol are defined in the `consensus.parbft` domain and include the following:

- `type`: Specifies the protocol type. Currently, two types are supported:
  - `parbft1`: Corresponds to the ParBFT1 protocol described in [ParBFT](https://eprint.iacr.org/2023/679.pdf).
  - `parbft2`: Corresponds to the ParBFT2 protocol described in [ParBFT](https://eprint.iacr.org/2023/679.pdf).

- `timeout`: The delay time (in milliseconds) for the slow track in ParBFT2 to begins. This field is valid only when `type` is set to `parbft2`.
- `shortcut`: Indicates whether to enable the output shortcut of sMVBA, which is used as a black box by ParBFT's slow track.
- `rotation`: The leader selection method for the fast track. Two methods are supported:
  - `fixed`: Selects a fixed leader (replica 0).
  - `roundrobin`: Selects leaders using a Round-Robin approach.


Example configuration:

```yaml
consensus:
  type: parbft
  parbft:
    type: parbft2
    timeout: 1000
    shortcut: false
    rotation: roundrobin
```

#### Dumbo-NG protocol

The parameters for the Dumbo-NG protocol are defined in the `consensus.dumbo_ng` domain and include the following:

- `consensus_type`: Dumbo-NG consists of two concurrent modules, broadcast and consensus. The consensus module orders batches of transactions. We use the aforementioned asynchronous consensus protocols to instantiate it. Currently, seven types are supported: `tockowl`,` tockowl+`, `tockcat`, `sdumbo`, `bkr`, `parbft1`, and `parbft2`.
- `timeout`: The delay time (in milliseconds) for the slow track. This field is valid only when `type` is set to `tockowl+` or `parbft2`.
- `shortcut`: Indicates whether to enable the output shortcut. This field is valid only when `consensus_type` is set to `tockcat`, `sdumbo`, `parbft1`, or `parbft2`.
- `rotation`: The method of selecting the leader of the fast track. This field is valid only if `consensus_type` is set to `tockowl+`, `parbft1`, or `parbft2`.

Example configuration:

```yaml
consensus:
  type: dumbo_ng
  dumbo_ng:
    consensus_type: tockowl+
    timeout: 1000
    shortcut: false
    rotation: roundrobin
```

### Fault configuration

The parameters for replica faults are defined in the `fault` domain and include the following:

- `open`: Whether to enable replica faults.
- `type`: The type of fault. This parameter is valid only when `open` is set to `true`. Currently, two types are supported:
  - `crash`: Crash faults.
  - `byz`: Byzantine faults.

- fault_nodes: The IDs of the fault replicas. Multiple replica IDs can be provided. This parameter is valid only when `open` is set to `true`.

Example configuration:

```yaml
fault:
  type: crash
  open: true
  fault_nodes: [3,4,6]
```

### Run the program

After configuring the consensus and fault parameters, set the key_gen field to false:

```yaml
key_gen: false
```

Run the consensus program:

```shell
./server
```

The program will continue running until it is manually terminated using `Ctrl+C`. Upon termination, it will output three data points: throughput, latency, and the running epoch.

```shell
Throughput: 4145.866886
AvgLatency: 3212 ms
Epoch: 25
```

## Remote Experiment

This section describes how to conduct remote experiments. To perform these experiments, we need several Linux servers and a host machine. The servers are used to run the consensus replicas, while the host initiates the experiment. Note that the host can also function as one of the servers.

The following configuration steps must be completed on each Linux server:

1. Install Golang.

2. Install make.

3. Enable SFTP.

4. Enable SSH password login.

5. Download and compile the TockOwl program.

#### Configuration about servers

On the host machine, navigate to the `cmd/remote/remoteconfig` directory of the repository and edit the `remote_config.yaml` file. The server_list field in this file specifies the server information and is used to establish SSH connections. The following example demonstrates the configuration for two servers:

```yaml
server_list:
  - host: 192.168.198.138
    port: 22
    username: "server1"
    password: "123456"
  - host: 192.168.198.139
    port: 22
    username: "server2"
    password: "123456"
```

#### Configuration about remote experiment

On the host machine, enter the `cmd/remote/remoteconfig` directory and edit the `consensus_config.yaml `file. The configuration for remote experiments is defined in the remote_experiment field:

- `open`: Specifies whether to conduct a remote or local experiment.
- `time`: The duration (in seconds) for which the consensus program will run during the remote experiment. This does not include the time required for operations such as SSH connections or configuration file uploads. This field is valid only when `open` is set to `true`.

Example configuration:

```yaml
remote_experiment: 
  open: true
  time: 10
```

To execute a remote experiment, run the following command. **Do not terminate the program manually**, as it will stop automatically after the specified duration.

```shell
./remote
```

After the program ends, the relevant data will be printed to the console.

