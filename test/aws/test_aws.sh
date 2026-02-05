#!/bin/bash
# AWS CLI comparison tests using LocalStack
# Usage: ./test_aws.sh

set -e

LOCALSTACK_ENDPOINT="http://localhost:4566"
AWS_OPTS="--endpoint-url $LOCALSTACK_ENDPOINT --region us-east-1"
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
export AWS_DEFAULT_REGION=us-east-1

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

OMNI="../../omni"
PASSED=0
FAILED=0

# Build omni first
echo "Building omni..."
cd ../..
go build -o omni .
cd test/aws

compare_output() {
    local name="$1"
    local aws_cmd="$2"
    local omni_cmd="$3"

    echo -n "Testing: $name... "

    aws_out=$(eval "$aws_cmd" 2>&1 || true)
    omni_out=$(eval "$omni_cmd" 2>&1 || true)

    # For JSON output, normalize and compare
    if echo "$aws_out" | jq . > /dev/null 2>&1; then
        aws_normalized=$(echo "$aws_out" | jq -S '.' 2>/dev/null || echo "$aws_out")
        omni_normalized=$(echo "$omni_out" | jq -S '.' 2>/dev/null || echo "$omni_out")

        if [ "$aws_normalized" = "$omni_normalized" ]; then
            echo -e "${GREEN}PASS${NC}"
            ((PASSED++))
            return 0
        fi
    fi

    # For non-JSON, do simple comparison
    if [ "$aws_out" = "$omni_out" ]; then
        echo -e "${GREEN}PASS${NC}"
        ((PASSED++))
        return 0
    fi

    echo -e "${RED}FAIL${NC}"
    echo "  AWS output:"
    echo "$aws_out" | head -5 | sed 's/^/    /'
    echo "  Omni output:"
    echo "$omni_out" | head -5 | sed 's/^/    /'
    ((FAILED++))
    return 1
}

check_output() {
    local name="$1"
    local omni_cmd="$2"
    local expected_pattern="$3"

    echo -n "Testing: $name... "

    omni_out=$(eval "$omni_cmd" 2>&1 || true)

    if echo "$omni_out" | grep -q "$expected_pattern"; then
        echo -e "${GREEN}PASS${NC}"
        ((PASSED++))
        return 0
    fi

    echo -e "${RED}FAIL${NC}"
    echo "  Expected pattern: $expected_pattern"
    echo "  Omni output:"
    echo "$omni_out" | head -5 | sed 's/^/    /'
    ((FAILED++))
    return 1
}

echo ""
echo "========================================="
echo "AWS CLI Comparison Tests with LocalStack"
echo "========================================="
echo ""

# Wait for LocalStack to be ready
echo "Waiting for LocalStack..."
for i in {1..30}; do
    if curl -s "$LOCALSTACK_ENDPOINT/_localstack/health" | grep -q "running"; then
        echo "LocalStack is ready!"
        break
    fi
    sleep 1
done

echo ""
echo "--- STS Tests ---"

# STS get-caller-identity
check_output "sts get-caller-identity" \
    "$OMNI aws sts get-caller-identity --region us-east-1" \
    "Account"

echo ""
echo "--- S3 Tests ---"

# Create test bucket
TEST_BUCKET="omni-test-bucket-$$"
echo "Creating test bucket: $TEST_BUCKET"
aws $AWS_OPTS s3 mb "s3://$TEST_BUCKET" > /dev/null 2>&1 || true

# S3 ls (list buckets)
check_output "s3 ls (list buckets)" \
    "$OMNI aws s3 ls --region us-east-1" \
    "$TEST_BUCKET"

# Create test file
echo "test content" > /tmp/test_file.txt

# S3 cp (upload)
echo -n "Testing: s3 cp (upload)... "
$OMNI aws s3 cp /tmp/test_file.txt "s3://$TEST_BUCKET/test_file.txt" --region us-east-1 > /dev/null 2>&1
if aws $AWS_OPTS s3 ls "s3://$TEST_BUCKET/test_file.txt" > /dev/null 2>&1; then
    echo -e "${GREEN}PASS${NC}"
    ((PASSED++))
else
    echo -e "${RED}FAIL${NC}"
    ((FAILED++))
fi

# S3 ls (list objects)
check_output "s3 ls (list objects)" \
    "$OMNI aws s3 ls s3://$TEST_BUCKET/ --region us-east-1" \
    "test_file.txt"

# S3 cp (download)
echo -n "Testing: s3 cp (download)... "
rm -f /tmp/downloaded_file.txt
$OMNI aws s3 cp "s3://$TEST_BUCKET/test_file.txt" /tmp/downloaded_file.txt --region us-east-1 > /dev/null 2>&1
if [ -f /tmp/downloaded_file.txt ] && grep -q "test content" /tmp/downloaded_file.txt; then
    echo -e "${GREEN}PASS${NC}"
    ((PASSED++))
else
    echo -e "${RED}FAIL${NC}"
    ((FAILED++))
fi

# S3 rm
echo -n "Testing: s3 rm... "
$OMNI aws s3 rm "s3://$TEST_BUCKET/test_file.txt" --region us-east-1 > /dev/null 2>&1
if ! aws $AWS_OPTS s3 ls "s3://$TEST_BUCKET/test_file.txt" > /dev/null 2>&1; then
    echo -e "${GREEN}PASS${NC}"
    ((PASSED++))
else
    echo -e "${RED}FAIL${NC}"
    ((FAILED++))
fi

# S3 presign
check_output "s3 presign" \
    "$OMNI aws s3 presign s3://$TEST_BUCKET/file.txt --region us-east-1" \
    "http"

# S3 rb (remove bucket)
echo -n "Testing: s3 rb... "
$OMNI aws s3 rb "s3://$TEST_BUCKET" --region us-east-1 > /dev/null 2>&1
if ! aws $AWS_OPTS s3 ls "s3://$TEST_BUCKET" > /dev/null 2>&1; then
    echo -e "${GREEN}PASS${NC}"
    ((PASSED++))
else
    echo -e "${RED}FAIL${NC}"
    ((FAILED++))
fi

echo ""
echo "--- SSM Tests ---"

# SSM put-parameter
echo -n "Testing: ssm put-parameter... "
$OMNI aws ssm put-parameter --name "/test/param1" --value "testvalue" --type String --region us-east-1 > /dev/null 2>&1
if aws $AWS_OPTS ssm get-parameter --name "/test/param1" | grep -q "testvalue"; then
    echo -e "${GREEN}PASS${NC}"
    ((PASSED++))
else
    echo -e "${RED}FAIL${NC}"
    ((FAILED++))
fi

# SSM get-parameter
check_output "ssm get-parameter" \
    "$OMNI aws ssm get-parameter --name /test/param1 --region us-east-1" \
    "testvalue"

# SSM get-parameters-by-path
aws $AWS_OPTS ssm put-parameter --name "/test/param2" --value "value2" --type String > /dev/null 2>&1 || true
check_output "ssm get-parameters-by-path" \
    "$OMNI aws ssm get-parameters-by-path --path /test --region us-east-1" \
    "param"

echo ""
echo "========================================="
echo "Results: ${GREEN}$PASSED passed${NC}, ${RED}$FAILED failed${NC}"
echo "========================================="

# Cleanup
rm -f /tmp/test_file.txt /tmp/downloaded_file.txt

exit $FAILED
