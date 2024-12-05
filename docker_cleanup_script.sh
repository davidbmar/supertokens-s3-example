#!/bin/bash

# Check if the user is part of the docker group
if ! groups | grep -q "\bdocker\b"; then
    echo "You are not in the Docker group. Run 'sudo usermod -aG docker $USER' and re-login."
    exit 1
fi

echo "Stopping all Docker containers..."
docker stop $(docker ps -aq) 2>/dev/null || echo "No containers to stop."

echo "Removing all Docker containers..."
docker rm $(docker ps -aq) 2>/dev/null || echo "No containers to remove."

echo "Removing all Docker images..."
docker rmi $(docker images -q) -f 2>/dev/null || echo "No images to remove."

echo "Removing all Docker networks..."
docker network prune -f || echo "No networks to remove."

echo "Removing all Docker volumes..."
docker volume prune -f || echo "No volumes to remove."

echo "Cleaning up Docker system (dangling resources)..."
docker system prune -a -f

echo "Docker cleanup complete!"

