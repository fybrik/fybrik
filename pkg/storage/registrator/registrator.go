// Copyright 2023 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package registrator

import (
	"errors"
	"sync"

	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/storage/registrator/agent"
)

var (
	agentsMu sync.RWMutex
	// map of implementation agents by the connection types they support
	agents = make(map[taxonomy.ConnectionType]agent.AgentInterface)
)

// register a new implementation agent
func Register(worker agent.AgentInterface) error {
	// mutex for the writing operation
	agentsMu.Lock()
	defer agentsMu.Unlock()
	if worker == nil {
		return errors.New("attemting to register a nil object")
	}
	key := worker.GetConnectionType()
	if _, dup := agents[key]; dup {
		return errors.New("attempting to register an agent for an existing connection type")
	}
	agents[key] = worker
	return nil
}

// return the appropriate agent
func GetAgent(key taxonomy.ConnectionType) (agent.AgentInterface, error) {
	worker, exists := agents[key]
	if !exists {
		return nil, errors.New("unsupported connection type " + string(key))
	}
	return worker, nil
}

// return the registered connection types
func GetRegisteredTypes() []taxonomy.ConnectionType {
	res := make([]taxonomy.ConnectionType, 0)
	for connType := range agents {
		res = append(res, connType)
	}
	return res
}
