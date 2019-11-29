#!/bin/bash

rm -r logs 2> /dev/null
mkdir logs 2> /dev/null

cd ..
go build
cd client
go build
cd ..

rm -r _SharedFiles/* > /dev/null 2>&1
rm -r _Downloads/* > /dev/null 2>&1
pkill -f Peerster 2> /dev/null

cd _SharedFiles
echo 'hello world!' > hello.txt # > /dev/null 2>&1
wget https://media1.tenor.com/images/4c65228622e5b79fc6af8cba3c189fa9/tenor.gif > /dev/null 2>&1
mv tenor.gif kiddo.gif > /dev/null 2>&1
wget https://www.ietf.org/rfc/rfc1918.txt > /dev/null 2>&1

cd ..

./Peerster -name Gossiper0 -GUIPort 8080 -UIPort 8080 -gossipAddr 127.0.0.1:5000 -rtimer 0 -peers 127.0.0.1:5001 -debug 222 > scripts/logs/Gossiper0.out &
./Peerster -name Gossiper1 -GUIPort 8081 -UIPort 8081 -gossipAddr 127.0.0.1:5001 -rtimer 0 -peers 127.0.0.1:5000 -debug 222 > scripts/logs/Gossiper1.out &
./Peerster -name Gossiper2 -GUIPort 8082 -UIPort 8082 -gossipAddr 127.0.0.1:5002 -rtimer 0 -peers 127.0.0.1:5000 -debug 222 > scripts/logs/Gossiper2.out &
