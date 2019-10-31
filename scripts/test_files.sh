
rm -r logs 2> /dev/null
mkdir logs 2> /dev/null
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
	./Peerster -name $name -GUIPort $UIport -UIPort $UIport -gossipAddr 127.0.0.1:$Gport -rtimer 1 -peers 127.0.0.1:5000 > scripts/logs/$name.out & 
	((counter++))
done

seq -s "a" 8192 | sed 's/[0-9]//g' > _SharedFiles/file1 
seq -s "b" 8192 | sed 's/[0-9]//g' >> _SharedFiles/file1 
seq -s "c" 8192 | sed 's/[0-9]//g' >> _SharedFiles/file1

echo "hello world" > _SharedFiles/file2

./client/client -UIPort 8080 -file file1 # Gossiper0 indexes the file
./client/client -UIPort 8080 -file file2 # Gossiper2 indexes the file

hash1="292572baeeef9056338c705811c7ae95b1798f212f66eb6f0b855aa25c4244be"
hash2="f83e4b6bba3efac41f1ff56ee97adf7454680fee778924cb5ba06311d136ad1c"

passed="T"

sleep 12

counter=1
while [ $counter -lt $n ]
do
	let UIport=8080+$counter
	name='Gossiper'$counter
	let id=$counter-1
	./client/client -UIPort $UIport -file file1_$counter -dest Gossiper0 -request $hash1
	./client/client -UIPort $UIport -file file2_$counter -dest Gossiper$id -request $hash2
	((counter++))
done

counter=1
while [ $counter -lt $n ]
do
	name='Gossiper'$counter
	if [ -f "_Downloads/file1_$counter" ]; then
		cmp -s _SharedFiles/file1 _Downloads/file1_$counter
		if [ $? -eq 1 ]; then
			echo "FAILED file1 for "$name
			passed="F"
		fi
	else
		echo "FAILED file1 for "$name
		passed="F"
	fi

	if [ -f "_Downloads/file2_$counter" ]; then
		cmp -s _SharedFiles/file2 _Downloads/file2_$counter
		if [ $? -eq 1 ]; then
			echo "FAILED file2 for "$name
			passed="F"
		fi
	else
		echo "FAILED file2 for "$name
		passed="F"
	fi
	((counter++))
done


#sleep 20

pkill -f Peerster 2> /dev/null
rm _Downloads/file* 2> /dev/null
rm _SharedFiles/file* 2> /dev/null

if [[ "$passed" == "T" ]]; then
	echo "=D PASSED =D"
else
	echo "=O FAILED =O"
fi