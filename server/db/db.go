package db

import (
	"log"

	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type DBConnector struct {
	db *sql.DB
}

type Agent struct {
	Id uint64
	Status Status
	Address string
}

type Status int

const (
	Running Status = iota
	Stop
)

func StatusFromStr(status string) (result Status) {
	switch status {
	case "STOP":
		result = Stop
	case "Running":
		result = Running
	default:
		log.Fatalln("Invalid status")
	}
	return
}

func NewDBConnector (url string) (*DBConnector, error) {
	db, err := sql.Open("mysql", url)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	connector := &DBConnector {
		db: db,
	}
	return connector, nil
}

func (connector *DBConnector) Close() {
	defer connector.db.Close()
}

// GetAgentList get all info of agent from database
func (connector *DBConnector) GetAgentList() ([]Agent, error) {
	db := connector.db
	rows, err := db.Query("select id, address, status from agent")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	agents := []Agent {}
	for rows.Next() {
		agent := Agent{}
		var status string
		err := rows.Scan(&agent.Id, &agent.Address, &status)
		if err != nil {
			log.Fatal(err)
		}
		agent.Status = StatusFromStr(status)
		log.Printf("Get Agent from db: %+v", agent)
		agents = append(agents, agent)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return agents, nil
}

// UpdateAgent update database with given agent information
func (connector *DBConnector) UpdateAgent(agent Agent) error {
	return nil
}
