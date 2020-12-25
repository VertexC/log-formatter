package db

import (
	"log"
)

type Agent struct {
	Id     uint64 `json:"id"`
	Status Status `json:"status"`
	// rpc connection address of agent
	Address string `json:"address"`
}

type Status int

const (
	Running Status = iota
	Stop
)

func StatusFromStr(status string) (result Status) {
	switch status {
	case "Stop":
		result = Stop
	case "Running":
		result = Running
	default:
		log.Fatalln("Invalid status: %s", status)
	}
	return
}

func (status Status) String() string {
	return []string{"Stop", "Running"}[status]
}
