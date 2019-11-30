#!/bin/bash

cd ..
go build

rm -r _SharedFiles/* > /dev/null 2>&1
rm -r _Downloads/* > /dev/null 2>&1

cd _SharedFiles
echo 'hello world!' > hello.txt #> /dev/null 2>&1
wget https://media1.tenor.com/images/4c65228622e5b79fc6af8cba3c189fa9/tenor.gif > /dev/null 2>&1
mv tenor.gif kiddo.gif > /dev/null 2>&1
wget https://www.ietf.org/rfc/rfc1918.txt > /dev/null 2>&1

cd ..

counter=0
n=$1
debug=${2-111}
echo $debug
while [ $counter -lt $n ]
do
	let UIport=8080+$counter
	let Gport=5000+$counter
	name='Gossiper'$counter
	#nohup tilix -x bash -c 'echo "'$UIPort' and '$UIPort'"; exec bash' > /dev/null 2>&1 &
	nohup tilix -t $name -s $name --window-style=disable-csd-hide-toolbar -e bash -c './Peerster -name '$name' -GUIPort '$UIport' -UIPort '$UIport' -gossipAddr 127.0.0.1:'$Gport' -rtimer 2 -peers 127.0.0.1:5000 -hw3ex2=true -debug='$debug'; exec bash' > /dev/null 2>&1 &
	((counter++))
done
# ./Peerster -GUIPort 8081 -UIPort 8081 -gossipAddr 127.0.0.1:5001 -name Issou
