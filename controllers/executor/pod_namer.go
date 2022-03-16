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

package executor

import (
	"math/rand"
	"strings"
	"bytes"
	corev1 "k8s.io/api/core/v1"
)

func NewNameForPod(pod *corev1.Pod) string {
	name := pod.Name
	name_parts := strings.Split(name, "-")
	N := len(name_parts)

	if N == 1 {
		name_parts = append(name_parts, genEndPart())
		N++
	} else {
		name_parts[N - 1] = genEndPart()
	}

	buf := bytes.Buffer{}
	for i := 0; i < N; i++ {
		buf.WriteString(name_parts[i])
		if i != N - 1 {
			buf.WriteString("-")
		}
	}
	return buf.String()
}

func genEndPart() string {
	part := make([]byte, 5)
	for i := 0; i < 5; i++ {
		part[i] = genRandomChar()
	}
	return string(part)
}

func genRandomChar() byte {
	c := rand.Intn(36)
	if c < 26 {
		return byte('a' + c)
	} else {
		c -= 26
		return byte('0' + c)
	}
}

