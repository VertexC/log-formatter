package server

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/VertexC/log-formatter/config"
	ctr "github.com/VertexC/log-formatter/controller"
	agentpb "github.com/VertexC/log-formatter/proto/pkg/agent"
	"github.com/VertexC/log-formatter/server/db"
	"github.com/VertexC/log-formatter/util"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type AppConfig struct {
	Base       config.ConfigBase
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
	dbConn      *db.DBConnector
	router      *gin.Engine
	config      *AppConfig
	agentsMap   *db.AgentsSyncMap
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
func NewApp(content interface{}) (*App, error) {
	logger := util.NewLogger("WebServer")

	// set config
	contentMapStr, ok := content.(map[string]interface{})
	if !ok {
		err := fmt.Errorf("Failed to convert given config to mapStr")
		logger.Error.Printf("%s\n", err)
		return nil, err
	}

	config := &AppConfig{
		Base: config.ConfigBase{
			Content:          contentMapStr,
			MandantoryFields: []string{"serverport", "rpcport"},
		},
	}
	if err := config.Base.Validate(); err != nil {
		logger.Error.Printf("%s\n", err)
		return nil, err
	}
	util.YamlConvert(contentMapStr, config)

	// create db connection. TODO: DB reconnectin until receive a valid connection
	dbConn, err := db.NewDBConnector("test:test@tcp(127.0.0.1:3306)/logformatter")
	if err != nil {
		logger.Error.Fatalf("Failed to connect with DB: %s", err)
	}

	router := gin.Default()
	router.Use(CORSMiddleware())

	heartbeatCh := make(chan *agentpb.HeartBeat, 1000)
	ctr := ctr.NewController(config.RpcPort, heartbeatCh)

	app := &App{
		dbConn:      dbConn,
		router:      router,
		config:      config,
		ctr:         ctr,
		logger:      logger,
		heartbeatCh: heartbeatCh,
	}
	app.agentsMap = db.NewAgentsSyncMap()
	// register end points
	router.GET("/app", app.listAgents)
	router.GET("/agent", app.refreshAgent)

	return app, nil
}

func (app *App) Start() {
	app.initAgentsFromDB()
	go func() {
		err := app.router.Run(":" + app.config.ServerPort)
		if err != nil {
			app.logger.Error.Fatalln(err)
		}
	}()
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
	agents := app.agentsMap.GetAll()
	data, err := json.Marshal(agents)
	if err != nil {
		c.JSON(503, "Failed to get agents")
	} else {
		response := gin.H{"agent": string(data)}
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
			agent.Status = db.Unknown
			app.agentsMap.Update(agent)
		}()
		c.JSON(503, fmt.Sprintf("Failed to get agent heartbeat with error: %v", err))
		return
	}
	app.logger.Debug.Printf("%+v %v", *heartbeat, err)
	app.handleHeartBeat(heartbeat)
	c.JSON(200, "Success")
}

func (app *App) initAgentsFromDB() {
	agents, err := app.dbConn.GetAgentList()
	if err != nil {
		log.Fatalln("Failed to get agent list: %s", err)
	}

	for _, agent := range agents {
		log.Printf("id:%d agent:%+v\n", agent.Id, agent)
		app.agentsMap.Update(*agent)
	}
}

func (app *App) handleHeartBeat(heartbeat *agentpb.HeartBeat) {
	app.logger.Info.Printf("handleHeartbeat: %+v\n config: %v\n", *heartbeat, string(heartbeat.Config))
	agent := db.Agent{
		Id:      heartbeat.Id,
		Address: heartbeat.Address,
		Status:  db.StatusFromStr(heartbeat.Status.String()),
		Config:  string(heartbeat.Config),
	}
	app.agentsMap.Update(agent)
}
