package clientcore

import (
	"context"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yrdsm666/tockowl/config"
	"github.com/yrdsm666/tockowl/logging"
	"github.com/yrdsm666/tockowl/pb"
	"github.com/yrdsm666/tockowl/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
)

var logger = logging.GetLogger()

const timeoutDuration = time.Second * 2

type command string

type Client struct {
	consensusResult        map[command]reply
	config                 *config.Config
	requestTimeout         *utils.Timer
	requestTimeoutDuration time.Duration
	replyChan              chan *pb.Msg
	closeChan              chan int
	p2pClient              pb.P2PClient
	wg                     *sync.WaitGroup
	mut                    *sync.Mutex
	log                    *logrus.Entry

	pendingTx *atomic.Int64
	successTx *atomic.Int64

	// latency
	sendTime    map[command]time.Time
	sendTimeMut *sync.Mutex

	pb.UnimplementedP2PServer
}

type reply struct {
	result string
	count  int
}

// receive packet from consensus
func (gc *Client) Send(ctx context.Context, in *pb.Packet) (*pb.Empty, error) {
	msg := new(pb.Msg)
	err := proto.Unmarshal(in.Msg, msg)
	if err != nil {
		logger.WithError(err).Warn("received msg error")
		return &pb.Empty{}, nil
	}
	gc.replyChan <- msg
	return &pb.Empty{}, nil
}

func NewClient(log *logrus.Entry) *Client {
	cfg, err := config.NewConfig("config/config.yaml", 0, "")
	utils.PanicOnError(err)
	conn, err := grpc.Dial(cfg.Nodes[0].RpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	p2pClient := pb.NewP2PClient(conn)
	// rand.Seed(time.Now().UnixNano())

	client := &Client{
		consensusResult:        make(map[command]reply),
		config:                 cfg,
		requestTimeoutDuration: timeoutDuration,
		requestTimeout:         utils.NewTimer(timeoutDuration),
		replyChan:              make(chan *pb.Msg),
		closeChan:              make(chan int),
		p2pClient:              p2pClient,
		wg:                     new(sync.WaitGroup),
		mut:                    new(sync.Mutex),
		log:                    log,
		pendingTx:              new(atomic.Int64),
		successTx:              new(atomic.Int64),
		UnimplementedP2PServer: pb.UnimplementedP2PServer{},
		sendTime:               make(map[command]time.Time),
		sendTimeMut:            new(sync.Mutex),
	}

	client.requestTimeout.Init()
	client.requestTimeout.Stop()
	return client
}

func (client *Client) getResults(cmd command) (re reply, b bool) {
	client.mut.Lock()
	defer client.mut.Unlock()
	re, b = client.consensusResult[cmd]
	return
}

func (client *Client) setResult(cmd command, re reply) {
	client.mut.Lock()
	defer client.mut.Unlock()
	client.consensusResult[cmd] = re
}

func (client *Client) getSendTime(cmd command) (st time.Time, b bool) {
	client.sendTimeMut.Lock()
	defer client.sendTimeMut.Unlock()
	st, b = client.sendTime[cmd]
	return
}

func (client *Client) setSendTime(cmd command, st time.Time) {
	client.sendTimeMut.Lock()
	defer client.sendTimeMut.Unlock()
	client.sendTime[cmd] = st
}

func (client *Client) receiveReply() {
	var totalReceive int
	var numLock sync.Mutex
	var timeLock sync.Mutex
	var totalLatencyTime time.Duration
	numLock.Lock()
	totalReceive = 0
	numLock.Unlock()
	timeLock.Lock()
	totalLatencyTime = 0
	timeLock.Unlock()
	for {
		select {
		case msg := <-client.replyChan:
			replyMsg := msg.GetReply()
			client.log.Debug("got reply message", string(replyMsg.Tx))
			tx := new(pb.Transaction)
			if proto.Unmarshal(replyMsg.Tx, tx) != nil {
				continue
			}
			cmd := command(tx.Payload)
			if re, ok := client.getResults(cmd); ok {
				if re.result == string(replyMsg.Receipt) {
					re.count++
					if re.count == (len(client.config.Nodes)+2)/3 {
						client.pendingTx.Add(-1)
						client.successTx.Add(1)
						client.log.WithFields(logrus.Fields{
							"cmd":     cmd,
							"result":  re.result,
							"total":   client.pendingTx.Load() + client.successTx.Load(),
							"pending": client.pendingTx.Load(),
							"success": client.successTx.Load(),
						}).Info("Consensus success.")

						var latency time.Duration
						if sTime, ok := client.getSendTime(cmd); !ok {
							client.log.Warn("the reply message has not been sent")
							continue
						} else {
							latency = time.Since(sTime)
							numLock.Lock()
							totalReceive += 1
							numLock.Unlock()
							timeLock.Lock()
							// totalLatencyTime += time.Duration(int64(latency) * int64(num))
							totalLatencyTime += latency
							timeLock.Unlock()
						}
					}
					client.setResult(cmd, re)
				}
			} else {
				re := reply{
					result: string(replyMsg.Receipt),
					count:  1,
				}
				client.setResult(cmd, re)
			}
		case <-client.closeChan:
			client.log.Info("The total number of transactions received is: ", totalReceive)
			client.log.Info("The average latency is: ", totalLatencyTime/time.Duration(totalReceive))
			client.wg.Done()
			return
		}
	}
}

func (client *Client) sendTx(tx *pb.Transaction) {
	btx, err := proto.Marshal(tx)
	utils.PanicOnError(err)
	request := &pb.Request{Tx: btx}
	rawRequest, err := proto.Marshal(request)
	utils.PanicOnError(err)
	packet := &pb.Packet{
		Msg:         rawRequest,
		ConsensusID: -1,
		Epoch:       -1,
		Type:        pb.PacketType_CLIENTPACKET,
	}
	client.pendingTx.Add(1)
	client.log.WithFields(logrus.Fields{
		"total":   client.pendingTx.Load() + client.successTx.Load(),
		"pending": client.pendingTx.Load(),
		"success": client.successTx.Load(),
		"tx":      string(tx.Payload),
	}).Info("sending tx")
	_, err = client.p2pClient.Send(context.Background(), packet)
	utils.LogOnError(err, "send packet failed", client.log)

	cmd := command(tx.Payload)
	client.setSendTime(cmd, time.Now())
}

func (client *Client) NormalRun() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM,
		syscall.SIGQUIT)
	client.wg.Add(2)

	// use goroutine send msg
	go func() {
		ticker := time.NewTicker(1000 * time.Millisecond)
		for {
			select {
			case <-client.closeChan:
				client.wg.Done()
				return
			case <-ticker.C:
				innerTx := strconv.Itoa(rand.Intn(100)) + "," + strconv.Itoa(rand.Intn(100))
				//client.log.WithField("content", cmd).Info("[CLIENT] Send request")
				tx := &pb.Transaction{Type: pb.TransactionType_NORMAL, Payload: []byte(innerTx)}
				client.sendTx(tx)
			}
		}
	}()

	// start client server
	clientServer := grpc.NewServer()
	pb.RegisterP2PServer(clientServer, client)
	listen, err := net.Listen("tcp", "localhost:9999")
	if err != nil {
		panic(err)
	}
	go client.receiveReply()
	go clientServer.Serve(listen)
	<-c
	client.log.Info("[CLIENT] Client exit...")
	clientServer.Stop()
}

func (client *Client) Stop() {
	close(client.closeChan)
	client.wg.Wait()
}
