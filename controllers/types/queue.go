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
    "time"
)

const minCapacity int = 16

type MetricsQueue interface {
    Push(m MetricsPackage)
    Pop() bool
    Size() int
    Get(idx int) MetricsPackage
    Shrink()
    Lock()
    Unlock()
    SetMaxAge(maxAge time.Duration)
}

type MetricsQueueImpl struct {
    M ChanMutex

    buf        []MetricsPackage
    waitingBuf []MetricsPackage
    start      int
    end        int
    capacity   int

    maxAge time.Duration
}

func NewMetricsQueue() *MetricsQueueImpl {
    return &MetricsQueueImpl{
        M:          *NewChanMutex(),
        buf:        make([]MetricsPackage, minCapacity),
        waitingBuf: make([]MetricsPackage, 0),
        start:      0,
        end:        0,
        capacity:   minCapacity,
        maxAge:     time.Second * 3600,
    }
}

func (q *MetricsQueueImpl) Push(m MetricsPackage) {
    if q.M.TryLock() {
        defer q.M.Unlock()
        if len(q.waitingBuf) > 0 {
            for i := range q.waitingBuf {
                q.push(q.waitingBuf[i])
            }
            q.waitingBuf = make([]MetricsPackage, 0)
        }
        q.push(m)
    } else {
        q.waitingBuf = append(q.waitingBuf, m)
    }
}

func (q *MetricsQueueImpl) push(m MetricsPackage) {
    if q.Size() == q.capacity-1 {
        q.resize()
    }

    q.buf[q.end] = m
    if q.end == q.capacity-1 {
        q.end = 0
    } else {
        q.end++
    }
}

func (q *MetricsQueueImpl) Pop() bool {
    if q.start == q.end {
        return false
    }

    if q.start == q.capacity-1 {
        q.start = 0
    } else {
        q.start++
    }

    return true
}

func (q *MetricsQueueImpl) Size() int {
    if q.end >= q.start {
        return q.end - q.start
    } else {
        return q.end - q.start + q.capacity
    }
}

func (q *MetricsQueueImpl) Get(idx int) MetricsPackage {
    if idx < 0 || idx > q.Size() {
        panic("IndexOutOfRangeError")
    }

    if q.start+idx >= q.capacity {
        return q.buf[idx-q.capacity+q.start]
    } else {
        return q.buf[idx+q.start]
    }
}

func (q *MetricsQueueImpl) Shrink() {
    if q.M.TryLock() {
        defer q.M.Unlock()
        for q.buf[q.start].Timestamp.Before(time.Now().Add(-q.maxAge)) {
            if !q.Pop() {
                break
            }
        }
    }
}

func (q *MetricsQueueImpl) Lock() {
    q.M.Lock()
}

func (q *MetricsQueueImpl) Unlock() {
    q.M.Unlock()
}

func (q *MetricsQueueImpl) SetMaxAge(maxAge time.Duration) {
    q.maxAge = maxAge
}

func (q *MetricsQueueImpl) resize() {
    newBuf := make([]MetricsPackage, q.capacity<<1)
    size := q.Size()

    if q.end >= q.start {
        copy(newBuf, q.buf[q.start:q.end])
    } else {
        n := copy(newBuf, q.buf[q.start:])
        copy(newBuf[n:], q.buf[:q.end+1])
    }

    q.start = 0
    q.end = size
    q.buf = newBuf
    q.capacity = q.capacity << 1
}
