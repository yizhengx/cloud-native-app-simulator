# istioctl install -f istio-operator.yaml -y
kubectl delete ns hydragen
kubectl create ns hydragen
kubectl create configmap rate-limit-filter --from-file=rate-limit-filter.wasm="cpp.wasm" -n hydragen
kubectl label namespace hydragen istio-injection=enabled
for d in yamls/service*; do
    echo "-> Apply service yaml file ${d}"
    kubectl apply -f $d -n hydragen
done