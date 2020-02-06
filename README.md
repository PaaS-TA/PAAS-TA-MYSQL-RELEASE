# openpaas-mysql-release

##1. Mysql Configuration
- paasta-mysql-broker	 :: 1 machine
- proxy :: 1 machine
- mysql_z1 :: 1 machine

##2. download
- $rm -rf ./src/*
- $cd ./src
- $wget -O download.zip http://45.248.73.44/index.php/s/nRiKxoQrWjXg9MS/download
- $unzip download.zip

##3. Deploy
>`$ cd $BOSH_RELEASE_DIR`<br>
>`$ bosh deployment paasta-mysql-vsphere-1.0.yml`<br>
>`$ bosh deploy`
