#!/bin/bash
cd web

webpack -p

cp -r 'dist' '../server'

rm -rf '../server/public'

mv '../server/dist' '../server/public'
