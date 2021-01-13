package controller

import (
	"context"
	"fmt"
	"net"
	"time"

	agentpb "github.com/VertexC/log-formatter/proto/pkg/agent"
	ctrpb "github.com/VertexC/log-formatter/proto/pkg/controller"
	"github.com/VertexC/log-formatter/util"
	"google.golang.org/grpc/peer"

	"google.golang.org/grpc"
)

type HeartBeat struct {
	HeartBeat *agentpb.HeartBeat
	Addr      string
}

type Controller struct {
	ctrpb.UnimplementedControllerServer
	// agent address
	agent   string
	RpcPort string
	logger  *util.Logger
	hbCh    chan *HeartBeat
}

func NewController(port string, hbCh chan *HeartBeat) *Controller {
	logger := util.NewLogger("controller")

	return &Controller{
		logger:  logger,
		RpcPort: port,
		hbCh:    hbCh,
	}
}

func (ctr *Controller) UpdateAgentStatusRequest(ctx context.Context, heartbeat *agentpb.HeartBeat) (*ctrpb.ControllerRequestDone, error) {
	ctr.logger.Info.Printf("Get UpdateAgentStatusRequest with hearbeat: %+v\n", heartbeat)
	p, _ := peer.FromContext(ctx)
	ctr.logger.Info.Printf("Receive Heartbeat from %s, %s", p.Addr.String(), p.Addr.Network())
	ctr.hbCh <- &HeartBeat{
		HeartBeat: heartbeat,
		Addr:      getAddr(p.Addr, heartbeat.RpcPort),
	}
	res := &ctrpb.ControllerRequestDone{}
	return res, nil
}

func (ctr *Controller) UpdateConfig(rpcAddr string, configBytes []byte) (*agentpb.UpdateConfigResponse, error) {
	conn, err := grpc.Dial(rpcAddr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(time.Duration(1)*time.Second))

	if err != nil {
		err = fmt.Errorf("Can not connect: %v", err)
		ctr.logger.Error.Printf("%s\n", err)
		return nil, err
	}

	defer conn.Close()
	ctr.logger.Info.Printf("Start to Update Agent's Config\n")
	client := agentpb.NewLogFormatterAgentClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err != nil {
		ctr.logger.Error.Fatalf("could not greet: %v", err)
	}

	updateConfigRequest := &agentpb.UpdateConfigRequest{
		Config: configBytes,
	}

	r, err := client.UpdateConfig(ctx, updateConfigRequest)
	if err != nil {
		ctr.logger.Error.Printf("Failed to get response: %s\n", err)
		return nil, err
	}
	ctr.logger.Info.Printf("Got Response: %+v\n", *r)
	return r, nil
}

func (ctr *Controller) GetAgentHeartBeat(rpcAddr string) (*agentpb.HeartBeat, error) {
	// set out of time logic
	conn, err := grpc.Dial(rpcAddr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(time.Duration(1)*time.Second))

	if err != nil {
		err = fmt.Errorf("Can not connect: %v", err)
		ctr.logger.Error.Printf("%s\n", err)
		return nil, err
	}

	defer conn.Close()
	ctr.logger.Info.Printf("Start to Request Agent Status\n")
	client := agentpb.NewLogFormatterAgentClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err != nil {
		ctr.logger.Error.Fatalf("could not greet: %v", err)
	}

	heartbeatRequest := &agentpb.HeartBeatRequest{}

	r, err := client.GetHeartBeat(ctx, heartbeatRequest)
	if err != nil {
		ctr.logger.Error.Printf("Failed to get response: %s\n", err)
		return nil, err
	}
	ctr.logger.Info.Printf("Got Response: %+v\n", *r)
	return r, nil
}

func (ctr *Controller) Run() {
	ctr.startRpcService()
}

func (ctr *Controller) startRpcService() {
	port := ctr.RpcPort
	list, err := net.Listen("tcp", ":"+port)

	if err != nil {
		ctr.logger.Error.Fatalf("Failed to listen %s: %s\n", port, err)
	}
	s := grpc.NewServer()
	ctrpb.RegisterControllerServer(s, ctr)
	ctr.logger.Info.Printf("Start rpc listen: %s\n", port)
	go func() {
		if err := s.Serve(list); err != nil {
			ctr.logger.Error.Fatalf("Failed to serve: %s\n", err)
		}
	}()
}

// TODO: not sure if net.Addr.String() == ip:port
func getAddr(addr net.Addr, port string) string {
	reg := `^(?P<ip>.*?)\:(?P<port>[0-9]+)$`
	mapStr, err := util.SubMatchMapRegex(reg, addr.String())
	if err != nil {
		panic(err)
	}
	return mapStr["ip"] + ":" + port
}
