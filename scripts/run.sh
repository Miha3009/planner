minikube delete
minikube start --nodes 1
kubectl apply -f ../dist
kubectl apply -f ../config/samples/apps_v1_planner.yaml
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/download/v0.4.2/components.yaml
minikube addons enable metrics-server
