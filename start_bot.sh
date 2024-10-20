#!/bin/bash

# Проверяем наличие обновлений в репозитории (если используется)
git pull

# Останавливаем предыдущую версию бота (если запущена)
docker-compose down

# Собираем новый образ
docker-compose build

# Запускаем бота
docker-compose up -d

# Выводим статус
docker-compose ps

echo "Бот успешно запущен!"