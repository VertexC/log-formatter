package monitor

import (
	"fmt"
	"strconv"

	ctr "github.com/VertexC/log-formatter/controller"
	agentpb "github.com/VertexC/log-formatter/proto/pkg/agent"
	"github.com/VertexC/log-formatter/util"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/static"
	_ "github.com/go-sql-driver/mysql"
)

type AppConfig struct {
	ServerPort string `yaml: "serverport"`
	RpcPort    string `yaml: "rpcport"`
}

// App instance at Run time
// most recent agents information is maintained in memory
// db updates only happens when
// 1) create a new agent instance
// 2) delete a agent instance
// 3) a heartbeat comes from a new agent
type App struct {
	router      *gin.Engine
	config      *AppConfig
	agentsMap   *AgentsSyncMap
	ctr         *ctr.Controller
	heartbeatCh chan *agentpb.HeartBeat
	logger      *util.Logger
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
		} else {
			c.Next()
		}
	}
}

// NewApp
func NewApp(rpcPort string) (*App, error) {
	logger := util.NewLogger("monitor-web-server")

	router := gin.Default()
	router.Use(static.Serve("/", static.LocalFile("./build/dist", true)))
	router.Use(CORSMiddleware())

	heartbeatCh := make(chan *agentpb.HeartBeat, 1000)
	ctr := ctr.NewController(rpcPort, heartbeatCh)

	app := &App{
		router:      router,
		config:      &AppConfig {
			ServerPort: "8080",
			RpcPort: rpcPort,
		},
		ctr:         ctr,
		logger:      logger,
		heartbeatCh: heartbeatCh,
	}
	app.agentsMap = NewAgentsSyncMap()
	// register end points
	router.GET("/app", app.listAgents)
	router.GET("/agent", app.refreshAgent)
	router.PUT("/config", app.updateConfig)

	return app, nil
}

func (app *App) Start() {
	go func() {
		err := app.router.Run(":" + app.config.ServerPort)
		if err != nil {
			app.logger.Error.Fatalln(err)
		}
	}()
	// start agents Tick
	go app.agentsMap.Tick()
	// start controller
	go app.ctr.Run()
	// process heartbaet
	go func() {
		for heartbeat := range app.heartbeatCh {
			app.handleHeartBeat(heartbeat)
		}
	}()
}

// listAgents show each agent's status from database
func (app *App) listAgents(c *gin.Context) {
	data, err := app.agentsMap.ToJson()
	if err != nil {
		c.JSON(503, fmt.Sprintf("Failed to get agents: %s", err))
	} else {
		response := gin.H{"agents": string(data)}
		// TODO: render page with form
		c.JSON(200, response)
	}
}

func (app *App) refreshAgent(c *gin.Context) {
	data, err := strconv.Atoi(c.Query("id"))
	if err != nil {
		c.JSON(400, fmt.Sprintf("Invalid id %d", c.Query("id")))
		return
	}
	id := uint64(data)
	agent, err := app.agentsMap.TryGet(id)
	if err != nil {
		c.JSON(503, err)
		return
	}
	address := agent.Address
	heartbeat, err := app.ctr.GetAgentHeartBeat(address)
	if err != nil {
		defer func() {
			app.agentsMap.Update(agent)
		}()
		c.JSON(503, fmt.Sprintf("Failed to get agent heartbeat with error: %v", err))
		return
	}
	app.handleHeartBeat(heartbeat)
	agentBytes, err := app.agentsMap.AgentToJson(id)
	if err != nil {
		c.JSON(503, fmt.Sprintf("Failed to get agent data with error: %v", err))
	} else {
		response := gin.H{"agent": string(agentBytes)}
		c.JSON(200, response)
	}
}

func (app *App) updateConfig(c *gin.Context) {
	data, err := strconv.Atoi(c.Query("id"))
	if err != nil {
		c.JSON(400, fmt.Sprintf("Invalid id %d", c.Query("id")))
		return
	}
	id := uint64(data)
	agent, err := app.agentsMap.TryGet(id)
	if err != nil {
		c.JSON(503, err)
		return
	}
	address := agent.Address

	var param = struct {
		CONFIG string `json:"config" binding:"required"`
	}{}
	err = c.BindJSON(&param)
	if err != nil {
		c.JSON(400, err)
		return
	}
	config := param.CONFIG

	app.logger.Trace.Printf("Try to update agent %d with config:\n%s\n", id, config)
	r, err := app.ctr.UpdateConfig(address, []byte(config))
	if err != nil {
		c.JSON(400, err)
		return
	}
	app.handleHeartBeat(r.Heartbeat)
	c.JSON(200, "Success")
}

func (app *App) handleHeartBeat(heartbeat *agentpb.HeartBeat) {
	app.logger.Info.Printf("handleHeartbeat: %+v\n config: %v\n", *heartbeat, string(heartbeat.Config))
	agent := Agent{
		Id:      heartbeat.Id,
		Address: heartbeat.Address,
		Status:  heartbeat.Status.String(),
		Config:  string(heartbeat.Config),
	}
	app.agentsMap.Update(agent)
}
