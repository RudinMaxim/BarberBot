#!/bin/bash

REPO_PATH="$HOME/BarberBot"

cd $REPO_PATH

git pull origin main

docker-compose up --build -d
