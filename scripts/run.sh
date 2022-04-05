minikube delete
minikube start --nodes 1
kubectl apply -f dist/1-Namespace.yaml
kubectl apply -f dist/2-service_account.yaml
kubectl apply -f dist/3-role.yaml
kubectl apply -f dist/4-role_binding.yaml
kubectl apply -f dist/5-apps.hse.ru_planners.yaml
kubectl apply -f config/samples/apps_v1_planner.yaml
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/download/v0.4.2/components.yaml
minikube addons enable metrics-server
