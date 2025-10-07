#!/usr/bin/env bash
set -euo pipefail
which jq >/dev/null 2>&1 || { echo "Please install jq"; exit 1; }

AUTH_PORT="${AUTH_PORT:-8082}"
PRODUCT_PORT="${PRODUCT_PORT:-8084}"
ORDER_PORT="${ORDER_PORT:-8083}"
BASE_HOST="${BASE_HOST:-localhost}"

echo "== Signup =="
curl -sS -X POST "http://${BASE_HOST}:${AUTH_PORT}/api/v1/user/register"   -H "Content-Type: application/json"   -d '{"email":"tester@example.com","password":"StrongP@ssw0rd"}' || true
echo "----- == ----- XX ----- == ----- "


echo "== Login =="
AUTH_JSON="$(curl -sS -X POST "http://${BASE_HOST}:${AUTH_PORT}/api/v1/user/login"   -H "Content-Type: application/json"   -d '{"email":"tester@example.com","password":"StrongP@ssw0rd"}')"
#echo "Raw response: $AUTH_JSON"
AUTH_TOKEN="$(echo "$AUTH_JSON" | jq -r '.result.accessToken // .result.jwt')"
AUTH_REFRESH_TOKEN="$(echo "$AUTH_JSON" | jq -r '.result.token // .result.refresh_token // .result.refreshToken // .result.jwt')"
if [[ -z "${AUTH_TOKEN}" || "${AUTH_TOKEN}" == "null" ]]; then
  echo "Login did not return a token"; exit 1
fi
echo "Token acquired"
echo "----- == ----- XX ----- == ----- "


echo "== Create product =="
PRODUCT_JSON="$(curl -sS -X POST "http://${BASE_HOST}:${PRODUCT_PORT}/api/v1/products" \
  -H "Authorization: Bearer ${AUTH_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"name":"Iphone 15 air","sku":"iph2025","description":"asdasdqwdasd","price":85000,"stock":200}')"
#echo "Raw response: $PRODUCT_JSON"
PRODUCT_ID="$(echo "$PRODUCT_JSON" | jq -r '.result.id // .result.product.id')"
echo "Product ID: ${PRODUCT_ID}"
echo "----- == ----- XX ----- == ----- "

echo "== List products =="
curl -sS "http://${BASE_HOST}:${PRODUCT_PORT}/api/v1/products/list" -H "Authorization: Bearer ${AUTH_TOKEN}" | jq -r ' .result[0] // .products[0]'
echo "----- == ----- XX ----- == ----- "

echo "== Create order =="
ORDER_JSON="$(curl -sS -X POST "http://${BASE_HOST}:${ORDER_PORT}/api/v1/orders"   -H "Authorization: Bearer ${AUTH_TOKEN}"   -H "Content-Type: application/json"   -d "{\"items\":[{\"id\":\"${PRODUCT_ID}\",\"qty\":1, \"price\":85000}]}")"
#echo "Raw response: $ORDER_JSON"
ORDER_ID="$(echo "$ORDER_JSON" | jq -r '.result.id // .order.id // .data.id')"
echo "Order ID: ${ORDER_ID}"
echo "----- == ----- XX ----- == ----- "

echo "== Get order =="
curl -sS "http://${BASE_HOST}:${ORDER_PORT}/api/v1/orders/${ORDER_ID}" -H "Authorization: Bearer ${AUTH_TOKEN}" | jq .
echo "----- == ----- XX ----- == ----- "

echo "== Logout =="
curl -sS -X POST "http://${BASE_HOST}:${AUTH_PORT}/api/v1/user/logout"   -H "Content-Type: application/json"   -d "{\"refreshToken\":\"${AUTH_REFRESH_TOKEN}\"}" | jq
echo "----- == ----- XX ----- == ----- "


echo "E2E flow completed."
