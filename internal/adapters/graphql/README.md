# Rootly GraphQL Schemas

This directory contains the GraphQL schemas for all Rootly microservices.

## Architecture Overview

Each microservice has its own consolidated GraphQL schema file that follows a standardized structure and naming convention.

## File Structure

### Service Schema Files
- **Pattern**: `{service-name}.graphqls`
- **Purpose**: Consolidated GraphQL schema for each specific service
- **Content**: All types, inputs, queries, mutations, and subscriptions for that service
- **Language**: English comments
- **Scope**: Only implements what exists in the corresponding backend service

### Configuration Files
- **`gqlgen.yml`**: gqlgen configuration for code generation
- **`schema.graphqls`**: Main schema file (if needed for shared types)
- **`generated/`**: Auto-generated Go code from schemas

## Current Services

### Analytics Service (`metrics.graphqls`)
- **Backend**: `rootly-analytics-backend`
- **Queries**: 5 implemented queries
- **Mutations**: None (read-only service)
- **Subscriptions**: None (no real-time features)

## Schema Structure Standard

All service schema files should follow this organization:

```graphql
# ========================================
# SCALARS
# Shared scalar types for the service
# ========================================

# ========================================  
# ENUMS
# Enumeration types specific to the service
# ========================================

# ========================================
# INPUT TYPES
# Input types for queries and mutations
# ========================================

# ========================================
# DOMAIN TYPES
# Core domain entities for the service
# ========================================

# ========================================
# QUERIES
# Available GraphQL queries
# ========================================

# ========================================
# MUTATIONS
# Write operations (if implemented)
# ========================================

# ========================================
# SUBSCRIPTIONS
# Real-time subscriptions (if implemented)
# ========================================

# ========================================
# SCHEMA DEFINITION
# Schema configuration
# ========================================
```

## Adding New Services

To add a new microservice schema:

1. **Create schema file**: `{service-name}.graphqls`
2. **Follow structure**: Use the standard section organization
3. **Update gqlgen.yml**: Add the new schema file to the configuration
4. **Map domain types**: Add type mappings for Go domain structures
5. **Generate code**: Run `go run github.com/99designs/gqlgen generate`

### Example for new service:

```yaml
# In gqlgen.yml
schema:
  - schema.graphqls
  - metrics.graphqls
  - user-management.graphqls  # New service
  - notifications.graphqls    # Another service
```

## Implementation Guidelines

### Backend Integration
- Only implement GraphQL operations that exist in the backend service
- Map GraphQL types directly to existing Go domain types
- Use consistent naming conventions across services

### Type Mapping
- Scalars: Use shared types (`DateTime`, `UUID`, `JSON`)
- Domain types: Map to `/internal/domain/{Service}*` structures
- Input types: Map to service-specific request structures

### Documentation
- Use English comments for all types and fields
- Include backend endpoint references in query comments
- Mark unimplemented sections clearly

## Code Generation

To generate GraphQL code for all services:
```bash
cd /path/to/rootly-apigateway
go run github.com/99designs/gqlgen generate
```

## Domain Mapping Pattern

All GraphQL types should map to existing Go domain types:
```yaml
models:
  # Service-specific types
  ServiceEntityName:
    model:
      - github.com/swarch-2f-rootly/rootly-apigateway/internal/domain.ServiceEntityName
  
  # Input types  
  ServiceInputName:
    model:
      - github.com/swarch-2f-rootly/rootly-apigateway/internal/domain.ServiceRequestType
```

## Best Practices

1. **One schema per service**: Keep service boundaries clear
2. **Implement only existing functionality**: Don't define unimplemented features
3. **Consistent structure**: Follow the standard section organization
4. **English documentation**: Use clear, English comments
5. **Direct mapping**: Map GraphQL types 1:1 with backend domain types
6. **Version control**: Track schema changes alongside backend changes

This structure ensures maintainability, scalability, and consistency across all Rootly microservices.