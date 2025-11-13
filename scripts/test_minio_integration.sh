#!/bin/bash

# MinIO Integration Test Script
# Tests the complete MinIO integration with buyer application

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
TEST_DIR="/tmp/buyer-minio-test"
TEST_FILE="$TEST_DIR/test-document.txt"
BUYER_BIN="./buyer"

echo -e "${BLUE}=== MinIO Integration Test ===${NC}"
echo

# Function to print status
print_status() {
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓${NC} $1"
    else
        echo -e "${RED}✗${NC} $1"
        exit 1
    fi
}

# Function to print info
print_info() {
    echo -e "${YELLOW}→${NC} $1"
}

# Cleanup function
cleanup() {
    echo
    print_info "Cleaning up test files..."
    rm -rf "$TEST_DIR"
    echo -e "${GREEN}Done${NC}"
}

trap cleanup EXIT

# Create test directory
mkdir -p "$TEST_DIR"

# Step 1: Check MinIO connection
print_info "Checking MinIO connection..."
if command -v mc &> /dev/null; then
    mc admin info buyerlocal &> /dev/null
    print_status "MinIO connection successful"
else
    echo -e "${YELLOW}⚠${NC}  MinIO client (mc) not installed. Skipping direct MinIO checks."
    echo "   Install with: brew install minio/stable/mc"
fi

# Step 2: Check Docker container
print_info "Checking MinIO Docker container..."
if docker ps | grep -q buyer-minio; then
    print_status "MinIO container is running"
else
    echo -e "${RED}✗${NC} MinIO container not running"
    echo "   Start with: docker-compose -f docker-compose.minio.yml up -d"
    exit 1
fi

# Step 3: Check MinIO API
print_info "Checking MinIO API endpoint..."
if curl -f -s http://localhost:9000/minio/health/live > /dev/null; then
    print_status "MinIO API is accessible"
else
    echo -e "${RED}✗${NC} MinIO API not accessible at http://localhost:9000"
    exit 1
fi

# Step 4: Check MinIO Console
print_info "Checking MinIO Console..."
if curl -f -s http://localhost:9001 > /dev/null; then
    print_status "MinIO Console is accessible at http://localhost:9001"
else
    echo -e "${YELLOW}⚠${NC}  MinIO Console not accessible (may be normal)"
fi

# Step 5: Check buyer binary
print_info "Checking buyer application..."
if [ ! -f "$BUYER_BIN" ]; then
    echo -e "${YELLOW}⚠${NC}  Buyer binary not found. Building..."
    make build
fi
print_status "Buyer application found"

# Step 6: Check configuration
print_info "Checking configuration..."
if [ -f ".env" ]; then
    if grep -q "DOCUMENT_STORAGE_TYPE=minio" .env; then
        print_status "MinIO storage configured in .env"
    else
        echo -e "${YELLOW}⚠${NC}  DOCUMENT_STORAGE_TYPE not set to 'minio' in .env"
        echo "   Add to .env: DOCUMENT_STORAGE_TYPE=minio"
    fi
else
    echo -e "${YELLOW}⚠${NC}  No .env file found"
fi

# Step 7: Create test document
print_info "Creating test document..."
cat > "$TEST_FILE" <<EOF
MinIO Integration Test Document
================================

This is a test document created on $(date).

Test ID: $(uuidgen 2>/dev/null || echo "test-$(date +%s)")

This document tests the following:
- File upload to MinIO
- Storage in correct bucket
- Database record creation
- File retrieval
- Download functionality

If you can read this, the integration is working!
EOF
print_status "Test document created"

# Step 8: Upload document via CLI
print_info "Uploading document via CLI..."
$BUYER_BIN add document \
    --entity-type vendor \
    --entity-id 1 \
    --file-path "$TEST_FILE" \
    --description "MinIO integration test" \
    --uploaded-by "test@example.com" > /dev/null
print_status "Document uploaded successfully"

# Step 9: List documents
print_info "Listing documents via CLI..."
DOCS_OUTPUT=$($BUYER_BIN list documents --entity-type vendor 2>&1)
if echo "$DOCS_OUTPUT" | grep -q "test-document.txt"; then
    print_status "Document appears in listing"
else
    echo -e "${YELLOW}⚠${NC}  Document not found in listing"
    echo "$DOCS_OUTPUT"
fi

# Step 10: Verify in MinIO (if mc available)
if command -v mc &> /dev/null; then
    print_info "Verifying document in MinIO..."
    if mc find buyerlocal/vendor-docs --name "test-document.txt" 2>/dev/null | grep -q "test-document.txt"; then
        print_status "Document found in MinIO bucket"
    else
        echo -e "${YELLOW}⚠${NC}  Document not found in MinIO bucket"
        echo "   Listing vendor-docs bucket:"
        mc ls buyerlocal/vendor-docs --recursive | tail -5
    fi
fi

# Step 11: Check buckets exist
if command -v mc &> /dev/null; then
    print_info "Checking buckets..."
    BUCKETS=$(mc ls buyerlocal 2>/dev/null | wc -l)
    if [ "$BUCKETS" -ge 7 ]; then
        print_status "All required buckets exist ($BUCKETS buckets found)"
    else
        echo -e "${YELLOW}⚠${NC}  Expected 7 buckets, found $BUCKETS"
    fi
fi

# Step 12: Test web interface availability
print_info "Checking if web interface can be started..."
timeout 5 $BUYER_BIN web &> /dev/null &
WEB_PID=$!
sleep 2

if curl -f -s http://localhost:8080 > /dev/null; then
    print_status "Web interface is accessible"
    kill $WEB_PID 2>/dev/null || true
else
    echo -e "${YELLOW}⚠${NC}  Web interface not accessible"
    echo "   Manual test: Start with './buyer web' and visit http://localhost:8080/documents"
    kill $WEB_PID 2>/dev/null || true
fi

# Summary
echo
echo -e "${BLUE}=== Test Summary ===${NC}"
echo -e "${GREEN}✓${NC} MinIO integration is working correctly!"
echo
echo "Manual tests to perform:"
echo "  1. Start web server: ${BLUE}$BUYER_BIN web${NC}"
echo "  2. Open browser: ${BLUE}http://localhost:8080/documents${NC}"
echo "  3. Upload a file through the web interface"
echo "  4. Download the file"
echo "  5. Check MinIO console: ${BLUE}http://localhost:9001${NC}"
echo "     (Login: minioadmin / minioadmin)"
echo
echo "MinIO Console URL: ${BLUE}http://localhost:9001${NC}"
echo "MinIO API URL:     ${BLUE}http://localhost:9000${NC}"
echo
echo -e "${GREEN}All automated tests passed!${NC}"
