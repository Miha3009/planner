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
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"

    types "github.com/miha3009/planner/controllers/types"
    "github.com/prometheus/common/log"
)

type MoveMessage struct {
    Pod string
    OldNode string
    NewNode string
}

type PlanMessage struct {
    NodesChange int
    Moves []MoveMessage
}

func GetPlanMessage(plan *types.Plan) (PlanMessage, bool) {
    myPlan := PlanMessage{}
    if plan == nil {
        return myPlan, false
    }

    if len(plan.NodesToCreate) > 0 {
        myPlan.NodesChange = len(plan.NodesToCreate)
    } else if len(plan.NodesToDelete) > 0 {
        myPlan.NodesChange = -len(plan.NodesToCreate)
    } else {
        myPlan.NodesChange = 0
    }

    myPlan.Moves = make([]MoveMessage, len(plan.Movements))
    for i := range plan.Movements {
        myPlan.Moves[i] = MoveMessage{
            Pod: plan.Movements[i].Pod.Name,
            OldNode: plan.Movements[i].OldNode.Name,
            NewNode: plan.Movements[i].NewNode.Name,
        }
    }

    return myPlan, true
}

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
    
    http.HandleFunc("/plan", func(w http.ResponseWriter, r *http.Request) {
        if r.Method == "GET" {
            myPlan, ok := GetPlanMessage(reconciler.Cache.Plan)

            if ok {
                b, err := json.Marshal(myPlan)
                if err == nil {
                    fmt.Fprint(w, string(b))
                }
            }
        }
    })    

    http.HandleFunc("/planText", func(w http.ResponseWriter, r *http.Request) {
        if r.Method == "GET" {
            myPlan, ok := GetPlanMessage(reconciler.Cache.Plan)
            msg := ""

            if !ok {
                msg = "Plan not found.\n"
            } else if myPlan.NodesChange == 0 && len(myPlan.Moves) == 0 {
                msg = "Nothing will change.\n"
            } else {
                if myPlan.NodesChange == 0 {
                    msg = "Nodes will not be changed.\n"
                } else if myPlan.NodesChange > 0 {
                    msg = strconv.Itoa(myPlan.NodesChange) + " nodes will be created.\n"
                } else if myPlan.NodesChange < 0 {
                    msg = strconv.Itoa(-myPlan.NodesChange) + " nodes will be deleted.\n"
                }
                
                if len(myPlan.Moves) == 0 {
                    msg = msg + "Pod will not move\n"
                } else {
                    msg = msg + strconv.Itoa(len(myPlan.Moves)) + " pods will move.\n"
                    for _, move := range myPlan.Moves {
                        msg = msg + move.Pod + ": " + move.OldNode + " --> " + move.NewNode + ".\n"
                    }
                }
            }

            fmt.Fprint(w, msg)
        }
    })

    http.HandleFunc("/phase", func(w http.ResponseWriter, r *http.Request) {
        if r.Method == "GET" {
            fmt.Fprint(w, reconciler.Cache.Phase + "\n")
        }
    })

    log.Info("Starting server at port 9999")
    if err := http.ListenAndServe(":9999", nil); err != nil {
        log.Fatal(err)
    }
}
