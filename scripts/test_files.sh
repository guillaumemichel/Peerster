
cd ..
go build
cd client
go build
cd ..

# create the star topology with n peers
counter=0
n=8
while [ $counter -lt $n ]
do
	let UIport=8080+$counter
	let Gport=5000+$counter
	name='Gossiper'$counter
	#nohup tilix -x bash -c 'echo "'$UIPort' and '$UIPort'"; exec bash' > /dev/null 2>&1 &
	./Peerster -name $name -GUIPort $UIport -UIPort $UIport -gossipAddr 127.0.0.1:$Gport -peers 127.0.0.1:5000
	((counter++))
done

seq -s "a" 1000 | sed 's/[0-9]//g' > _SharedFiles/file1 
seq -s "b" 1000 | sed 's/[0-9]//g' >> _SharedFiles/file1 
seq -s "c" 1000 | sed 's/[0-9]//g' >> _SharedFiles/file1

./client/client -UIPort 8080 -file file1 # Gossiper0 indexes the file

while [ $counter -lt $n ]
do

done
