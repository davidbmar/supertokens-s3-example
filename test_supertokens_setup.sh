#!/bin/bash
set -e  # Exit on error

SUPERTOKENS_BASE_URL="http://127.0.0.1:8080"

echo "Testing SuperTokens setup..."

# Check if SuperTokens is running
echo "Checking if SuperTokens is running..."
if curl -s "$SUPERTOKENS_BASE_URL/hello" > /dev/null; then
    echo "SuperTokens /hello endpoint is working."
else
    echo "Error: SuperTokens /hello endpoint is not reachable."
    exit 1
fi

# Check tenant listing API
echo "Checking tenant listing API..."
TENANT_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X GET "$SUPERTOKENS_BASE_URL/recipe/tenant/list" \
    -H "Content-Type: application/json" \
    -H "api-key: supertokens-long-api-key-123456789")
if [ "$TENANT_RESPONSE" -eq 404 ]; then
    echo "Tenant listing API responded with 404 (expected if multitenancy is not configured)."
else
    echo "Tenant listing API returned HTTP $TENANT_RESPONSE"
fi

# Test login API
echo "Testing login API..."
LOGIN_PAYLOAD='{"email": "testuser@example.com"}'

LOGIN_RESPONSE_BODY=$(curl -s -X POST "$SUPERTOKENS_BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d "$LOGIN_PAYLOAD")

LOGIN_RESPONSE_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$SUPERTOKENS_BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d "$LOGIN_PAYLOAD")

if [ "$LOGIN_RESPONSE_CODE" -eq 200 ]; then
    echo "Login API responded successfully:"
    echo "$LOGIN_RESPONSE_BODY"
else
    echo "Error: Login API returned HTTP $LOGIN_RESPONSE_CODE"
    echo "Response Body: $LOGIN_RESPONSE_BODY"
    echo "Debugging information:"
    echo "  Endpoint: $SUPERTOKENS_BASE_URL/auth/login"
    echo "  Payload: $LOGIN_PAYLOAD"
    echo "  Expected HTTP 200, but got $LOGIN_RESPONSE_CODE"
    echo "  Response Body: $LOGIN_RESPONSE_BODY"
    exit 1
fi

echo -e "\nAll tests completed successfully!"

