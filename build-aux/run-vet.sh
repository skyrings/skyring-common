#!/bin/bash

for ENTRY in `ls . | grep -v '^vendor'`
do
	if [ -d ${ENTRY} ]
	then
		${PWD}/build-aux/vet $ENTRY
	fi
done
