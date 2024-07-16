// Copyright 1999-2020 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package outlier

import (
	"sync"

	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
)

var (
	breakers  = make(map[string]map[string][]circuitbreaker.CircuitBreaker)
	updateMux = new(sync.RWMutex)
)

func getNodesOfResource(resource string) map[string][]circuitbreaker.CircuitBreaker {
	updateMux.RLock()
	nodes := breakers[resource]
	updateMux.RUnlock()
	ret := make(map[string][]circuitbreaker.CircuitBreaker, len(nodes))
	for nodeID, val := range nodes {
		ret[nodeID] = val
	}
	return ret
}
