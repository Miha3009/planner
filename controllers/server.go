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

package controllers

import (
    "fmt"
    "net/http"

    types "github.com/miha3009/planner/controllers/types"
    "github.com/prometheus/common/log"
)

func RunServer(reconciler *PlannerReconciler) {
    http.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
        if r.Method == "POST" {
            reconciler.Events <- types.Start
            fmt.Fprint(w, "Planner started\n")
        }
    })

    http.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
        if r.Method == "POST" {
            reconciler.Events <- types.Stop
            fmt.Fprint(w, "Planner stopped\n")
        }
    })

    log.Info("Starting server at port 9999")
    if err := http.ListenAndServe(":9999", nil); err != nil {
        log.Fatal(err)
    }
}
