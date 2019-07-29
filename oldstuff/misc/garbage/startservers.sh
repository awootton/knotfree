#!/bin/bash -ex

# The currect directory should be src/knotfreeiot/deploy

kubectl create namespace servers | true

kubectl config set-context --current --namespace=servers

export N=server

export CPU=20m
export MEM=64Mi

export CPU=400m
export MEM=2048Mi

export REPLICAS=1

 ./template.sh server.yaml | kubectl apply -f -

POD=""
while [ "$POD" == "" ]
do
    POD=$(kubectl get pods -o name | grep -m1 knotfree${N} | cut -d'/' -f 2) 
done

#kubectl exec ${POD} -- bash -c "go get -u github.com/eclipse/paho.mqtt.golang"

#  log on to it:  kubectl exec -it ${POD} -- bash 

kubectl exec ${POD} -- bash -c "pkill main" | true

# copy the latest version up up the pod
echo "Copy source..."
kubectl cp ../../knotfree ${POD}:/go/src/

# start the process
echo "start"
kubectl exec ${POD} -- bash -c "cd src/knotfree && go run main.go server"

echo "stopping main.go"
kubectl exec ${POD} -- bash -c "pkill main" | true

echo "finished"
