package db

import (
	"fmt"
	"log"
	"sync"
)

type Agent struct {
	Id     uint64 `json:"id"`
	Status Status `json:"status"`
	// rpc connection address of agent
	Address string `json:"address"`
}

type AgentsSyncMap struct {
	agents map[uint64]*Agent
	lock   sync.RWMutex
}

type Status int

const (
	Running Status = iota
	Stop
	Unknown
)

func NewAgentsSyncMap() *AgentsSyncMap {
	return &AgentsSyncMap{
		agents: make(map[uint64]*Agent),
	}
}

func (agentsMap *AgentsSyncMap) Update(agents ...Agent) {
	agentsMap.lock.Lock()
	defer agentsMap.lock.Unlock()
	for _, agent := range agents {
		agentsMap.agents[agent.Id] = &agent
	}
}

func (agentsMap *AgentsSyncMap) TryGet(id uint64) (Agent, error) {
	agentsMap.lock.RLock()
	defer agentsMap.lock.RUnlock()
	if agent, ok := agentsMap.agents[id]; ok {
		return *agent, nil
	} else {
		return Agent{}, fmt.Errorf("Agent with Id %d not found", id)
	}
}

func (agentsMap *AgentsSyncMap) GetAll() []Agent {
	agentsMap.lock.RLock()
	defer agentsMap.lock.RUnlock()
	agents := []Agent{}
	for _, agent := range agentsMap.agents {
		agents = append(agents, *agent)
	}
	return agents
}

func StatusFromStr(status string) (result Status) {
	switch status {
	case "Stop":
		result = Stop
	case "Running":
		result = Running
	case "Unknow":
		result = Unknown
	default:
		log.Fatalln("Invalid status: %s", status)
	}
	return
}

func (status Status) String() string {
	return []string{"Stop", "Running"}[status]
}
