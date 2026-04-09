#!/bin/bash
set -e
BASE="http://localhost:8080/api"

# Setup + get token
RESP=$(curl -s $BASE/auth/status)
echo "Auth status: $RESP"

TOKEN="803b93e00174cce061356a7d24c369ed9fa469ec5c49f6e3b4504e08fa17f6e5"
AUTH="Authorization: Bearer $TOKEN"

# Test invite registration
echo ""
echo "=== Test Invite Registration ==="
INVITATIONS=$(curl -s $BASE/invitations -H "$AUTH")
INV_TOKEN=$(echo "$INVITATIONS" | python3 -c 'import sys,json; invs=json.loads(sys.stdin.read()); print(invs[0]["token"] if invs else "")')
echo "Invite token: ${INV_TOKEN:0:16}..."

REG_RESP=$(curl -s $BASE/auth/register -X POST -H 'Content-Type: application/json' -d "{\"token\":\"$INV_TOKEN\",\"username\":\"member1\",\"password\":\"pass123\"}")
echo "Register: $REG_RESP"

echo ""
echo "=== Test Marketplace ==="
TEMPLATES=$(curl -s $BASE/marketplace -H "$AUTH")
echo "Templates count: $(echo "$TEMPLATES" | python3 -c 'import sys,json; print(len(json.loads(sys.stdin.read())))')"

# Deploy a template
DEPLOY_RESP=$(curl -s "$BASE/marketplace/redis-cache/deploy" -X POST -H "$AUTH" -H 'Content-Type: application/json' -d '{"name":"my-redis"}')
echo "Deploy template: $DEPLOY_RESP"

echo ""
echo "=== Check Users After Registration ==="
USERS=$(curl -s $BASE/users -H "$AUTH")
echo "Users: $USERS"

echo ""
echo "=== Check Services After Template Deploy ==="
SERVICES=$(curl -s $BASE/../api/services -H "$AUTH")
echo "Services: $SERVICES"

echo ""
echo "=== RBAC Test: member tries admin endpoint ==="
MEMBER_TOKEN=$(echo "$REG_RESP" | python3 -c 'import sys,json; print(json.loads(sys.stdin.read()).get("token",""))')
MEMBER_AUTH="Authorization: Bearer $MEMBER_TOKEN"
RBAC_RESP=$(curl -s $BASE/users -H "$MEMBER_AUTH")
echo "Member accessing /users: $RBAC_RESP"

echo ""
echo "=== All Phase 5 Tests Done ==="
