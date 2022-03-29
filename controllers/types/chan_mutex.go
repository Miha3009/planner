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

type ChanMutex struct {
    lockChan chan struct{}
}

func NewChanMutex() *ChanMutex {
    return &ChanMutex{
        lockChan: make(chan struct{}, 1),
    }
}

func (m *ChanMutex) Lock() {
    m.lockChan <- struct{}{}
}

func (m *ChanMutex) Unlock() {
    <-m.lockChan
}

func (m *ChanMutex) TryLock() bool {
    select {
    case m.lockChan <- struct{}{}:
        return true
    default:
        return false
    }
}
