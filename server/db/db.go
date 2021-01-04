package db

import (
	"fmt"
	"log"

	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type DBConnector struct {
	db *sql.DB
}

func NewDBConnector(url string) (*DBConnector, error) {
	db, err := sql.Open("mysql", url)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	connector := &DBConnector{
		db: db,
	}
	return connector, nil
}

func (connector *DBConnector) Close() {
	defer connector.db.Close()
}

// GetAgentList get all info of agent from database
func (connector *DBConnector) GetAgentList() ([]*Agent, error) {
	db := connector.db
	rows, err := db.Query("select id, address, status from agent")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	agents := []*Agent{}
	for rows.Next() {
		agent := new(Agent)
		var status string
		err := rows.Scan(&agent.Id, &agent.Address, &status)
		if err != nil {
			log.Fatal(err)
		}
		agent.Status = status
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

func (connector *DBConnector) AddAgent(agent Agent) error {
	sql := fmt.Sprintf(
		"insert into agent(id, address, status) values (%d, '%s', '%s')",
		agent.Id,
		agent.Address,
		agent.Status,
	)
	_, err := connector.db.Exec(sql)
	if err != nil {
		log.Println("exec failed:", err, ", sql:", sql)
		return err
	}
	log.Println("Add Agent Asucess")
	return nil
}
