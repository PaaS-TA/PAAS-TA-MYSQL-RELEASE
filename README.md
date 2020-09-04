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
    $ wget -O src.zip http://45.248.73.44/index.php/s/LbZXfZJdfCtepiM/download   

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
      │   └── check-0.9.13.tar.gz
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
      │   └── galera-25.3.23.tar.gz
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
      │   └── mariadb-10.1.38.tar.gz
      ├── mariadb-patch
      │   └── add_sst_interrupt.patch
      ├── mysqlclient
      │   └── mariadb-connector-c-2.1.0-src.tar.gz
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
      │   └── scons-2.3.1.tar.gz
      └── xtrabackup
          ├── autoconf-2.65.tar.gz
          ├── automake-1.14.1.tar.gz
          ├── libaio_0.3.110.orig.tar.gz
          ├── libev-4.22.tar.gz
          ├── libtool-2.4.2.tar.gz
          ├── percona-xtrabackup-2.4.8.tar.gz
          └── socat-1.7.3.2.tar.gz
      
    ```   
  - Create PaaS-TA Mysql Release   
    ```   
    ## <VERSION> :: release version (e.g. 2.0.1)   
    ## <RELEASE_TARBALL_PATH> :: release file path (e.g. /home/ubuntu/workspace/paasta-mysql-<VERSION>.tgz)   
    $ bosh -e <bosh_name> create-release --name=paasta-mysql --version=<VERSION> --tarball=<RELEASE_TARBALL_PATH> --force   
    ```   
### Deployment   
- https://github.com/PaaS-TA/service-deployment   
