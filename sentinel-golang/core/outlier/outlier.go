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
	"net"
	"time"

	"github.com/alibaba/sentinel-golang/logging"
)

var retryers = make(map[string]*Retryer)

type Retryer struct {
	resource            string
	recoveryInterval    time.Duration
	maxRecoveryAttempts uint32
	timers              map[string]uint32 // nodeID的已重试次数
	addresses           map[string]string
}

func getRetryerOfResource(resource string) *Retryer {
	if _, ok := retryers[resource]; !ok {
		retryer := &Retryer{}
		rules := getOutlierRulesOfResource(resource)
		if len(rules) != 0 {
			retryer.maxRecoveryAttempts = rules[0].MaxRecoveryAttempts
			retryer.recoveryInterval = time.Duration(rules[0].RecoveryInterval * 1e6)
		}
		retryers[resource] = retryer
	}
	return retryers[resource]
}

func (r *Retryer) ConnectNode(nodeID string) {
	r.timers[nodeID]++
	ok, rt := isPortOpen(r.addresses[nodeID])
	if ok {
		r.OnCompleted(nodeID, rt)
	} else {
		count := r.timers[nodeID]
		if count > r.maxRecoveryAttempts {
			count = r.maxRecoveryAttempts
		}
		time.AfterFunc(r.recoveryInterval*time.Duration(count), func() {
			r.ConnectNode(nodeID)
		})
	}
}

func (r *Retryer) scheduleRetry(nodes []string) {
	for _, node := range nodes {
		if _, ok := r.timers[node]; !ok {
			logging.Debug("r.ConnectNode ", node)
			r.ConnectNode(node)
		}
	}
}

func isPortOpen(address string) (bool, uint64) {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		return false, 0
	}
	defer conn.Close()
	end := time.Now()
	return true, uint64(end.Sub(start).Milliseconds())
}

func (r *Retryer) OnCompleted(nodeID string, rt uint64) {
	nodes := getNodeBreakersOfResource(r.resource)
	for _, breaker := range nodes[nodeID] {
		breaker.OnRequestComplete(rt, nil)
	}
}