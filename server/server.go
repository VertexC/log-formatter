package server

import (
	"fmt"
	"log"

	"github.com/VertexC/log-formatter/config"
	"github.com/VertexC/log-formatter/controller"
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
type App struct {
	dbConn *db.DBConnector
	router *gin.Engine
	config *AppConfig
	agents map[uint64]db.Agent
	ctr    *controller.Controller
	logger *util.Logger
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

	ctr := controller.NewController(config.RpcPort)

	app := &App{
		dbConn: dbConn,
		router: router,
		config: config,
		ctr:    ctr,
		logger: logger,
	}
	// register end points
	router.GET("/app", app.listAgents)

	// FIXME: change to POST later
	return app, nil
}

func (app *App) Start() {
	go func() {
		err := app.router.Run(":" + app.config.ServerPort)
		if err != nil {
			app.logger.Error.Fatalln(err)
		}
	}()
	go app.ctr.Run()
}

func (app *App) listAgents(c *gin.Context) {
	agents, err := app.dbConn.GetAgentList()
	if err != nil {
		log.Fatalln("Failed to get agent list: %s", err)
	}
	app.agents = make(map[uint64]db.Agent)

	for _, agent := range agents {
		log.Printf("id:%d agent:%+v\n", agent.Id, agent)
		app.agents[agent.Id] = agent
	}
	response := gin.H{"agent": app.agents}

	// TODO: render page with form
	c.JSON(200, response)
}

func (app *App) updateAgents(c *gin.Context) {
	// TODO: replace file
}
