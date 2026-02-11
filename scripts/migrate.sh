#!/bin/bash

# Warna untuk output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

case $1 in
  "up")
    echo -e "${YELLOW} Running migrations...${NC}"
    go run cmd/migrate/main.go -cmd up
    ;;
    
  "reset")
    echo -e "${RED}Resetting database...${NC}"
    go run cmd/migrate/main.go -cmd reset
    ;;
    
  "fresh")
    echo -e "${YELLOW}Fresh migration...${NC}"
    go run cmd/migrate/main.go -cmd fresh
    ;;
    
  "status")
    echo -e "${YELLOW}Migration status:${NC}"
    psql $(grep DB_CONN .env.development | cut -d= -f2) -c "SELECT version, name, applied_at FROM schema_migrations ORDER BY applied_at;"
    ;;
    
  *)
    echo "Usage: ./scripts/migrate.sh {up|reset|fresh|status}"
    exit 1
    ;;
esac