#!/bin/bash
set -e  # Exit on error

SUPERTOKENS_BASE_URL="http://127.0.0.1:8080"
PUBLIC_IP="3.131.82.143"

echo "Testing SuperTokens setup..."

# Check if SuperTokens is running
echo "Checking if SuperTokens is running..."
HELLO_ENDPOINT="$SUPERTOKENS_BASE_URL/hello"
echo "Sending request to: $HELLO_ENDPOINT"
if curl -s "$HELLO_ENDPOINT" > /dev/null; then
    echo "SuperTokens /hello endpoint is working."
else
    echo "Error: SuperTokens /hello endpoint is not reachable."
    exit 1
fi

# Check tenant listing API
echo "Checking tenant listing API..."
TENANT_LIST_ENDPOINT="$SUPERTOKENS_BASE_URL/recipe/tenant/list"
TENANT_LIST_CURL="curl -s -o /dev/null -w \"%{http_code}\" -X GET \"$TENANT_LIST_ENDPOINT\" -H \"Content-Type: application/json\" -H \"api-key: supertokens-long-api-key-123456789\""
echo "Sending request to: $TENANT_LIST_ENDPOINT"
TENANT_RESPONSE=$(eval $TENANT_LIST_CURL)
if [ "$TENANT_RESPONSE" -eq 404 ]; then
    echo "Tenant listing API responded with 404 (expected if multitenancy is not configured)."
else
    echo "Tenant listing API returned HTTP $TENANT_RESPONSE"
fi

# Test login API
echo "Testing login API..."
LOGIN_ENDPOINT="$SUPERTOKENS_BASE_URL/auth/login"
LOGIN_PAYLOAD='{"email": "testuser@example.com"}'
echo "Sending request to: $LOGIN_ENDPOINT"
echo "Payload: $LOGIN_PAYLOAD"
LOGIN_RESPONSE_BODY=$(curl -s -X POST "$LOGIN_ENDPOINT" \
    -H "Content-Type: application/json" \
    -d "$LOGIN_PAYLOAD")
LOGIN_RESPONSE_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$LOGIN_ENDPOINT" \
    -H "Content-Type: application/json" \
    -d "$LOGIN_PAYLOAD")

if [ "$LOGIN_RESPONSE_CODE" -eq 200 ]; then
    echo "Login API responded successfully:"
    echo "$LOGIN_RESPONSE_BODY"

    # Extract and prepare the magic link for testing
    MAGIC_LINK=$(echo "$LOGIN_RESPONSE_BODY" | grep -o '"link":"[^"]*' | sed 's/"link":"//')
    FIXED_MAGIC_LINK=$(echo "$MAGIC_LINK" | sed "s/127.0.0.1/$PUBLIC_IP/" | sed "s/\\\u0026/\&/g")

    echo -e "\nNext Step:"
    echo "  1. Copy the following URL and paste it into your browser to test verification:"
    echo "     $FIXED_MAGIC_LINK"
    echo "  2. Ensure the IP address has been updated to $PUBLIC_IP."
    echo "  3. Replace '\\u0026' with '&' in the query string parameters if necessary."
else
    echo "Error: Login API returned HTTP $LOGIN_RESPONSE_CODE"
    echo "Response Body: $LOGIN_RESPONSE_BODY"
    echo "Debugging information:"
    echo "  Endpoint: $LOGIN_ENDPOINT"
    echo "  Payload: $LOGIN_PAYLOAD"
    echo "  Expected HTTP 200, but got $LOGIN_RESPONSE_CODE"
    echo "  Response Body: $LOGIN_RESPONSE_BODY"
    exit 1
fi

echo -e "\nAll tests completed successfully!"

