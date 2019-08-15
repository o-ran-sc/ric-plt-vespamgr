# RIC VES Agent configuration

Configuration files and instructions for deploying VES Agent in RIC

# Unit Tests
```
go test ./... -v
```

# Building inside Docker with hard coded configuration
* Address of running Prometheus
* Address of VES Collector
* metric names

```
sudo docker build -t vesmgr .
sudo docker tag vesmgr <helm-repo>:5000/vesmgr
sudo docker push <helm-repo>:5000/vesmgr
```

# Deploy with kubernetes

```
kubectl run vesmgr --image=<helm-repo>:5000/vesmgr --env="VESMGR_HB_INTERVAL=60s" --env="VESMGR_MEAS_INTERVAL=30s" --env="VESMGR_PRICOLLECTOR_ADDR=10.144.54.115" --env="VESMGR_PRICOLLECTOR_PORT=8443" --env="VESMGR_PROMETHEUS_ADDR=http://10.244.0.62:9090"
```

# Deploy with helm chart with hard coded configuration

```
helm install ves-agent-chart/ --name vesmgr
```

## License

See [LICENSES.txt](LICENSES.txt) file.
