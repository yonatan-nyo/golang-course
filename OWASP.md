1. A01:2021 - Broken Access Control
Test Case: Access admin-only endpoints without proper authorization

# Test accessing user management without admin privileges
curl -X GET "http://localhost:8080/api/users" 

# Test accessing other users' data
curl -X GET "http://localhost:8080/api/users/1" 
curl -X PUT "http://localhost:8080/api/users/2" \
  -H "Content-Type: application/json" \
  -d '{"name":"modified","identifier":"hacked123"}'

# Test course management without proper role
curl -X POST "http://localhost:8080/api/courses" \
  -H "Content-Type: application/json" \
  -d '{"title":"Hacked Course","description":"Unauthorized course creation"}'

### Aman karena make authorization dan authentication middleware

2. A03:2021 - Injection (SQL Injection)
Test Case: SQL injection in API parameters and request bodies

# Test authentication endpoints
curl -X POST "http://localhost:8080/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"identifier":"admin'\'' OR '\''1'\''='\''1","password":"anything"}'

# Test course search/filter parameters (alternative method)
curl -X GET --get "http://localhost:8080/api/courses" --data-urlencode "search=test';DROP TABLE courses;--"
curl -X GET --get "http://localhost:8080/api/courses" --data-urlencode "filter=1' UNION SELECT * FROM users--"

# Test module endpoints with injection (alternative method)
curl -X GET --get "http://localhost:8080/api/modules" --data-urlencode "course_id=1' OR '1'='1"

# Test user endpoints
curl -X GET --get "http://localhost:8080/api/users" --data-urlencode "search=admin' UNION SELECT password FROM users WHERE identifier='admin'--"

### aman karena make gorm

3. A07:2021 - Identification and Authentication Failures
Test Case: Weak authentication mechanisms and session handling

# Test JWT token manipulation
curl -X GET "http://localhost:8080/api/users/profile" \
  -H "Authorization: Bearer eyJhbGciOiJub25lIn0.eyJ1c2VyX2lkIjoxfQ."

# Test brute force protection
for i in {1..100}; do
  curl -X POST "http://localhost:8080/api/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"identifier":"admin","password":"wrong'$i'"}'
done

### go strong!