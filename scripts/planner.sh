#!/bin/bash

if [[ $# -eq 0 ]]
then
  echo Not enough input argument
else
  if [[ $1 == "start" ]]
  then
    echo $(curl -X POST localhost:9999/start -s)
  elif [[ $1 == "stop" ]]
  then
    echo $(curl -X POST localhost:9999/stop -s)
  elif [[ $1 == "phase" ]]
  then
    echo Current phase is $(curl -X GET localhost:9999/phase -s)
  elif [[ $1 == "plan" ]]
  then
    #echo "str\nstr2"
    echo -e $(curl -X GET localhost:9999/planText -s)
  else
    echo Unknown argument \"${1}\"
  fi
fi
