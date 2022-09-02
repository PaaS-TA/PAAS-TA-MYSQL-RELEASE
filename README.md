## Related Repositories

<table>
  <tr>
    <td colspan=2 align=center>플랫폼</td>
    <td colspan=2 align=center><a href="https://github.com/PaaS-TA/paasta-deployment">어플리케이션 플랫폼</a></td>
    <td colspan=2 align=center><a href="https://github.com/PaaS-TA/paas-ta-container-platform">컨테이너 플랫폼</a></td>
  </tr>
  <tr>
    <td colspan=2 rowspan=2 align=center>포털</td>
    <td colspan=2 align=center><a href="https://github.com/PaaS-TA/portal-deployment">AP 포털</a></td>
    <td colspan=2 align=center><a href="https://github.com/PaaS-TA/container-platform-portal-release">CP 포털</a></td>
  </tr>
  <tr align=center>
    <td colspan=4><a href="https://github.com/PaaS-TA/PaaS-TA-Monitoring">모니터링 대시보드</a></td>
  </tr>
  <tr align=center>
    <td rowspan=2 colspan=2><a href="https://github.com/PaaS-TA/monitoring-deployment">모니터링</a></td>
    <td><a href="https://github.com/PaaS-TA/PaaS-TA-Monitoring-Release">Monitoring</a></td>
    <td><a href="https://github.com/PaaS-TA/paas-ta-monitoring-logsearch-release">Logsearch</a></td>
    <td><a href="https://github.com/PaaS-TA/paas-ta-monitoring-influxdb-release">InfluxDB</a></td>
    <td><a href="https://github.com/PaaS-TA/paas-ta-monitoring-redis-release">Redis</a></td>
  </tr>
  <tr align=center>
    <td><a href="https://github.com/PaaS-TA/PAAS-TA-PINPOINT-MONITORING-RELEASE">Pinpoint</td>
    <td><a href="https://github.com/PaaS-TA/PAAS-TA-PINPOINT-MONITORING-BUILDPACK">Pinpoint Buildpack</td>
    <td></td>
    <td></td>
  </tr>
  </tr>
  <tr align=center>
    <td rowspan=4 colspan=2><a href="https://github.com/PaaS-TA/service-deployment">AP 서비스</a></td>
    <td><a href="https://github.com/PaaS-TA/PAAS-TA-CUBRID-RELEASE">Cubrid</a></td>
    <td><a href="https://github.com/PaaS-TA/PAAS-TA-API-GATEWAY-SERVICE-RELEASE">Gateway</a></td>
    <td><a href="https://github.com/PaaS-TA/PAAS-TA-GLUSTERFS-RELEASE">GlusterFS</a></td>
    <td><a href="https://github.com/PaaS-TA/PAAS-TA-APP-LIFECYCLE-SERVICE-RELEASE">Lifecycle</a></td>
  </tr>
  <tr align=center>
    <td><a href="https://github.com/PaaS-TA/PAAS-TA-LOGGING-SERVICE-RELEASE">Logging</a></td>
    <td><a href="https://github.com/PaaS-TA/PAAS-TA-MONGODB-SHARD-RELEASE">MongoDB</a></td>
    <td><a href="https://github.com/PaaS-TA/PAAS-TA-MYSQL-RELEASE">🚩 MySQL</a></td>
    <td><a href="https://github.com/PaaS-TA/PAAS-TA-PINPOINT-RELEASE">Pinpoint APM</a></td>
  </tr>
  <tr align=center>
    <td><a href="https://github.com/PaaS-TA/PAAS-TA-DELIVERY-PIPELINE-RELEASE">Pipeline</a></td>
    <td align=center><a href="https://github.com/PaaS-TA/rabbitmq-release">RabbitMQ</a></td>
    <td><a href="https://github.com/PaaS-TA/PAAS-TA-ON-DEMAND-REDIS-RELEASE">Redis</a></td>
    <td><a href="https://github.com/PaaS-TA/PAAS-TA-SOURCE-CONTROL-RELEASE">Source Control</a></td>
  </tr>
  <tr align=center>
    <td><a href="https://github.com/PaaS-TA/PAAS-TA-WEB-IDE-RELEASE-NEW">WEB-IDE</a></td>
    <td></td>
    <td></td>
    <td></td>
  </tr>
  <tr align=center>
    <td rowspan=1 colspan=2><a href="https://github.com/PaaS-TA/paas-ta-container-platform-deployment">CP 서비스</a></td>
    <td><a href="https://github.com/PaaS-TA/container-platform-pipeline-release">Pipeline</a></td>
    <td><a href="https://github.com/PaaS-TA/container-platform-source-control-release">Source Control</a></td>
    <td></td>
    <td></td>
  </tr>
</table>
<i>🚩 You are here.</i>



  

  


## PAAS-TA-MYSQL-RELEASE   

### Notices   
  - Use PAAS-TA-MYSQL-RELEASE >= v.2.0.1   
    - PaaS-TA >= v.5.0.2   
    - service-deployment >= v5.0.2   
  - Use PAAS-TA-MYSQL-RELEASE =< v.2.0.0   
    - PaaS-TA =< v.5.0.1   
    - service-deployment =< v5.0.1   

### PaaS-TA Mysql Release Configuration    
  - mysql : N machine(s)   
  - mysql-broker : 1 machine   
  - proxy : 1 machine   
  - arbitrator : 1 machine   

### Create PaaS-TA Mysql Release   
  - Download the latest PaaS-TA Mysql Release    
    ```   
    $ git clone https://github.com/PaaS-TA/PAAS-TA-MYSQL-RELEASE.git
    $ cd PAAS-TA-MYSQL-RELEASE   
    ```   
  - Download & Copy "source files" into the src directory   
    ```   
    ## download source files   
    $ wget -O src.zip https://nextcloud.paas-ta.org/index.php/s/6DqdDkmbk4qqimF/download

    ## unzip download source files   
    $ unzip src.zip   

    ## final src directory (2-depth)  
    src
      ├── boost
      │   └── boost_1_59_0.tar.gz
      ├── cf-mysql-common
      │   ├── logging.sh
      │   └── pid_utils.sh
      ├── check
      │   └── check_0.10.0.orig.tar.gz
      ├── cipher_finder
      │   └── cipher_finder.jar
      ├── cli
      │   └── cf-cli_6.36.1_linux_x86-64.tgz
      ├── cluster_schema_verifier
      │   ├── Gemfile
      │   ├── Gemfile.lock
      │   ├── lib
      │   └── spec
      ├── galera
      │   └── galera-25.3.34.tar.gz
      ├── generate-auto-tune-mysql
      │   ├── auto_tune_generator.go
      │   ├── auto_tune_generator_test.go
      │   ├── generate_auto_tune_mysql_suite_test.go
      │   ├── main.go
      │   └── vendor
      ├── github.com
      │   ├── cloudfoundry
      │   ├── cloudfoundry-incubator
      │   └── onsi
      ├── golang-1.11-linux
      │   ├── compile.env.generic
      │   ├── compile.env.linux
      │   ├── go1.11.1.linux-amd64.tar.gz
      │   └── runtime.env.linux
      ├── gra-log-purger
      │   ├── Gopkg.lock
      │   ├── Gopkg.toml
      │   ├── README.md
      │   ├── gra_log_purger.go
      │   ├── gra_log_purger_suite_test.go
      │   ├── gra_log_purger_test.go
      │   └── vendor
      ├── mariadb
      │   └── mariadb-10.2.44-linux-x86_64.tar.gz
      ├── mariadb-patch
      │   └── add_sst_interrupt.patch
      ├── mysqlclient
      │   └── mariadb-connector-c-3.1.12-src.tar.gz
      ├── op-mysql-java-broker
      │   └── openpaas-service-java-broker-mysql.jar
      ├── openjdk
      │   └── openjdk-1.8.0_45.tar.gz
      ├── pcre-8.35.tar.gz
      ├── python
      │   └── Python-2.7.13.tgz
      ├── quota-enforcer
      │   ├── LICENSE
      │   ├── README.md
      │   ├── bin
      │   ├── clock
      │   ├── config
      │   ├── config-default.yaml
      │   ├── config-example.yaml
      │   ├── database
      │   ├── enforcer
      │   ├── integration
      │   ├── main.go
      │   └── vendor
      ├── ruby
      │   ├── ruby-2.3.8.tar.gz
      │   ├── rubygems-2.7.6.tgz
      │   └── yaml-0.1.7.tar.gz
      ├── scons
      │   └── scons-3.0.5.tar.gz
      └── xtrabackup
          ├── autoconf-2.69.tar.gz
          ├── automake-1.15.1.tar.gz
          ├── libaio_0.3.110.orig.tar.gz
          ├── libev-4.33.tar.gz
          ├── libtool-2.4.6.tar.gz
          ├── percona-xtrabackup-2.4.24.tar.gz
          └── socat-1.7.4.1.tar.gz
      
    ```   
  - Create PaaS-TA Mysql Release   
    ```   
    ## <VERSION> :: release version (e.g. 2.1.2)   
    ## <RELEASE_TARBALL_PATH> :: release file path (e.g. /home/ubuntu/workspace/paasta-mysql-<VERSION>.tgz)   
    $ bosh -e <bosh_name> create-release --name=paasta-mysql --version=<VERSION> --tarball=<RELEASE_TARBALL_PATH> --force   
    ```   
### Deployment   
- https://github.com/PaaS-TA/service-deployment   

## Contributors ✨

<a href="https://github.com/PaaS-TA/PAAS-TA-MYSQL-RELEASE/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=PaaS-TA/PAAS-TA-MYSQL-RELEASE" />
</a>
