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

package base

import (
	"testing"
	
	"net/http"
	"io/ioutil"
)

func TestGood (t *testing.T) {
	resp, err := http.Get("https://192.168.49.2:8443/api/v1/pods")
	if err != nil {
		t.Log(err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	t.Log(string(body))
}


