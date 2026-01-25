#!/bin/bash

# Script untuk migrate database di VM production
# Usage: ./scripts/migrate-production.sh

set -e  # Exit on error

echo "🚀 Starting database migration for production..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if DATABASE_URL is set
if [ -z "$DATABASE_URL" ]; then
    echo -e "${RED}❌ Error: DATABASE_URL environment variable is not set${NC}"
    echo "Please set it with:"
    echo "  export DATABASE_URL='postgresql://user:password@host:5432/database'"
    exit 1
fi

echo -e "${GREEN}✓ DATABASE_URL is set${NC}"

# Check if Prisma is installed
if ! command -v npx &> /dev/null; then
    echo -e "${RED}❌ Error: npx is not installed${NC}"
    exit 1
fi

# Check if node_modules exists
if [ ! -d "node_modules" ]; then
    echo -e "${YELLOW}⚠ node_modules not found. Installing dependencies...${NC}"
    npm install
fi

# Check if Prisma is installed
if [ ! -d "node_modules/.prisma" ] && [ ! -f "node_modules/@prisma/client" ]; then
    echo -e "${YELLOW}⚠ Prisma Client not found. Generating...${NC}"
    npx prisma generate
fi

# Check migration status
echo -e "${GREEN}📊 Checking migration status...${NC}"
npx prisma migrate status

# Deploy migrations
echo -e "${GREEN}🔄 Deploying migrations...${NC}"
npx prisma migrate deploy

echo -e "${GREEN}✅ Migration completed successfully!${NC}"

# Optional: Verify connection
echo -e "${GREEN}🔍 Verifying database connection...${NC}"
npx prisma db execute --stdin <<< "SELECT version();" || echo -e "${YELLOW}⚠ Could not verify connection (this is okay)${NC}"

echo -e "${GREEN}✨ All done!${NC}"

