package server

import (
	"log"

	"github.com/VertexC/log-formatter/server/db"
	"github.com/VertexC/log-formatter/controller"
	"github.com/VertexC/log-formatter/util"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type AppConfig struct {
	Port string
}

// App instance at Run time
type App struct {
	dbConn *db.DBConnector
	router *gin.Engine
	config *AppConfig
	agents map[uint64]db.Agent
	ctr *controller.Controller
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

// TODO: add configuration control
// NewApp
func NewApp() (*App, error) {
	logger := util.NewLogger("WebServer")
	// update and reset agents
	dbConn, err := db.NewDBConnector("test:test@tcp(127.0.0.1:3306)/logformatter")
	if err != nil {
		log.Fatalf("Failed to connect with DB: %s", err)
	}

	router := gin.Default()

	router.Use(CORSMiddleware())

	ctr := controller.NewController(nil)

	app := &App{
		dbConn: dbConn,
		router: router,
		config: &AppConfig{
			Port: ":8080",
		},
		ctr: ctr,
		logger: logger,
	}
	// register end points
	router.GET("/app", app.listAgents)

	return app, nil
}

func (app *App) Start() {
	go app.router.Run(app.config.Port)
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
