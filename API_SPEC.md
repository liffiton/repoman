# Repoman API Specification

This document defines the REST API endpoints required by the Repoman CLI tool to manage student repositories.

## Base URL
All endpoints are prefixed with `/api/v1` (or as configured).
Example: `https://repoman.example.com/api/v1`

## Authentication
All requests must include a Bearer token in the `Authorization` header.
```http
Authorization: Bearer <your_api_key>
```

---

## 1. Courses
### `GET /courses`
Returns a list of courses the authenticated user has access to.

**Response Body:**
```json
[
  {
    "id": "string",
    "name": "string"
  }
]
```

---

## 2. Assignments
### `GET /courses/{course_id}/assignments`
Returns a list of assignments for a specific course.

**Response Body:**
```json
[
  {
    "id": "string",
    "name": "string"
  }
]
```

---

## 3. Repositories
### `GET /assignments/{assignment_id}/repos`
Returns all student repositories for a specific assignment.

**Response Body:**
```json
[
  {
    "name": "string",
    "url": "string"
  }
]
```
*Note: The `name` field should ideally be unique within the assignment (e.g., `lab1-studentusername`) as Repoman uses this for the local directory name.*

---

## Error Handling
Standard HTTP status codes should be used:
- `200 OK`: Success.
- `401 Unauthorized`: Invalid or missing API key.
- `403 Forbidden`: User does not have access to the resource.
- `404 Not Found`: Resource (Course or Assignment) not found.
- `500 Internal Server Error`: Server-side failure.
