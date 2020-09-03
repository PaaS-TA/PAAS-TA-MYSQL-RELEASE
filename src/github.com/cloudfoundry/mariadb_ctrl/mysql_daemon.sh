#!/bin/bash

set -e

mode=$1


case "$mode" in
  'stop')
      echo "Stopping the cluster"
      USER=vcap /var/vcap/packages/mariadb/support-files/mysql.server stop --pid-file=/var/vcap/sys/run/mysql/mysql.pid --user=vcap > /dev/null 2>&1
      ;;

  'bootstrap')
      # Bootstrap the cluster, start the first node
      # that initiate the cluster
      echo "Bootstrapping the cluster"
      /var/vcap/packages/mariadb/bin/mysqld_safe --defaults-file=/var/vcap/jobs/mysql/config/my.cnf --wsrep-new-cluster &
      ;;

  'stand-alone')
      echo "Starting the node in stand-alone mode"
      /var/vcap/packages/mariadb/bin/mysqld_safe --defaults-file=/var/vcap/jobs/mysql/config/my.cnf --wsrep-on=OFF --wsrep-desync=ON --wsrep-OSU-method=RSU --wsrep-provider='none' --skip-networking &
      ;;

  'status')
      echo "Getting status of mysql process (exit 0 == running)"
      /var/vcap/packages/mariadb/support-files/mysql.server status --pid-file=/var/vcap/sys/run/mysql/mysql.pid
      ;;
esac

exit 0
