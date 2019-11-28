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

# create the star topology with n peers
counter=0
n=4
while [ $counter -lt $n ]
do
	let UIport=8080+$counter
	let Gport=5000+$counter
	name='Gossiper'$counter
	./Peerster -name $name -GUIPort $UIport -UIPort $UIport -gossipAddr 127.0.0.1:$Gport -rtimer 1 -peers 127.0.0.1:5000 -debug 012 > scripts/logs/$name.out &
	((counter++))
done

sleep 2

hellohash="2225bf0bfcdf14b11bfaff5bf6ca258b543e546d06f4b944e7504b8f5990e298"
kiddohash="c7d5926bd51911b4ffa9ae00446e2bb72c28a9e33330665a2b9ecc55584d3aa6"
rfchash="71375f41d6a93154ad7b372f1cd397c3823831dfc65d3deeaec708a197c3a686"

cd client

counter=2
while [ $counter -lt $n ]
do
	let UIport=8080+$counter
    ./client -UIPort $UIport -file hello.txt
    ./client -UIPort $UIport -file rfc1918.txt
    ./client -UIPort $UIport -file kiddo.gif
	((counter++))
done

sleep 10

./client -UIPort 8081 -keywords hello,rfc -budget 4
./client -UIPort 8081 -file rfc1918.txt -request $rfchash
./client -UIPort 8081 -file hello.txt -request $hellohash

sleep 5

pkill -f Peerster 2> /dev/null
