minikube delete
minikube start
kubectl apply -f ../../dist/1-Namespace.yaml
kubectl apply -f ../../dist/2-service_account.yaml
kubectl apply -f ../../dist/3-role.yaml
kubectl apply -f ../../dist/4-role_binding.yaml
kubectl apply -f ../../dist/5-apps.hse.ru_planners.yaml
kubectl apply -f apps_v1_planner.yaml
kubectl apply -f sample1.yaml
kubectl apply -f sample2.yaml
sleep 5s
minikube node add
sleep 5s
kubectl describe nodes
go run ../../main.go
