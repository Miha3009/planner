/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package algorithm

import (
    "context"
    "math"
    "math/rand"
    "sort"
    //"time" // for testing
    //"strconv" // for testing
    
    "github.com/prometheus/common/log"
    
    "github.com/miha3009/planner/controllers/rescheduler/algorithm/glpk"
    types "github.com/miha3009/planner/controllers/types"
    helper "github.com/miha3009/planner/controllers/helper"
)

type Optimizer struct {
    TimeLimit int
    MaxNodesPerCycle int
    MaxFailAttemps int
    
    nodes []types.NodeInfo 
    pods []types.PodInfo
    lp *glpk.Prob
    ind []int32
    rowCount int
}

func NewOptimizer(timeLimit, maxNodesPerCycle int) *Optimizer {
    return &Optimizer{TimeLimit: timeLimit, MaxNodesPerCycle: maxNodesPerCycle, MaxFailAttemps: 3}
}

func (o *Optimizer) Optimize(ctx context.Context, oldNodes []types.NodeInfo) []types.NodeInfo {
    nodes := helper.DeepCopyNodes(oldNodes)
    failAttemps := 0

    //times := make([]int64, 0) // for testing
    for {
        if len(nodes) <= 1 || helper.ContextEnded(ctx) {
            break
        }
    
    	sort.Slice(nodes, func(i, j int) bool { return o.freeSpace(&nodes[i]) < o.freeSpace(&nodes[j]) })
        nodeI := o.findFirstNonEmptyNode(nodes)
        if nodeI == -1 {
            break
        }
        nodes = nodes[:nodeI+1]

        L := nodeI - o.MaxNodesPerCycle
        R := nodeI
        if L < 0 {
            L = 0
        }
        podsCount := len(nodes[nodeI].Pods)
        pod := nodes[nodeI].Pods[rand.Intn(podsCount)]
        
        //start := time.Now().UnixMilli() // for testing
        if o.lpSolve(nodes[L:R], []types.PodInfo{pod}) {
            nodes[nodeI].RemovePod(pod)
            failAttemps = 0
        } else {
            failAttemps++
            if failAttemps >= o.MaxFailAttemps {
                break
            }
        }
        //times = append(times, time.Now().UnixMilli() - start) // for testing
    }
    
    //o.addTimes(nodes, times) // for testing
    
    return o.sortNodesBack(nodes, oldNodes)
}

func (o *Optimizer) lpSolve(nodes []types.NodeInfo, pods []types.PodInfo) bool {
    for i := range nodes {
    	pods = append(pods, nodes[i].Pods...)
    }

    o.lp = glpk.New()
    o.lp.SetObjDir(glpk.MAX)
    o.nodes = nodes
    o.pods = pods
    
    N := len(nodes)
    M := len(pods)

    o.lp.AddCols(N*M)
    for i := 1; i <= N*M; i++ {
        o.lp.SetObjCoef(i, 1.0)
        o.lp.SetColKind(i, glpk.BV)
    }

    o.ind = make([]int32, N*M+1)
    for i := range o.ind {
    	o.ind[i] = int32(i)
    }
    o.rowCount = 0

    o.addOnePodConstraint()
    o.addNodeCpuConstraint()
    o.addNodeMemoryConstraint()

    iocp := glpk.NewIocp()
    iocp.SetPresolve(true)
    iocp.SetMsgLev(glpk.MSG_OFF)
    iocp.SetTmLim(o.TimeLimit)

    if err := o.lp.Intopt(iocp); err != nil {
        if err != glpk.ETMLIM {
            log.Warn("Mip error: %v", err)
        }
        return false
    }
    
    if o.lp.MipObjVal() < float64(len(pods)) {
        return false
    }
    
    for i := 0; i < N; i++ {
        for len(nodes[i].Pods) > 0 {
            nodes[i].RemovePod(nodes[i].Pods[0])
        }
        nodes[i].PodsCpu = 0
        nodes[i].PodsMemory = 0
    }
    
    for i := 0; i < N*M; i++ {
    	if o.lp.MipColVal(i+1) > 0.0 {
	    nodes[i / M].AddPod(pods[i % M])
    	}
    }

    return true
}

func (o *Optimizer) findFirstNonEmptyNode(nodes []types.NodeInfo) int {
    for i := len(nodes) - 1; i >= 0; i-- {
        if len(nodes[i].Pods) != 0 {
            return i
        }
    }
    return -1
}

func (o *Optimizer) addOnePodConstraint() {
    N := len(o.nodes)
    M := len(o.pods)

    o.lp.AddRows(M)
    for k := 0; k < M; k++ {
    	val := make([]float64, N*M+1)
    	for i := 0; i < N; i++ {
    	    val[i*M+k+1] = 1.0
    	}
    	o.lp.SetRowBnds(o.rowCount + k + 1, glpk.DB, 0.0, 1.0)
    	o.lp.SetMatRow(o.rowCount + k + 1, o.ind, val)
    }
    o.rowCount += M
}

func (o *Optimizer) addNodeCpuConstraint() {
    N := len(o.nodes)
    M := len(o.pods)

    o.lp.AddRows(N)
    for k := 0; k < N; k++ {
    	val := make([]float64, N*M+1)
    	for i := 0; i < M; i++ {
    	    val[k*M+i+1] = float64(o.pods[i].Cpu)
    	}
    	o.lp.SetRowBnds(o.rowCount + k + 1, glpk.DB, 0.0, float64(o.nodes[k].MaxCpu))
    	o.lp.SetMatRow(o.rowCount + k + 1, o.ind, val)
    }
    o.rowCount += N
}

func (o *Optimizer) addNodeMemoryConstraint() {
    N := len(o.nodes)
    M := len(o.pods)

    o.lp.AddRows(N)
    for k := 0; k < N; k++ {
    	val := make([]float64, N*M+1)
    	for i := 0; i < M; i++ {
    	    val[k*M+i+1] = float64(o.pods[i].Memory)
    	}
    	o.lp.SetRowBnds(o.rowCount + k + 1, glpk.DB, 0.0, float64(o.nodes[k].MaxMemory))
    	o.lp.SetMatRow(o.rowCount + k + 1, o.ind, val)
    }
    o.rowCount += N
}

func (o *Optimizer) freeSpace(node *types.NodeInfo) float64 {
    return math.Min(
        float64(node.MaxCpu - node.PodsCpu) / float64(node.MaxCpu),
        float64(node.MaxMemory - node.PodsMemory) / float64(node.MaxMemory))
}

func (o *Optimizer) sortNodesBack(nodes []types.NodeInfo, oldNodes []types.NodeInfo) []types.NodeInfo {
    nodesByName := make(map[string]int)
    for i := range oldNodes {
        nodesByName[oldNodes[i].Name] = i
    }
    sort.Slice(nodes, func(i, j int) bool { return nodesByName[nodes[i].Name] < nodesByName[nodes[j].Name] })
    return nodes
}

/*func (o *Optimizer) addTimes(nodes []types.NodeInfo, times []int64) {
    timeStrings := make([]string, len(times))
    for i := range times {
        timeStrings[i] = strconv.Itoa(int(times[i]))
    }

    for i := range nodes {
        nodes[i].UntoleratedPods = timeStrings
    }
}*/ // for testing

