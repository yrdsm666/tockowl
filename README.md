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

## Reproducing the experiments in the paper

In this section, we describe how to reproduce the experiments in the paper.

### Experimental setup

We deploy a consensus network on Amazon Web Services, using m7g.8xlarge instances across 5 different AWS regions: N. Virginia (us-east-1), Sydney (ap-southeast-2), Tokyo (ap-northeast-1), Stockholm (eu-north-1), and Frankfurt (eu-central-1). They provide 15Gbps of bandwidth, 32 virtual CPUs on AWS Graviton3 processor, and 128GB memory and run Linux Ubuntu server 22.04.

We implement the mempool of Dumbo-NG to facilitate the synchronization of data blocks among replicas. Dumbo-NG decouples the process of broadcasting and consensus. Specifically, Dumbo-NG has a broadcast module that propagates transactions among replicas, and we use this module for transaction synchronization. Dumbo-NG also has a consensus module that orders batches of transactions, and we instantiate this module with our asynchronous protocols. 

Each transaction in the mempool is set to a size of 250 bytes.

On the host machine, navigate to the `cmd/remote/remoteconfig` directory and edit the `consensus_config.yaml` file.

### Experiment 1 (Figure 5(a))

Figure 5(a) presents eight lines, each corresponding to a different consensus protocol and fault scenario:

| Lines             | Protocols | Total replicas | Faults    | Shortcut |
| ----------------- | --------- | -------------- | --------- | -------- |
| BKR (0 crash)     | BKR       | 10             | no faults | -        |
| TockOwl (0 crash) | TockOwl   | 10             | no faults | -        |
| TockCat (0 crash) | TockCat   | 10             | no faults | Yes      |
| sMVBA (0 crash)   | sMVBA     | 10             | no faults | Yes      |
| BKR (3 crash)     | BKR       | 10             | 3 crash   | -        |
| TockOwl (3 crash) | TockOwl   | 10             | 3 crash   | -        |
| TockCat (3 crash) | TockCat   | 10             | 3 crash   | Yes      |
| sMVBA (3 crash)   | sMVBA     | 10             | 3 crash   | Yes      |

Each data point in the figure represents a single experiment, and several data points together form a line.

To illustrate how the data for a line is obtained, we use the line **TockCat (3 crash)** as an example. Part of the configuration file for the first data point in this line is shown below:

```yaml
fault:
  open: true            # Replica faults exist
  type: crash           # Fault type is crash
  fault_nodes: [3,4,6]  # Replicas 3, 4, and 6 are crashed replicas
 
total: 10  # A total of 10 replicas

remote_experiment: 
  open: true  # Running a Remote Experiment
  time: 300   # The running time of the remote experiment is 300s

consensus:
  batch_size: 1000   # Each block contains 1000 transactions
  payload_size: 250  # The size of each transaction is 250 bytes
  epoch: 100         # Run 100 epochs
  type: dumbo_ng     # Dumbo-NG mempool
  dumbo_ng:
    consensus_type: tockcat  # Reach consensus on transactions in the mempool based on TockCat
    shortcut: true  # Enable output shortcut
```

After setting up the configuration file, start the experiment on the host machine:

```shell
./remote
```

After the experiment concludes, the console outputs the throughput and latency results.

To generate the data for the line corresponding to **TockOwl (3 crash)** in Figure 5(a), gradually increase the value of the `consensus.batch_size` field until the throughput reaches an extreme value. Conduct experiments for each value to obtain a series of data points that form the corresponding line. The data for the other lines in Figure 5(a) can be obtained in a similar manner.

Note that the maximum value of `batch_size` depends on the server's bandwidth and memory. If the server has limited bandwidth or memory, the `batch_size` should be set to a smaller value.

### Experiment 2 (Figure 5(b))

Figure 5(b) presents six lines, each corresponding to a different consensus protocol and fault scenario:

| Lines              | Protocols | Total replicas | Faults    | Shortcut |
| ------------------ | --------- | -------------- | --------- | -------- |
| TockOwl (0 crash)  | TockOwl   | 100            | no faults | -        |
| TockCat (0 crash)  | TockCat   | 100            | no faults | Yes      |
| sMVBA (0 crash)    | sMVBA     | 100            | no faults | Yes      |
| TockOwl (33 crash) | TockOwl   | 100            | 33 crash  | -        |
| TockCat (33 crash) | TockCat   | 100            | 33 crash  | Yes      |
| sMVBA (33 crash)   | sMVBA     | 100            | 33 crash  | Yes      |

To illustrate how the data for a line is obtained, we use the line **sMVBA (33 crash)** as an example. Part of the configuration file for the first data point in this line is shown below:

```yaml
fault:
  open: true            # Replica faults exist
  type: crash           # Fault type is crash
  fault_nodes: [3, 6, 9, 12, 15, 18, 21, 24, 27, 30, 33, 36, 39, 42, 45, 48, 51, 54, 57, 60, 63, 66, 69, 72, 75, 78, 81, 84, 87, 90, 93, 96, 99] # These replicas are crashed replicas
 
total: 100  # A total of 100 replicas

remote_experiment: 
  open: true  # Running a Remote Experiment
  time: 300   # The running time of the remote experiment is 300s

consensus:
  batch_size: 100   # Each block contains 100 transactions
  payload_size: 250  # The size of each transaction is 250 bytes
  epoch: 100         # Run 100 epochs
  type: dumbo_ng     # Dumbo-NG mempool
  dumbo_ng:
    consensus_type: sdumbo  # Reach consensus on transactions in the mempool based on sMVBA
    shortcut: true  # Enable output shortcut
```

To generate the data for the line corresponding to **sMVBA (33 crash)** in Figure 5(b), gradually increase the value of the `consensus.batch_size` field until the throughput reaches an extreme value. Conduct experiments for each value to obtain a series of data points that form the corresponding line. The data for the other lines in Figure 5(b) can be obtained in a similar manner.

### Experiment 3 (Figure 6)

Figure 6 presents seven lines, each corresponding to a different consensus protocol and fault scenario:

1. BKR (10 replicas, 3 faulty): total number of replicas 10, 3 Byzantine replicas, BKR protocol;

2. TockOwl (10 replicas, 3 faulty): total number of replicas 10, 3 Byzantine replicas, TockOwl protocol;

3. TockCat (10 replicas, 3 faulty): total number of replicas 10, 3 Byzantine replicas, TockCat protocol, output shortcut disabled;

4. sMVBA (10 replicas, 3 faulty): total number of replicas 10, 3 Byzantine replicas, sMVBA protocol, output shortcut disabled;

5. TockOwl (100 replicas, 33 faulty): total number of replicas 100, 33 Byzantine replicas, TockOwl protocol;

6. TockCat (100 replicas, 33 faulty): 100 total replicas, 33 Byzantine replicas, TockCat protocol, output shortcut disabled;
7. sMVBA (100 replicas, 33 faulty): 100 total replicas, 33 Byzantine replicas, sMVBA protocol, output shortcut disabled;

| Lines                             | Protocols | Total replicas | Faults       | Shortcut |
| --------------------------------- | --------- | -------------- | ------------ | -------- |
| BKR (10 replicas, 3 faulty)       | BKR       | 10             | 3 Byzantine  | -        |
| TockOwl (10 replicas, 3 faulty)   | TockOwl   | 10             | 3 Byzantine  | -        |
| TockCat (10 replicas, 3 faulty)   | TockCat   | 10             | 3 Byzantine  | No       |
| sMVBA (10 replicas, 3 faulty)     | sMVBA     | 10             | 3 Byzantine  | No       |
| TockOwl (100 replicas, 33 faulty) | TockOwl   | 100            | 33 Byzantine | -        |
| TockOwl (100 replicas, 33 faulty) | TockCat   | 100            | 33 Byzantine | No       |
| sMVBA (100 replicas, 33 faulty)   | sMVBA     | 100            | 33 Byzantine | No       |

To illustrate how the data for a line is obtained, we use the line **TockOwl (100 replicas, 33 faulty)** as an example. Part of the configuration file for the first data point in this line is shown below:

```yaml
byzantine:
  open: true     # Replica faults exist
  type: byz    # Fault type is Byzantine
  fault_nodes: [3, 6, 9, 12, 15, 18, 21, 24, 27, 30, 33, 36, 39, 42, 45, 48, 51, 54, 57, 60, 63, 66, 69, 72, 75, 78, 81, 84, 87, 90, 93, 96, 99] # These replicas are Byzantine replicas

total: 100  # A total of 100 replicas

remote_experiment: 
  open: true  # Running a Remote Experiment
  time: 300   # The running time of the remote experiment is 300s

consensus:
  batch_size: 100    # Each block contains 100 transactions
  payload_size: 250  # The size of each transaction is 250 bytes
  epoch: 100         # Run 100 epochs
  type: dumbo_ng     # Dumbo-NG mempool
  dumbo_ng:
    consensus_type: tockowl  # Reach consensus on transactions in the mempool based on TockOwl
```

To generate the data for the line corresponding to **TockOwl (100 replicas, 33 faulty)** in Figure 6, gradually increase the value of the `consensus.batch_size` field until the throughput reaches an extreme value. Conduct experiments for each value to obtain a series of data points that form the corresponding line. The data for the other lines in Figure 6 can be obtained in a similar manner.

### Experiment 4 (Figure 7)

Figure 7 presents eight lines, each corresponding to a different consensus protocol and fault scenario:

| Lines                                | Protocols | Total replicas | Leader | Faults         | Shortcut |
| ------------------------------------ | --------- | -------------- | ------ | -------------- | -------- |
| TockOwl+ (10 replicas, good-leader)  | TockOwl+  | 10             | fixed  | no faults      | -        |
| TockOwl+ (10 replicas, bad-leader)   | TockOwl+  | 10             | fixed  | leader crashes | -        |
| ParBFT (10 replicas, good-leader)    | ParBFT1   | 10             | fixed  | no faults      | Yes      |
| ParBFT (10 replicas, bad-leader)     | ParBFT1   | 10             | fixed  | leader crashes | Yes      |
| TockOwl+ (100 replicas, good-leader) | TockOwl+  | 100            | fixed  | no faults      | -        |
| TockOwl+ (100 replicas, bad-leader)  | TockOwl+  | 100            | fixed  | leader crashes | -        |
| ParBFT (100 replicas, good-leader)   | ParBFT1   | 100            | fixed  | no faults      | Yes      |
| ParBFT (100 replicas, bad-leader)    | ParBFT1   | 100            | fixed  | leader crashes | Yes      |

To illustrate how the data for a line is obtained, we use the line **ParBFT (10 replicas, bad-leader)** as an example. Part of the configuration file for the first data point in this line is shown below:

```yaml
fault:
  open: true       # Replica faults exist
  type: crash      # Fault type is crash
  fault_nodes: [0] # # Only the leader (replicas 0) is crashed

total: 10  # A total of 10 replicas

remote_experiment: 
  open: true  # Running a Remote Experiment
  time: 300   # The running time of the remote experiment is 300s

consensus:
  batch_size: 1000   # Each block contains 1000 transactions
  payload_size: 250  # The size of each transaction is 250 bytes
  epoch: 100         # Run 100 epochs
  type: dumbo_ng     # Dumbo-NG mempool
  dumbo_ng:
    consensus_type: parbft1  # Reach consensus on transactions in the mempool based on ParBFT1
    shortcut: true   # Enable output shortcut    
    rotation: fixed  # The method for selecting leaders on the fast track is Fixed
```

To generate the data for the line corresponding to **ParBFT (10 replicas, bad-leader)** in Figure 7, gradually increase the value of the `consensus.batch_size` field until the throughput reaches an extreme value. Conduct experiments for each value to obtain a series of data points that form the corresponding line. The data for the other lines in Figure 7 can be obtained in a similar manner.

### Experiment 5 (Figure 8)

Figure 8 presents four lines, each corresponding to a different consensus protocol and fault scenario:

| Lines                             | Protocols | Total replicas | Leader | Faults                                    | Shortcut |
| --------------------------------- | --------- | -------------- | ------ | ----------------------------------------- | -------- |
| TockOwl+ (10 replicas, 3 crash)   | TockOwl+  | 10             | fixed  | 3 replicas crashed, including the leader  | -        |
| ParBFT (10 replicas, 3 crash)     | ParBFT1   | 10             | fixed  | 3 replicas crashed, including the leader  | No       |
| TockOwl+ (100 replicas, 33 crash) | TockOwl+  | 100            | fixed  | 33 replicas crashed, including the leader | -        |
| TockOwl+ (100 replicas, 33 crash) | ParBFT1   | 100            | fixed  | 33 replicas crashed, including the leader | No       |

To illustrate how the data for a line is obtained, we use the line **TockOwl+ (100 replicas, 33 crash)** as an example. Part of the configuration file for the first data point in this line is shown below:

```yaml
fault:
  open: true       # Replica faults exist
  type: crash      # Fault type is crash
  fault_nodes: [0, 6, 9, 12, 15, 18, 21, 24, 27, 30, 33, 36, 39, 42, 45, 48, 51, 54, 57, 60, 63, 66, 69, 72, 75, 78, 81, 84, 87, 90, 93, 96, 99] # These replicas are Byzantine replicas, including the leader (replica 0) of the fast track

total: 100  # A total of 100 replicas

remote_experiment: 
  open: true  # Running a Remote Experiment
  time: 300   # The running time of the remote experiment is 300s

consensus:
  batch_size: 100    # Each block contains 100 transactions
  payload_size: 250  # The size of each transaction is 250 bytes
  epoch: 100         # Run 100 epochs
  type: dumbo_ng     # Dumbo-NG mempool
  dumbo_ng:
    consensus_type: tockowl+  # Reach consensus on transactions in the mempool based on TockOwl+
    rotation: fixed  # The method for selecting leaders on the fast track is Fixed
```

To generate the data for the line corresponding to **TockOwl+ (100 replicas, 33 crash)** in Figure 8, gradually increase the value of the `consensus.batch_size` field until the throughput reaches an extreme value. Conduct experiments for each value to obtain a series of data points that form the corresponding line. The data for the other lines in Figure 8 can be obtained in a similar manner.

### Experiment 6 (Figure 9)

Figure 9 presents four lines, each corresponding to a different consensus protocol and fault scenario:

| Lines              | Protocols | Total replicas | Leader | Faults    | Timeout | Shortcut |
| ------------------ | --------- | -------------- | ------ | --------- | ------- | -------- |
| TockOwl+           | TockOwl+  | 100            | fixed  | no faults | -       | -        |
| ParBFT             | ParBFT1   | 100            | fixed  | no faults | -       | Yes      |
| TockOwl+ (1000 ms) | TockOwl+  | 100            | fixed  | no faults | 1000 ms | -        |
| ParBFT (1000 ms)   | ParBFT2   | 100            | fixed  | no faults | 1000 ms | Yes      |

To illustrate how the data for a line is obtained, we use the line **ParBFT (1000 ms)** as an example. Part of the configuration file for the first data point in this line is shown below:

```yaml
p2p:
  type: grpc      # Network type is grpc
  extra_delay: 0  # The additional delay for message transmission is 0 ms

fault:
  open: false       # No faults
  type: crash       # This field is valid only when open is set to true
  fault_nodes: [1]  # This field is valid only when open is set to true

total: 100  # A total of 100 replicas

remote_experiment: 
  open: true  # Running a Remote Experiment
  time: 300   # The running time of the remote experiment is 300s

consensus:
  batch_size: 1000   # Each block contains 1000 transactions
  payload_size: 250  # The size of each transaction is 250 bytes
  epoch: 100         # Run 100 epochs
  type: dumbo_ng     # Dumbo-NG mempool
  dumbo_ng:
    consensus_type: parbft2  # Reach consensus on transactions in the mempool based on ParBFT2
    timeout: 1000    # The slow track will not start until 1000 ms later
    rotation: fixed  # The method for selecting leaders on the fast track is Fixed
```

To generate the data for the line corresponding to **ParBFT (1000 ms)** in Figure 9, modify the `p2p.extra_delay` field to values of 500, 1,000, 1,500, 2,000, 2,500, 3,000, 3,500. Conduct experiments for each value to obtain a series of data points that form the corresponding line. The data for the other lines in Figure 9 can be obtained in a similar manner.
