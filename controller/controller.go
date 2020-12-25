package controller

import (
	"context"
	"fmt"
	"net"
	"time"

	agentpb "github.com/VertexC/log-formatter/proto/pkg/agent"
	ctrpb "github.com/VertexC/log-formatter/proto/pkg/controller"
	"github.com/VertexC/log-formatter/util"

	"google.golang.org/grpc"
)

// TODO: consume heartbeat with message queue
type Controller struct {
	ctrpb.UnimplementedControllerServer
	// agent address
	agent       string
	RpcPort     string
	logger      *util.Logger
	heartbeatCh chan *agentpb.HeartBeat
}

func NewController(port string, heartbeatCh chan *agentpb.HeartBeat) *Controller {
	logger := util.NewLogger("controller")

	// FIXME: hardcode for now
	return &Controller{
		logger:      logger,
		RpcPort:     port,
		heartbeatCh: heartbeatCh,
	}
}

func (ctr *Controller) UpdateAgentStatusRequest(c context.Context, heartbeat *agentpb.HeartBeat) (*ctrpb.ControllerRequestDone, error) {
	ctr.logger.Info.Printf("Get UpdateAgentStatusRequest with hearbeat: %+v\n", heartbeat)
	ctr.heartbeatCh <- heartbeat
	res := &ctrpb.ControllerRequestDone{}
	return res, nil
}

func (ctr *Controller) GetAgentHeartBeat(rpcAddr string) (*agentpb.HeartBeat, error) {
	var (
		conn *grpc.ClientConn
		err  error
	)
	// FIXME: harcoded agent rpc address for now
	// set out of time logic
	conn, err = grpc.Dial(rpcAddr, grpc.WithInsecure(), grpc.WithBlock())

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
	} else {
		ctr.logger.Info.Printf("Got Response: %+v\n", *r)
	}

	return r, err
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
