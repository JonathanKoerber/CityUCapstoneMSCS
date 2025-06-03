#### Test Modbus Server

build image 

`sudo docker build -f ./honeypot-core/app/plc-node/Dockerfile-Modbus-TCP -t modbus-node:latest ./honeypot-core/app/plc-node`

run images 

`sudo docker run -d --name pumpTest --network honeynet --ip 172.18.0.15 modbus-node:latest`

Test ports

`mbpoll -m tcp -a 1 -r 0 -p 1502 -c 10 172.18.0.15`

`sudo docker build -t modbus-server -f Dockfile-Modbus-TCP .`