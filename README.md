# Планировщик

## Описание

Планировщик следит за выполнением заданных ему правил и в случае их невыполнения перераспределяет ресурсы или маштабирует систему. 

## Сборка и запуск

Основан на Operator SDK. Для сборки и запуска требуется (написаны фактически использованные версии):

* Go v1.17.7
* Docker 20.10.12
* kubectl v1.23.1
* operator-sdk v1.17.0 (если нужно менять api)
* Доступ к кластеру

Для сборки вызывается команда: 
<code>
make docker-build IMG=username/planner:tag
</code>

Теперь запушим докер образ на сервер: 
<code>
make docker-push IMG=username/planner:tag
</code>

Пропишем в dist/6-controller.yaml в пункте image значение username/planner:tag. 
Чтобы запустить планировщик нужно применить все .yaml файлы из папки dist:
<code>
kubectl apply -f dist
</code>

Осталось задать и применить конфигурацию для планировщика. Пример для конфигурации лежит в config/samples/apps_v1_planner.yaml. 
<code>
kubectl apply -f config/samples/apps_v1_planner.yaml
</code>

Проверить, что планировщик запустился можно так:
<code>
kubectl get pods -n planner-system
</code>
