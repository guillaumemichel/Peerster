#!/bin/bash

cd ..
go build
rm -r _SharedFiles/*
rm -r _Downloads/*
cd _SharedFiles
echo 12345678 > gui_test.txt
cd ..

nohup tilix -t A -s A --window-style=disable-csd-hide-toolbar -e bash -c './Peerster -name A -GUIPort 8080 -UIPort 8080 -gossipAddr 127.0.0.1:5000 -rtimer 1 -debug 010; exec bash' > /dev/null 2>&1 &
nohup tilix -t B -s B --window-style=disable-csd-hide-toolbar -e bash -c './Peerster -name B -GUIPort 8081 -UIPort 8081 -gossipAddr 127.0.0.1:5001 -rtimer 1 -peers 127.0.0.1:5000 -debug 010; exec bash' > /dev/null 2>&1 &
nohup tilix -t C -s C --window-style=disable-csd-hide-toolbar -e bash -c './Peerster -name C -GUIPort 8082 -UIPort 8082 -gossipAddr 127.0.0.1:5002 -rtimer 1 -peers 127.0.0.1:5001 -debug 010; exec bash' > /dev/null 2>&1 &
nohup tilix -t D -s D --window-style=disable-csd-hide-toolbar -e bash -c './Peerster -name D -GUIPort 8083 -UIPort 8083 -gossipAddr 127.0.0.1:5003 -rtimer 1 -peers 127.0.0.1:5002 -debug 010; exec bash' > /dev/null 2>&1 &

echo 9bf97d5561afe7e855eb2ec7969de172b0769148fda89f5543834cc4a1aaafe1

