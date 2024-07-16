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
	"fmt"
	"reflect"
	"sync"

	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
)

var (
	breakerRules  = make(map[string][]*circuitbreaker.Rule)
	breakers      = make(map[string]map[string][]circuitbreaker.CircuitBreaker)
	updateMux     = new(sync.RWMutex)
	currentRules  = make(map[string][]*circuitbreaker.Rule, 0)
	updateRuleMux = new(sync.Mutex)
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

// LoadRules replaces old rules with the given outlier ejection rules.
//
// return value:
//
// bool: was designed to indicate whether the internal map has been changed
// error: was designed to indicate whether occurs the error.
func LoadRules(rules []*Rule) (bool, error) {
	circuitRules := make([]*circuitbreaker.Rule, len(rules))
	for i, rule := range rules {
		circuitRules[i] = rule.Rule
	}

	resRulesMap := make(map[string][]*circuitbreaker.Rule, 16)
	for _, rule := range circuitRules {
		resRules, exist := resRulesMap[rule.Resource]
		if !exist {
			resRules = make([]*circuitbreaker.Rule, 0, 1)
		}
		resRulesMap[rule.Resource] = append(resRules, rule)
	}

	updateRuleMux.Lock()
	defer updateRuleMux.Unlock()
	isEqual := reflect.DeepEqual(currentRules, resRulesMap)
	if isEqual {
		logging.Info("[CircuitBreaker] Load rules is the same with current rules, so ignore load operation.")
		return false, nil
	}

	err := onRuleUpdate(resRulesMap)
	return true, err
}

// Concurrent safe to update rules
func onRuleUpdate(rawResRulesMap map[string][]*circuitbreaker.Rule) (err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%+v", r)
			}
		}
	}()
	// ignore invalid rules
	validResRulesMap := make(map[string][]*circuitbreaker.Rule, len(rawResRulesMap))
	for res, rules := range rawResRulesMap {
		validResRules := make([]*circuitbreaker.Rule, 0, len(rules))
		for _, rule := range rules {
			if err := circuitbreaker.IsValidRule(rule); err != nil {
				logging.Warn("[CircuitBreaker onRuleUpdate] Ignoring invalid circuit breaking rule when loading new rules", "rule", rule, "err", err.Error())
				continue
			}
			validResRules = append(validResRules, rule)
		}
		if len(validResRules) > 0 {
			validResRulesMap[res] = validResRules
		}
	}
	currentRules = rawResRulesMap
	updateMux.Lock()
	breakerRules = validResRulesMap
	updateMux.Unlock()

	updateAllBreakers()
	return nil
}

func updateAllBreakers() {
	start := util.CurrentTimeNano()
	updateMux.RLock()
	breakersClone := make(map[string]map[string][]circuitbreaker.CircuitBreaker, len(breakerRules))
	for res, val := range breakers {
		breakersClone[res] = make(map[string][]circuitbreaker.CircuitBreaker)
		for nodeID, tcs := range val {
			resTcClone := make([]circuitbreaker.CircuitBreaker, 0, len(tcs))
			resTcClone = append(resTcClone, tcs...)
			breakersClone[res][nodeID] = resTcClone
		}
	}
	updateMux.RUnlock()

	newBreakers := make(map[string]map[string][]circuitbreaker.CircuitBreaker, len(breakerRules))
	for res, resRules := range breakerRules {
		for nodeID, tcs := range breakersClone[res] {
			newCbsOfRes := circuitbreaker.BuildResourceCircuitBreaker(res, resRules, tcs)
			if len(newCbsOfRes) > 0 {
				newBreakers[res][nodeID] = newCbsOfRes
			}
		}
	}

	updateMux.Lock()
	breakers = newBreakers
	updateMux.Unlock()

	logging.Debug("[Outlier onRuleUpdate] Time statistics(ns) for updating circuit breaker rule", "timeCost", util.CurrentTimeNano()-start)
	circuitbreaker.LogRuleUpdate(breakerRules)
}
