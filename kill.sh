#!/bin/bash
# func getGoBuild : ps a | grep go | then grep "go-build" or "worker/main" | awk '{print $1}'
# func killGoBuild : kill -9 <pid>
# then kill -9 <pid>
#  read .PID file and kill the process if not empty
# function getGoBuild() {
#     ps a | grep go | grep "go-build" | grep -v grep | awk '{print $1}'
# }

# function getGoMain() {
#     ps a | grep go | grep "worker/main" | grep -v grep | awk '{print $1}'
# }
# kill -9 $(getGoBuild)
# kill -9 $(getGoMain)

pidFile=".PID"
pid=$(cat $pidFile)

if [ -z "$pid" ]; then
    echo "No PID found in $pidFile"
    exit 1
fi

kill -9 $pid