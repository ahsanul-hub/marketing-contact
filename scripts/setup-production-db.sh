#!/bin/bash

# Script untuk setup database production dari awal
# Usage: ./scripts/setup-production-db.sh

set -e

echo "🚀 Setting up production database..."

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Default values (ubah sesuai kebutuhan)
DB_NAME="${DB_NAME:-phone_data}"
DB_USER="${DB_USER:-marketing_user}"
DB_PASSWORD="${DB_PASSWORD:-}"

if [ -z "$DB_PASSWORD" ]; then
    echo -e "${RED}❌ Error: DB_PASSWORD is not set${NC}"
    echo "Please set it with:"
    echo "  export DB_PASSWORD='your_secure_password'"
    exit 1
fi

echo -e "${GREEN}📝 Database Configuration:${NC}"
echo "  Database: $DB_NAME"
echo "  User: $DB_USER"
echo ""

# Check if running as postgres user or with sudo
if [ "$EUID" -ne 0 ] && [ "$USER" != "postgres" ]; then
    echo -e "${YELLOW}⚠ This script needs to run as postgres user or with sudo${NC}"
    echo "Running with sudo..."
    SUDO_CMD="sudo -u postgres"
else
    SUDO_CMD=""
fi

# Create database and user
echo -e "${GREEN}🔧 Creating database and user...${NC}"

$SUDO_CMD psql <<EOF
-- Create database
CREATE DATABASE $DB_NAME;

-- Create user
CREATE USER $DB_USER WITH PASSWORD '$DB_PASSWORD';

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;
EOF

# Grant schema privileges (PostgreSQL 15+)
echo -e "${GREEN}🔧 Granting schema privileges...${NC}"
$SUDO_CMD psql -d $DB_NAME <<EOF
GRANT ALL ON SCHEMA public TO $DB_USER;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO $DB_USER;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO $DB_USER;
EOF

echo -e "${GREEN}✅ Database setup completed!${NC}"
echo ""
echo -e "${GREEN}📋 Next steps:${NC}"
echo "1. Set DATABASE_URL:"
echo "   export DATABASE_URL='postgresql://$DB_USER:$DB_PASSWORD@localhost:5432/$DB_NAME'"
echo ""
echo "2. Run migration:"
echo "   npx prisma migrate deploy"
echo ""

