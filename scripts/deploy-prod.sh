#!/bin/bash
# deploy-prod.sh

set -e

echo "deploying Kasir API to Production..."

# 1. Pull latest code
git pull origin main

# 2. Load environment variables
set -a
source .env.production
set +a

# 3. Build and run with production compose
docker-compose -f docker-compose.prod.yml down -v
docker-compose -f docker-compose.prod.yml build --no-cache
docker-compose -f docker-compose.prod.yml up -d

# 4. Check status
echo "container Status:"
docker-compose -f docker-compose.prod.yml ps

# 5. Show logs
echo "recent logs:"
docker-compose -f docker-compose.prod.yml logs --tail=50 api

echo "deployment complete!"
echo "API running at: http://localhost:8080"