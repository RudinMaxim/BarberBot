#!/bin/bash

# Путь к вашему репозиторию
REPO_PATH="$HOME/BarberBot"

# Перейти в каталог репозитория
cd $REPO_PATH

# Обновить репозиторий
git pull origin main  # Замените 'main' на нужную ветку, если это необходимо

# Пересобрать и запустить контейнеры
docker-compose up --build -d
