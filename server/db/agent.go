package db

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

const HeartbeatInterval = 20

type Agent struct {
	Id     uint64 `json:"id"`
	Status string `json:"status"`
	// rpc connection address of agent
	Address string `json:"address"`
	// config is a place holder of config
	Config        string `json:"config"`
	HeartbeatTick int    `json:"alive"`
}

type AgentsSyncMap struct {
	agents map[uint64]*Agent
	lock   sync.RWMutex
}

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
		agentsMap.agents[agent.Id].HeartbeatTick = 0
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

func (agentsMap *AgentsSyncMap) ToJson() ([]byte, error) {
	agentsMap.lock.RLock()
	defer agentsMap.lock.RUnlock()
	data, err := json.Marshal(agentsMap.agents)
	return data, err
}

func (agentsMap *AgentsSyncMap) AgentToJson(id uint64) ([]byte, error) {
	agentsMap.lock.RLock()
	defer agentsMap.lock.RUnlock()
	var (
		err  error
		data []byte
	)
	if agent, ok := agentsMap.agents[id]; ok {
		data, err = json.Marshal(agent)
	} else {
		err = fmt.Errorf("Agent with Id %d not found", id)
	}
	return data, err
}

func (agentsMap *AgentsSyncMap) Tick() {
	for {
		time.Sleep(1 * time.Second)
		agentsMap.tick()
	}
}

func (agentsMap *AgentsSyncMap) tick() {
	agentsMap.lock.Lock()
	defer agentsMap.lock.Unlock()
	for _, agent := range agentsMap.agents {
		agent.HeartbeatTick++
		if agent.HeartbeatTick > HeartbeatInterval {
			agent.Status = "Unknown"
		}
	}
}
