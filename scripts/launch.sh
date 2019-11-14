#!/bin/bash

cd ..
go build

counter=0
n=$1
debug=${2-111}
while [ $counter -lt $n ]
do
	let UIport=8080+$counter
	let Gport=5000+$counter
	name='Gossiper'$counter
	#nohup tilix -x bash -c 'echo "'$UIPort' and '$UIPort'"; exec bash' > /dev/null 2>&1 &
	nohup tilix -t $name -s $name --window-style=disable-csd-hide-toolbar -e bash -c './Peerster -name '$name' -GUIPort '$UIport' -UIPort '$UIport' -gossipAddr 127.0.0.1:'$Gport' -rtimer 2 -peers 127.0.0.1:5000 -debug '$debug'; exec bash' > /dev/null 2>&1 &
	((counter++))
done
# ./Peerster -GUIPort 8081 -UIPort 8081 -gossipAddr 127.0.0.1:5001 -name Issou
