#!/bin/bash

if [ "$#" -ne 2 ]; then
  exit 1
fi

if [ ! -e "$1" ]; then
  exit 1
fi

RIAK_ADMIN=$1
NODE_IP=$2

# run the admin command to get list of cluster members
# pick lines that list nodes (by presence of "riak@")
# grab the status and node columns
# pick the ones that have the specified IP and have valid status
# count the number of lines that match this
# trim whitespace from line count

$RIAK_ADMIN member-status | grep 'riak@*' | awk '{print $1 " " $4}' | grep \'riak@$NODE_IP\' | grep '^valid' | wc -l | awk '{$1=$1}1'