# REST API Contracts

**Feature**: 001-cognitive-backbone-poc  
**User Story**: US2 - Graph Query and Exploration  
**Generated**: 2026-01-24

## Base URL

```
http://localhost:8080/api/v1
```

## Endpoints

### 1. Get Entity by ID

**Endpoint**: `GET /entities/:id`

**Description**: Retrieve a single entity by its ID

**Path Parameters**:
- `id` (integer, required): Entity ID

**Response**:
- **200 OK**: Entity found
  ```json
  {
    "id": 123,
    "unique_id": "jeff.skilling@enron.com",
    "type_category": "person",
    "name": "Jeff Skilling",
    "properties": {
      "title": "CEO",
      "department": "Executive"
    },
    "confidence_score": 0.95,
    "created_at": "2026-01-20T10:30:00Z"
  }
  ```
- **404 Not Found**: Entity does not exist
  ```json
  {
    "error": "entity not found",
    "id": 123
  }
  ```
- **400 Bad Request**: Invalid ID format
  ```json
  {
    "error": "invalid entity id"
  }
  ```

---

### 2. Search Entities

**Endpoint**: `GET /entities`

**Description**: Search and filter entities by type, name, or other criteria

**Query Parameters**:
- `type` (string, optional): Filter by entity type category (e.g., "person", "organization", "concept")
- `name` (string, optional): Filter by name (case-insensitive partial match)
- `min_confidence` (float, optional): Minimum confidence score (0.0-1.0)
- `limit` (integer, optional): Maximum results to return (default: 100, max: 1000)
- `offset` (integer, optional): Pagination offset (default: 0)

**Examples**:
- `/entities?type=person&name=jeff`
- `/entities?min_confidence=0.8&limit=50`

**Response**:
- **200 OK**: Entities found (may be empty array)
  ```json
  {
    "entities": [
      {
        "id": 123,
        "unique_id": "jeff.skilling@enron.com",
        "type_category": "person",
        "name": "Jeff Skilling",
        "properties": {
          "title": "CEO"
        },
        "confidence_score": 0.95,
        "created_at": "2026-01-20T10:30:00Z"
      },
      {
        "id": 456,
        "unique_id": "jeff.dasovich@enron.com",
        "type_category": "person",
        "name": "Jeff Dasovich",
        "properties": {},
        "confidence_score": 0.88,
        "created_at": "2026-01-20T11:15:00Z"
      }
    ],
    "total": 2,
    "limit": 100,
    "offset": 0
  }
  ```
- **400 Bad Request**: Invalid query parameters
  ```json
  {
    "error": "invalid query parameter",
    "details": "min_confidence must be between 0 and 1"
  }
  ```

---

### 3. Get Entity Relationships

**Endpoint**: `GET /entities/:id/relationships`

**Description**: Get all relationships for an entity

**Path Parameters**:
- `id` (integer, required): Entity ID

**Query Parameters**:
- `type` (string, optional): Filter by relationship type (e.g., "SENT", "RECEIVED", "MENTIONS")
- `limit` (integer, optional): Maximum results (default: 100, max: 1000)
- `offset` (integer, optional): Pagination offset (default: 0)

**Response**:
- **200 OK**: Relationships found
  ```json
  {
    "entity_id": 123,
    "relationships": [
      {
        "id": 789,
        "type": "SENT",
        "from_type": "discovered_entity",
        "from_id": 123,
        "to_type": "email",
        "to_id": 456,
        "timestamp": "2001-05-15T14:30:00Z",
        "confidence_score": 1.0,
        "properties": {}
      },
      {
        "id": 790,
        "type": "COMMUNICATES_WITH",
        "from_type": "discovered_entity",
        "from_id": 123,
        "to_type": "discovered_entity",
        "to_id": 789,
        "timestamp": "2001-05-15T14:30:00Z",
        "confidence_score": 0.85,
        "properties": {
          "frequency": 12
        }
      }
    ],
    "total": 2,
    "limit": 100,
    "offset": 0
  }
  ```
- **404 Not Found**: Entity does not exist

---

### 4. Get Entity Neighbors (Traversal)

**Endpoint**: `GET /entities/:id/neighbors`

**Description**: Traverse relationships to find connected entities

**Path Parameters**:
- `id` (integer, required): Source entity ID

**Query Parameters**:
- `depth` (integer, optional): Traversal depth (default: 1, max: 5)
- `type` (string, optional): Filter by relationship type
- `entity_type` (string, optional): Filter neighbor entities by type
- `limit` (integer, optional): Maximum neighbors per level (default: 50, max: 100)

**Examples**:
- `/entities/123/neighbors?depth=1` - Immediate neighbors
- `/entities/123/neighbors?depth=3&type=COMMUNICATES_WITH` - 3-hop communication network

**Response**:
- **200 OK**: Neighbors found
  ```json
  {
    "source_entity_id": 123,
    "depth": 2,
    "neighbors": [
      {
        "id": 456,
        "unique_id": "ken.lay@enron.com",
        "type_category": "person",
        "name": "Ken Lay",
        "distance": 1,
        "relationship_path": [
          {
            "type": "COMMUNICATES_WITH",
            "from_id": 123,
            "to_id": 456
          }
        ]
      },
      {
        "id": 789,
        "unique_id": "andy.fastow@enron.com",
        "type_category": "person",
        "name": "Andy Fastow",
        "distance": 2,
        "relationship_path": [
          {
            "type": "COMMUNICATES_WITH",
            "from_id": 123,
            "to_id": 456
          },
          {
            "type": "COMMUNICATES_WITH",
            "from_id": 456,
            "to_id": 789
          }
        ]
      }
    ],
    "total": 2
  }
  ```
- **400 Bad Request**: Invalid depth or parameters
  ```json
  {
    "error": "invalid parameter",
    "details": "depth must be between 1 and 5"
  }
  ```
- **404 Not Found**: Source entity does not exist

---

### 5. Find Shortest Path

**Endpoint**: `POST /entities/path`

**Description**: Find the shortest path between two entities

**Request Body**:
```json
{
  "source_id": 123,
  "target_id": 789,
  "max_depth": 6
}
```

**Parameters**:
- `source_id` (integer, required): Starting entity ID
- `target_id` (integer, required): Target entity ID
- `max_depth` (integer, optional): Maximum search depth (default: 6, max: 10)

**Response**:
- **200 OK**: Path found
  ```json
  {
    "source_id": 123,
    "target_id": 789,
    "path_length": 3,
    "path": [
      {
        "entity_id": 123,
        "entity_name": "Jeff Skilling",
        "entity_type": "person"
      },
      {
        "relationship_id": 100,
        "relationship_type": "COMMUNICATES_WITH"
      },
      {
        "entity_id": 456,
        "entity_name": "Ken Lay",
        "entity_type": "person"
      },
      {
        "relationship_id": 101,
        "relationship_type": "COMMUNICATES_WITH"
      },
      {
        "entity_id": 789,
        "entity_name": "Andy Fastow",
        "entity_type": "person"
      }
    ]
  }
  ```
- **404 Not Found**: No path exists within max_depth
  ```json
  {
    "error": "no path found",
    "source_id": 123,
    "target_id": 789,
    "max_depth": 6
  }
  ```
- **400 Bad Request**: Invalid request body
  ```json
  {
    "error": "invalid request",
    "details": "source_id and target_id are required"
  }
  ```

---

### 6. Semantic Search

**Endpoint**: `POST /entities/search`

**Description**: Search for entities by semantic similarity using embeddings

**Request Body**:
```json
{
  "query": "energy trading executives",
  "limit": 10,
  "min_similarity": 0.7,
  "type_filter": "person"
}
```

**Parameters**:
- `query` (string, required): Natural language search query
- `limit` (integer, optional): Maximum results (default: 10, max: 100)
- `min_similarity` (float, optional): Minimum cosine similarity (default: 0.0, range: 0.0-1.0)
- `type_filter` (string, optional): Filter by entity type

**Response**:
- **200 OK**: Similar entities found
  ```json
  {
    "query": "energy trading executives",
    "results": [
      {
        "entity": {
          "id": 123,
          "unique_id": "jeff.skilling@enron.com",
          "type_category": "person",
          "name": "Jeff Skilling",
          "properties": {
            "title": "CEO"
          },
          "confidence_score": 0.95
        },
        "similarity": 0.89
      },
      {
        "entity": {
          "id": 456,
          "unique_id": "ken.lay@enron.com",
          "type_category": "person",
          "name": "Ken Lay",
          "properties": {
            "title": "Chairman"
          },
          "confidence_score": 0.92
        },
        "similarity": 0.85
      }
    ],
    "total": 2
  }
  ```
- **400 Bad Request**: Invalid request body
  ```json
  {
    "error": "invalid request",
    "details": "query is required"
  }
  ```

---

## Error Responses

All endpoints return consistent error structures:

```json
{
  "error": "error message",
  "details": "optional detailed explanation",
  "field": "optional field name for validation errors"
}
```

**HTTP Status Codes**:
- `200 OK`: Request successful
- `400 Bad Request`: Invalid input parameters
- `404 Not Found`: Resource does not exist
- `500 Internal Server Error`: Server error (with error tracking ID)

---

## Rate Limiting

Not implemented in POC. Future consideration for production.

---

## CORS

CORS headers enabled for all origins in development. Will be restricted in production.

---

## Authentication

Not implemented in POC. All endpoints are public.
