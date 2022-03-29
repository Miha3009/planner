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

package testing

import (
    "math"
    "math/rand"
)

func Schedule(nodes []ResourceInfo, pods []ResourceInfo, strategy string) [][]ResourceInfo {
    rand.Seed(0)
    switch strategy {
    case "minFreeSpace":
        return scheduleForFreeSpace(nodes, pods, math.MaxInt, func(a, b int) bool { return a < b })
    case "maxFreeSpace":
        return scheduleForFreeSpace(nodes, pods, -1, func(a, b int) bool { return a > b })
    default:
        return nil
    }
}

func scheduleForFreeSpace(oldNodes []ResourceInfo, pods []ResourceInfo, init int, cmp func(a, b int) bool) [][]ResourceInfo {
    nodes := make([]ResourceInfo, len(oldNodes))
    copy(nodes, oldNodes)

    res := make([][]ResourceInfo, len(nodes))
    for i := range res {
        res[i] = make([]ResourceInfo, 0)
    }

    for i := range pods {
        maxFree := init
        maxIdx := make([]int, 0)
        for j := range nodes {
            if nodes[j].Cpu >= pods[i].Cpu && nodes[j].Memory >= pods[i].Memory {
                free := nodes[j].Cpu + nodes[j].Memory
                if cmp(free, maxFree) {
                    maxIdx = []int{j}
                    maxFree = free
                } else if free == maxFree {
                    maxIdx = append(maxIdx, j)
                }
            }
        }
        idx := maxIdx[rand.Intn(len(maxIdx))]
        nodes[idx].Cpu -= pods[i].Cpu
        nodes[idx].Memory -= pods[i].Memory
        res[idx] = append(res[idx], pods[i])
    }

    return res
}
