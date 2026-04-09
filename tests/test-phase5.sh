#!/bin/bash
set -e

BASE="http://localhost:8080/api"

echo "=== Test Phase 5 ==="

# Login
echo "1. Login..."
RESP=$(curl -s $BASE/auth/login -X POST -H 'Content-Type: application/json' -d '{"username":"admin","password":"admin123"}')
echo "   Response: $RESP"
TOKEN=$(echo "$RESP" | python3 -c 'import sys,json; print(json.loads(sys.stdin.read()).get("token",""))' 2>/dev/null || echo "")

if [ -z "$TOKEN" ]; then
    echo "   Login failed, trying setup..."
    RESP=$(curl -s $BASE/auth/setup -X POST -H 'Content-Type: application/json' -d '{"username":"admin","password":"admin123"}')
    echo "   Setup Response: $RESP"
    TOKEN=$(echo "$RESP" | python3 -c 'import sys,json; print(json.loads(sys.stdin.read()).get("token",""))' 2>/dev/null || echo "")
fi

if [ -z "$TOKEN" ]; then
    echo "FATAL: Could not get auth token"
    exit 1
fi
echo "   Token: ${TOKEN:0:16}..."
AUTH="Authorization: Bearer $TOKEN"

# Test Backups
echo ""
echo "2. Create backup..."
RESP=$(curl -s $BASE/backups -X POST -H "$AUTH")
echo "   $RESP"

echo "3. List backups..."
RESP=$(curl -s $BASE/backups -H "$AUTH")
echo "   $RESP"

# Test Users
echo ""
echo "4. List users..."
RESP=$(curl -s $BASE/users -H "$AUTH")
echo "   $RESP"

# Test Invite
echo "5. Invite user..."
RESP=$(curl -s $BASE/invitations -X POST -H "$AUTH" -H 'Content-Type: application/json' -d '{"email":"test@example.com","role":"member"}')
echo "   $RESP"

echo "6. List invitations..."
RESP=$(curl -s $BASE/invitations -H "$AUTH")
echo "   $RESP"

# Test Audit Logs
echo ""
echo "7. Audit logs..."
RESP=$(curl -s $BASE/audit -H "$AUTH")
echo "   $RESP"

# Test Marketplace
echo ""
echo "8. List templates..."
RESP=$(curl -s $BASE/marketplace -H "$AUTH")
echo "   $(echo "$RESP" | python3 -c 'import sys,json; data=json.loads(sys.stdin.read()); print(f"{len(data)} templates: {[t[\"name\"] for t in data]}")')"

# Test Volumes
echo ""
echo "9. Test volumes (need a project)..."
PROJECTS=$(curl -s $BASE/projects -H "$AUTH")
echo "   Projects: $PROJECTS"

echo ""
echo "=== Phase 5 Tests Complete ==="
