minikube delete
minikube start
kubectl apply -f ../../dist
kubectl apply -f apps_v1_planner.yaml
kubectl apply -f sample1.yaml
kubectl apply -f sample2.yaml
sleep 5s
minikube node add
sleep 5s
kubectl describe nodes
go run ../../main.go
