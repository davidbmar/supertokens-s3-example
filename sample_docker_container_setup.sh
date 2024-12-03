#!/bin/bash
#
#
#
echo "sudo usermod -aG docker $USER" 
echo "Before running this script be sure to do the above!"
read -p "Press any key to resume ..."

#!/bin/bash
set -e  # Exit on error

echo "Setting up SuperTokens and PostgreSQL containers..."

# Stop and remove existing containers
echo "Stopping and removing existing containers..."
docker stop supertokens-core supertokens-postgres 2>/dev/null || true
docker rm supertokens-core supertokens-postgres 2>/dev/null || true

# Create network if it doesn't exist
if ! docker network inspect supertokens >/dev/null 2>&1; then
    echo "Creating supertokens network..."
    docker network create supertokens
fi

# Start PostgreSQL with healthcheck
echo "Starting PostgreSQL container..."
docker run -d \
    --name supertokens-postgres \
    --network supertokens \
    --health-cmd="pg_isready -U supertokens" \
    --health-interval=5s \
    --health-timeout=3s \
    --health-retries=5 \
    -e POSTGRES_USER=supertokens \
    -e POSTGRES_PASSWORD=your-strong-password \
    -e POSTGRES_DB=supertokens \
    postgres:latest

# Wait for PostgreSQL to be healthy
echo "Waiting for PostgreSQL to be healthy..."
WAIT_SECONDS=0
MAX_WAIT=30

while [ $WAIT_SECONDS -lt $MAX_WAIT ]; do
    if [ "$(docker inspect --format='{{.State.Health.Status}}' supertokens-postgres)" == "healthy" ]; then
        echo "PostgreSQL is ready!"
        break
    fi
    echo "Waiting for PostgreSQL to be ready... ($WAIT_SECONDS seconds)"
    sleep 5
    WAIT_SECONDS=$((WAIT_SECONDS + 5))
done

if [ $WAIT_SECONDS -ge $MAX_WAIT ]; then
    echo "Error: PostgreSQL failed to become ready within $MAX_WAIT seconds"
    echo "PostgreSQL logs:"
    docker logs supertokens-postgres
    exit 1
fi

# Start SuperTokens
echo "Starting SuperTokens container..."
docker run -d \
    --name supertokens-core \
    --network supertokens \
    -p 3567:3567 \
    -e POSTGRESQL_CONNECTION_URI="postgresql://supertokens:your-strong-password@supertokens-postgres:5432/supertokens" \
    -e API_KEYS="supertokens-long-api-key-123456789" \
    -e SUPERTOKENS_APP_NAME="Transcription Service" \
    registry.supertokens.io/supertokens/supertokens-postgresql

# Wait for SuperTokens to be ready
echo "Waiting for SuperTokens to be ready..."
WAIT_SECONDS=0
MAX_WAIT=30

while [ $WAIT_SECONDS -lt $MAX_WAIT ]; do
    if curl -s http://localhost:3567/hello > /dev/null; then
        echo "SuperTokens is ready!"
        break
    fi
    echo "Waiting for SuperTokens to be ready... ($WAIT_SECONDS seconds)"
    sleep 5
    WAIT_SECONDS=$((WAIT_SECONDS + 5))
done

if [ $WAIT_SECONDS -ge $MAX_WAIT ]; then
    echo "Error: SuperTokens failed to become ready within $MAX_WAIT seconds"
    echo "SuperTokens logs:"
    docker logs supertokens-core
    exit 1
fi

# Show running containers
echo -e "\nRunning containers:"
docker ps

echo -e "\nSetup complete!"
