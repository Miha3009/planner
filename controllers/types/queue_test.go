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

package types

import (
    "testing"
    "time"

    types "github.com/miha3009/planner/controllers/types"
)

func TestSimple(t *testing.T) {
    queue := types.NewMetricsQueue(time.Second)
    for i := int64(0); i < 20; i++ {
        queue.Push(genPackage(i))
    }

    if getIndex(queue.Get(2)) != 2 {
        t.FailNow()
    }

    queue.Pop()

    if getIndex(queue.Get(2)) != 3 || queue.Size() != 19 {
        t.FailNow()
    }
}

func TestConcurrent(t *testing.T) {
    queue := types.NewMetricsQueue(time.Second)
    for i := int64(0); i < 10; i++ {
        queue.Push(genPackage(i))
    }
    queue.Lock()
    go addToQueue(queue)
    time.Sleep(time.Second / 10)
    if queue.Size() != 10 {
        t.FailNow()
    }
    queue.Unlock()
    time.Sleep(time.Second / 10)
    queue.Push(genPackage(int64(20)))
    if queue.Size() != 21 {
        t.FailNow()
    }
}

func addToQueue(queue types.MetricsQueue) {
    for i := int64(10); i < 20; i++ {
        queue.Push(genPackage(i))
    }
}

func genPackage(x int64) types.MetricsPackage {
    return types.MetricsPackage{Timestamp: time.UnixMilli(x)}
}

func getIndex(p types.MetricsPackage) int64 {
    return p.Timestamp.UnixMilli()
}
