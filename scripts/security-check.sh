#!/bin/bash

# Security Pre-flight Check Script
# This script validates critical security configurations before starting the server

set -e

echo "🔒 Running Security Pre-flight Checks..."

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check JWT_SECRET
if [ -z "$JWT_SECRET" ]; then
    echo -e "${RED}❌ CRITICAL SECURITY ERROR: JWT_SECRET environment variable is not set${NC}"
    echo -e "${YELLOW}Generate a secure key with: openssl rand -hex 32${NC}"
    exit 1
fi

if [ ${#JWT_SECRET} -lt 32 ]; then
    echo -e "${RED}❌ CRITICAL SECURITY ERROR: JWT_SECRET is too short (${#JWT_SECRET} chars)${NC}"
    echo -e "${YELLOW}JWT_SECRET must be at least 32 characters long${NC}"
    echo -e "${YELLOW}Generate a secure key with: openssl rand -hex 32${NC}"
    exit 1
fi

if [ "$JWT_SECRET" = "your-super-secret-jwt-key" ] || [ "$JWT_SECRET" = "REPLACE_WITH_SECURE_RANDOM_32PLUS_CHAR_KEY" ] || [ "$JWT_SECRET" = "your-super-secret-jwt-key-make-it-strong-and-long" ]; then
    echo -e "${RED}❌ CRITICAL SECURITY ERROR: Using default/placeholder JWT_SECRET${NC}"
    echo -e "${YELLOW}Generate a secure key with: openssl rand -hex 32${NC}"
    exit 1
fi

echo -e "${GREEN}✅ JWT_SECRET is properly configured${NC}"

# Check database password (warn if empty)
if [ -z "$DB_PASSWORD" ]; then
    echo -e "${YELLOW}⚠️  WARNING: DB_PASSWORD is empty (development only)${NC}"
fi

# Check Redis password (warn if empty)
if [ -z "$REDIS_PASSWORD" ]; then
    echo -e "${YELLOW}⚠️  WARNING: REDIS_PASSWORD is empty (development only)${NC}"
fi

# Check if running in production
if [ "$ENV" = "production" ] || [ "$ENVIRONMENT" = "production" ] || [ "$GO_ENV" = "production" ]; then
    echo -e "${YELLOW}🏭 Production environment detected - additional checks...${NC}"
    
    # Production-specific checks
    if [ "$LOG_LEVEL" = "debug" ]; then
        echo -e "${YELLOW}⚠️  WARNING: Debug logging enabled in production${NC}"
    fi
    
    if [ "$SERVER_TOKEN" = "" ]; then
        echo -e "${YELLOW}⚠️  WARNING: SERVER_API_TOKEN not set${NC}"
    fi
fi

echo -e "${GREEN}🚀 Security checks passed! Server can start safely.${NC}"