#!/bin/bash

docker build . --tag zanattabruno/ric-plt-vespamgr && docker push zanattabruno/ric-plt-vespamgr

kubectl rollout restart deployment deployment-ricplt-vespamgr -n ricplt