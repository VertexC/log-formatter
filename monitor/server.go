package monitor

import (
	"fmt"
	"hash/fnv"
	"strconv"

	ctr "github.com/VertexC/log-formatter/controller"
	"github.com/VertexC/log-formatter/util"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type AppConfig struct {
	WebPort string `yaml: "webport"`
	RpcPort string `yaml: "rpcport"`
}

// App instance at Run time
// most recent agents information is maintained in memory
// db updates only happens when
// 1) create a new agent instance
// 2) delete a agent instance
// 3) a heartbeat comes from a new agent
type App struct {
	router    *gin.Engine
	config    *AppConfig
	agentsMap *AgentsSyncMap
	ctr       *ctr.Controller
	hbCh      chan *ctr.HeartBeat
	logger    *util.Logger
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
func NewApp(rpcPort string, webPort string) (*App, error) {
	logger := util.NewLogger("monitor-web-server")

	router := gin.Default()
	router.Use(static.Serve("/", static.LocalFile("./build/dist", true)))
	router.Use(CORSMiddleware())

	hbCh := make(chan *ctr.HeartBeat, 1000)
	controller := ctr.NewController(rpcPort, hbCh)

	app := &App{
		router: router,
		config: &AppConfig{
			WebPort: webPort,
			RpcPort: rpcPort,
		},
		ctr:    controller,
		logger: logger,
		hbCh:   hbCh,
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
		err := app.router.Run(":" + app.config.WebPort)
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
		for hb := range app.hbCh {
			app.handleHeartBeat(hb)
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
		c.JSON(200, response)
	}
}

func (app *App) refreshAgent(c *gin.Context) {
	data, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil {
		app.logger.Error.Printf("Failed to get Id: %s", err)
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
		c.JSON(503, fmt.Sprintf("Failed to get agent heartbeat with error: %v", err))
		return
	}
	app.handleHeartBeat(&ctr.HeartBeat{
		HeartBeat: heartbeat,
		Addr:      address,
	})
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
	app.handleHeartBeat(&ctr.HeartBeat{
		HeartBeat: r.Heartbeat,
		Addr:      address,
	})
	c.JSON(200, "Success")
}

func (app *App) handleHeartBeat(hb *ctr.HeartBeat) {
	app.logger.Info.Printf("handleHeartbeat: %+v\n config: %v\n", *hb, string(hb.HeartBeat.Config))
	agent := Agent{
		Id:      hash(hb.Addr),
		Address: hb.Addr,
		Status:  hb.HeartBeat.Status.String(),
		Config:  string(hb.HeartBeat.Config),
	}
	app.agentsMap.Update(agent)
}

func hash(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}
