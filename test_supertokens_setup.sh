#!/bin/bash
set -e  # Exit on error

SUPERTOKENS_CORE_URL="http://127.0.0.1:3567"
GO_SERVICE_URL="http://127.0.0.1:8080"

# Detect if running on EC2 or Mac
if [ "$(uname)" == "Darwin" ]; then
    echo "Running test script on Mac. Using EC2 IP address."
    SUPERTOKENS_CORE_URL="http://<YOUR_EC2_IP>:3567" # Replace <YOUR_EC2_IP> with your EC2's public IP.
    GO_SERVICE_URL="http://<YOUR_EC2_IP>:8080"       # Replace <YOUR_EC2_IP> with your EC2's public IP.
else
    if grep -q "ec2" /sys/hypervisor/uuid 2>/dev/null ||
       [ -d /sys/devices/virtual/dmi/id ] && grep -q "amazon" /sys/devices/virtual/dmi/id/sys_vendor; then
        echo "Running test script on EC2, assuming localhost."
    else
        echo "Could not determine environment. Assuming EC2 and using localhost."
    fi
fi

echo "Using SuperTokens Core URL: $SUPERTOKENS_CORE_URL"
echo "Using Go Service URL: $GO_SERVICE_URL"
echo "Testing SuperTokens setup..."

# Check if SuperTokens Core is running
echo "Checking if SuperTokens Core is running..."
HELLO_RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" "$SUPERTOKENS_CORE_URL/hello")
HELLO_HTTP_CODE=$(echo "$HELLO_RESPONSE" | grep HTTP_CODE | cut -d':' -f2)
HELLO_BODY=$(echo "$HELLO_RESPONSE" | sed '/HTTP_CODE/d')

if [ "$HELLO_HTTP_CODE" -eq 200 ]; then
    echo "SuperTokens Core /hello endpoint is working."
else
    echo "Error: SuperTokens Core /hello endpoint returned HTTP $HELLO_HTTP_CODE"
    echo "Response Body: $HELLO_BODY"
    docker logs supertokens-core | tail -n 50
    exit 1
fi

# Check if Go service is running
echo "Checking if Go service is running..."
if ! curl -s -o /dev/null "$GO_SERVICE_URL/hello"; then
    echo "Error: Go service is not reachable at $GO_SERVICE_URL."
    echo "Please ensure the Go service is running and try again. (go run cmd/server/main.go)"
    exit 1
fi

# Check tenant listing API (SuperTokens Core)
echo "Checking tenant listing API..."
TENANT_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X GET "$SUPERTOKENS_CORE_URL/recipe/tenant/list" \
    -H "Content-Type: application/json" \
    -H "api-key: supertokens-long-api-key-123456789")
if [ "$TENANT_RESPONSE" -eq 404 ]; then
    echo "Tenant listing API responded with 404 (expected if multitenancy is not configured)."
else
    echo "Tenant listing API returned HTTP $TENANT_RESPONSE"
    docker logs supertokens-core | tail -n 50
    exit 1
fi

# Test login API (Go Service)
echo "Testing login API via Go service..."
LOGIN_PAYLOAD='{"email": "testuser@example.com"}'

LOGIN_RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X POST "$GO_SERVICE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d "$LOGIN_PAYLOAD")
LOGIN_HTTP_CODE=$(echo "$LOGIN_RESPONSE" | grep HTTP_CODE | cut -d':' -f2)
LOGIN_BODY=$(echo "$LOGIN_RESPONSE" | sed '/HTTP_CODE/d')

if [ "$LOGIN_HTTP_CODE" -eq 200 ]; then
    echo "Login API responded successfully:"
    echo "$LOGIN_BODY"
else
    echo "Error: Login API returned HTTP $LOGIN_HTTP_CODE"
    echo "Response Body: $LOGIN_BODY"
    echo "Debugging information:"
    echo "  Endpoint: $GO_SERVICE_URL/auth/login"
    echo "  Payload: $LOGIN_PAYLOAD"
    echo "  Expected HTTP 200, but got $LOGIN_HTTP_CODE"
    echo "  Response Body: $LOGIN_BODY"

    # Print Go Service container logs for debugging
    echo "Go Service logs:"
    docker logs $(docker ps --filter "name=transcription-service" -q) | tail -n 50
    exit 1
fi

echo -e "\nAll tests completed successfully!"

