#!/bin/bash

# copy Peerster to Desktop in src/github.com/guillaumemichel/Peerster
cd ~/Desktop
mkdir src
cd src
mkdir github.com
cd github.com
mkdir guillaumemichel
cd guillaumemichel
cp -r /mnt/guillaume/Documents/workspace/go/src/github.com/guillaumemichel/Peerster/ .

# remove useless files
cd Peerster
rm -r debug
rm -r scripts
rm -r testing
rm -r .git
rm .gitignore
rm *.pdf
rm _SharedFiles/*
rm _Downloads/*
rm LICENSE
rm Peerster
rm client/client

# tar everything
cd ~/Desktop
tar -czvf guillaumemichel.tar.gz src
rm -r src

