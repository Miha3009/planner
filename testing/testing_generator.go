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

import "math/rand"

func GenCluster(n, m int) ([]ResourceInfo, []ResourceInfo) {
    rand.Seed(0)
    nodes := make([]ResourceInfo, n)
    for i := 0; i < n; i++ {
        nodes[i] = ResourceInfo{
            Cpu:    2000,
            Memory: 2200,
        }
    }

    pods := make([]ResourceInfo, m)
    for i := 0; i < m; i++ {
        pods[i] = ResourceInfo{
            Cpu:    randomRange(2, 6)*100,
            Memory: randomRange(2, 6)*100,
            //Cpu:    randomRange(400, 1000),
            //Memory: randomRange(400, 1000),
        }
    }

    return nodes, pods
}

func randomRange(a, b int) int {
    return rand.Intn(b-a) + a
}
