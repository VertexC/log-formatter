package controller

import (
	"context"
	"net"

	agentpb "github.com/VertexC/log-formatter/proto/pkg/agent"
	ctrpb "github.com/VertexC/log-formatter/proto/pkg/controller"
	"github.com/VertexC/log-formatter/util"

	"google.golang.org/grpc"
)

// TODO: consume heartbeat with message queue
type Controller struct {
	ctrpb.UnimplementedControllerServer
	// agent address
	agent   string
	RpcPort string
	logger  *util.Logger
}

func NewController(port string) *Controller {
	logger := util.NewLogger("controller")

	// FIXME: hardcode for now
	return &Controller{
		logger:  logger,
		RpcPort: port,
	}
}

func (ctr *Controller) UpdateAgentStatusRequest(c context.Context, heartbeat *agentpb.HeartBeat) (*ctrpb.ControllerRequestDone, error) {
	ctr.logger.Info.Printf("Get UpdateAgentStatusRequest with hearbeat: %+v\n", heartbeat)
	res := &ctrpb.ControllerRequestDone{}
	return res, nil
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

// func main() {
// 	// Set up a connection to the server.
// 	conn, err := grpc.Dial("localhost:2001", grpc.WithInsecure(), grpc.WithBlock())
// 	if err != nil {
// 		log.Fatalf("did not connect: %v", err)
// 	}
// 	defer conn.Close()
// 	c := agentpb.NewLogFormatterAgentClient(conn)

// 	// Contact the server and print out its response.
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
// 	defer cancel()
// 	heartbeatRequest := &agentpb.HeartBeatRequest{}
// 	r, err := c.GetHeartBeat(ctx, heartbeatRequest)
// 	if err != nil {
// 		log.Fatalf("could not greet: %v", err)
// 	}
// 	log.Printf("Got Heartbeat: %+v\n", *r)
// }
