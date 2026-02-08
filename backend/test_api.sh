#!/bin/bash

# Taskalyst API Test Script
BASE_URL="http://localhost:8080"
EMAIL="test_$(date +%s)@example.com"
PASSWORD="password123"
NAME="Test User"

echo "--- 1. Auth ---"
echo "Signup..."
SIGNUP_RES=$(curl -s -X POST "$BASE_URL/auth/signup" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\", \"password\":\"$PASSWORD\", \"name\":\"$NAME\"}")
echo "$SIGNUP_RES" | jq .

echo "Login..."
LOGIN_RES=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\", \"password\":\"$PASSWORD\"}")
TOKEN=$(echo "$LOGIN_RES" | jq -r .access_token)
USER_ID=$(echo "$SIGNUP_RES" | jq -r .id)

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
  echo "Failed to get token"
  echo "$LOGIN_RES"
  exit 1
fi

AUTH_H="Authorization: Bearer $TOKEN"
JSON_H="Content-Type: application/json"

echo "--- 2. Categories & Projects ---"
echo "Create Category..."
CAT_RES=$(curl -s -X POST "$BASE_URL/api/categories" \
  -H "$AUTH_H" -H "$JSON_H" \
  -d '{"name":"Work", "root_type":"WORK", "color":"#0000FF"}')
CAT_ID=$(echo "$CAT_RES" | jq -r .id)
echo "$CAT_RES" | jq .

echo "Create Project..."
PROJ_RES=$(curl -s -X POST "$BASE_URL/api/projects" \
  -H "$AUTH_H" -H "$JSON_H" \
  -d "{\"category_id\":\"$CAT_ID\", \"title\":\"Backend Dev\", \"description\":\"API implementation\", \"color\":\"#FF0000\"}")
PROJ_ID=$(echo "$PROJ_RES" | jq -r .id)
echo "$PROJ_RES" | jq .

echo "--- 3. Tasks ---"
echo "Create Task..."
TASK_RES=$(curl -s -X POST "$BASE_URL/api/tasks" \
  -H "$AUTH_H" -H "$JSON_H" \
  -d "{\"project_id\":\"$PROJ_ID\", \"title\":\"Write Tests\", \"note\":\"Use curl for testing\"}")
TASK_ID=$(echo "$TASK_RES" | jq -r .id)
echo "$TASK_RES" | jq .

echo "--- 4. Time Tracking ---"
echo "Start Timer..."
TIME_RES=$(curl -s -X POST "$BASE_URL/api/time-entries" \
  -H "$AUTH_H" -H "$JSON_H" \
  -d "{\"project_id\":\"$PROJ_ID\", \"task_id\":\"$TASK_ID\", \"note\":\"Testing API\"}")
TIME_ID=$(echo "$TIME_RES" | jq -r .id)
echo "$TIME_RES" | jq .

echo "--- 5. Calendar & Events ---"
echo "Create Calendar..."
CAL_RES=$(curl -s -i -X POST "$BASE_URL/api/calendars" \
  -H "$AUTH_H" -H "$JSON_H" \
  -d "{\"name\":\"Work Cal\", \"project_id\":\"$PROJ_ID\", \"color\":\"#00FF00\", \"description\":\"Work events\"}")
echo "$CAL_RES"

CAL_ID=$(echo "$CAL_RES" | grep -oE '"id":"[^"]+"' | cut -d'"' -f4)

if [ -n "$CAL_ID" ]; then
    echo "Calendar ID: $CAL_ID"
else
    echo "Failed to create calendar"
fi

echo "Create Event..."
START=$(date -u -d "tomorrow 10:00" +"%Y-%m-%dT%H:%M:%SZ")
END=$(date -u -d "tomorrow 11:00" +"%Y-%m-%dT%H:%M:%SZ")
EVENT_RES=$(curl -s -i -X POST "$BASE_URL/api/events" \
  -H "$AUTH_H" -H "$JSON_H" \
  -d "{\"project_id\":\"$PROJ_ID\", \"title\":\"Scrum Meeting\", \"start_at\":\"$START\", \"end_at\":\"$END\"}")
echo "$EVENT_RES"

echo "--- 6. Results ---"
echo "Create Result..."
RES_RES=$(curl -s -X POST "$BASE_URL/api/results" \
  -H "$AUTH_H" -H "$JSON_H" \
  -d "{\"project_id\":\"$PROJ_ID\", \"type\":\"page_count\", \"value\":10}")
RES_ID=$(echo "$RES_RES" | jq -r .id)
echo "$RES_RES" | jq .

echo "--- 7. API Tokens ---"
echo "Create Token..."
TOKEN_RES=$(curl -s -X POST "$BASE_URL/api/tokens" \
  -H "$AUTH_H" -H "$JSON_H" \
  -d '{"name":"My PAT"}')
PAT_RAW=$(echo "$TOKEN_RES" | jq -r .token)
echo "$TOKEN_RES" | jq .

echo "Test PAT (Get Me)..."
curl -s -X GET "$BASE_URL/api/users/me" -H "X-API-KEY: $PAT_RAW" | jq .

echo "--- Done ---"
