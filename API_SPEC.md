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
Returns a list of courses the authenticated user has access to (owns or has staff access to).

**Response Body:**
```json
[
  {
    "id": "string",
    "name": "string",
    "archived": false,
    "archived_at": null
  }
]
```
- `archived_at` is an ISO 8601 timestamp, or `null` if the course has never been archived.

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
Returns all student repositories for a specific assignment. Repositories that have not yet been created on GitHub (no URL available) are omitted.

**Response Body:**
```json
[
  {
    "name": "string",
    "url": "string"
  }
]
```
- `name` is the repository's name on GitHub (the slugified name under the course's organization). It is unique within the assignment and safe to use as a local directory name. Repoman uses it as the local directory name.
- `url` is the repository's HTTPS clone URL on GitHub.

---

## Error Handling
Standard HTTP status codes are used:
- `200 OK`: Success.
- `401 Unauthorized`: Invalid or missing API key.
- `403 Forbidden`: User does not have access to the resource.
- `404 Not Found`: Resource (Course or Assignment) not found.
- `500 Internal Server Error`: Server-side failure.

Error responses are JSON objects:
```json
{
  "error": "Not Found",
  "message": "Course not found"
}
```
