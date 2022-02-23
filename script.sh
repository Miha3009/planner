cd ~/projects/go-apps/demo-operator
make docker-build IMG=miha3009/demo-operator:latest
make docker-push IMG=miha3009/demo-operator:latest
minikube delete
minikube start
cd dist
kubectl apply -f .
cd ../config
kubectl apply -f 7-deploy.yaml
