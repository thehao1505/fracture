# LLM Proxy Microservice - Technical Knowledge Base

## Technology Stack

**Runtime Environment:** Node.js 22 (Alpine Linux containers)  
**Primary Language:** TypeScript  
**Framework:** NestJS 11.x (Microservices architecture)  
**Package Manager:** pnpm with workspace support  
**Build System:** Nest CLI with Webpack compilation

**Key Libraries:**

- **Web Framework:** Express 5.x with NestJS platform
- **Validation & Transformation:** class-validator, class-transformer, Joi, Zod
- **Database ORMs:**
  - MongoDB: Mongoose 8.x with @nestjs/mongoose
  - PostgreSQL: TypeORM 0.3.x with @nestjs/typeorm
- **Authentication:** JWT with passport-jwt, passport-local
- **Logging:** Pino with nestjs-pino, pino-roll for file rotation
- **Caching:** cache-manager with Redis via @keyv/redis
- **HTTP Client:** Axios with @nestjs/axios
- **Configuration:** @nestjs/config with environment-based setup
- **Event Handling:** @nestjs/event-emitter for async events
- **Scheduling:** @nestjs/schedule for cron jobs

**Development Tools:**

- **Linting:** ESLint 9.x with TypeScript parser and Prettier integration
- **Testing:** Jest with ts-jest for TypeScript support
- **Git Hooks:** Husky with lint-staged for pre-commit validation
- **Containerization:** Docker with multi-stage builds

## Overall Architecture

**Pattern:** NestJS Monorepo with Microservices  
**Structure:** Multi-application workspace managed by Nest CLI

### Top-Level Directory Structure

- `apps/` - Contains 6 independent microservices, each with dedicated Dockerfile
- `libs/` - Shared libraries and common utilities
- `litellm/` - LiteLLM configuration for model routing and management
- `nginx/` - Reverse proxy configuration with SSL certificates
- `logs/` - Centralized logging directory with per-service subdirectories

### Microservices Architecture

The system implements a distributed microservices pattern with:

- Independent deployment via Docker containers
- Shared database connections (PostgreSQL + MongoDB)
- Inter-service communication through HTTP APIs
- Centralized logging and configuration management
- Redis-based caching layer

## Shared Code Conventions

**Location:** `libs/common/src/`

### Core Shared Components

**Database Abstractions:**

- `mongodb/` - Multi-tenant MongoDB connection management with tenant-specific database isolation
- `relational-database/` - PostgreSQL TypeORM configuration with connection pooling
- `entities/` - TypeORM entity definitions for relational data
- `schemas/` - Mongoose schema definitions for document data
- `repositories/` - Repository pattern implementations

**Cross-Cutting Concerns:**

- `logger/` - Pino-based logging with file rotation and structured output
- `cache/` - Redis caching utilities and decorators
- `health-check/` - Health check endpoints and monitoring
- `jwt-strategy/` - JWT authentication strategies and guards
- `exception-filter/` - Global exception handling and error formatting

**Data Transfer Objects:**

- `dtos/` - Request/response DTOs with validation decorators
- `interfaces/` - TypeScript interface definitions
- `constants/` - Application-wide constants and enums

**Utilities:**

- `decorators/` - Custom NestJS decorators
- `utils/` - Helper functions and utility classes
- `file-service/` - File handling and storage utilities
- `cryto/` - Cryptographic utilities and encryption helpers
- **TCP**: Internal microservice communication using NestJS microservices
- **Event-driven**: Event emitters for loose coupling between services

## Shared Code Architecture (`libs/common`)

The `common` library provides foundational abstractions and utilities shared across all microservices:

### Database Abstractions

- **`relational-database/`**: TypeORM module with abstract repository pattern
  - `AbstractEntity`: Base entity with id, timestamps, and soft delete
  - `AbstractRepository`: Generic repository with CRUD operations
  - `RelationalDatabaseModule`: PostgreSQL configuration module

- **`mongodb/`**: Mongoose module with abstract repository pattern
  - `MongoAbstractDocument`: Base document with ObjectId
  - `MongoAbstractRepository`: Generic MongoDB repository
  - `MongoDatabaseModule`: MongoDB configuration module

### Core Domain Entities

Located in `entities/`, using TypeORM decorators:

- `Team`: Organization/tenant entity with settings and LLM associations
- `User`: User entity with team relationships and usage tracking
- `Role`: Role-based access control
- `LLM`: Language model configurations
- `LLMConnection`: External LLM provider connection details
- `RateLimit`: Usage limits per team/user/model
- `DailyUsageTracking`: Daily usage statistics
- `ContinueConfiguration`: IDE-specific configurations

### Cross-Cutting Concerns

#### Authentication & Authorization

- **`jwt-strategy/`**: JWT validation strategies
  - `JwtUserVerifierGuard`: User-level authentication
  - `JwtAdminVerifierGuard`: Admin-level authentication
  - Asymmetric and symmetric key verification

#### Logging

- **`logger/`**: Pino-based structured logging
  - Auto-generates request IDs
  - Configurable log levels (development vs production)
  - File and console output streams
  - Sensitive data redaction (authorization headers, API keys)

#### Validation & DTOs

- **`dtos/`**: Data Transfer Objects with class-validator decorators
- **`decorators/`**: Custom decorators for metadata and validation
- **`interfaces/`**: TypeScript interfaces for type safety

#### Caching & Configuration

- **`cache/`**: Redis-based caching module
- **`constants/`**: Application-wide constants and enums
- **`utils/`**: Utility functions and helpers

#### Exception Handling

- **`exception-filter/`**: Global exception filters for consistent error responses

## Application-Specific Designs

### model-gateway (Main API Gateway)

**Design Pattern**: Layered Architecture with Provider Pattern

**Structure**:

```
src/
├── controllers/           # HTTP request handling
├── guards/               # Authentication and usage control
├── interceptors/         # Cross-cutting concerns (logging, reporting)
├── providers/            # Business logic for different LLM types
├── services/            # External service integration
└── utils/               # Utility functions
```

**Request Pipeline**:

1. `JwtUserVerifierGuard` - Authentication
2. `UsageControlGuard` - Rate limiting and access control
3. Controller method - Route handling
4. Provider - Business logic delegation
5. `ReportInterceptor` - Usage logging and analytics

**Key Providers**:

- `LlmProvider` - Chat and completion requests
- `FimProvider` - Fill-in-the-middle code completion
- `EmbeddingsProvider` - Text embedding generation
- `RerankProvider` - Document reranking

### configuration-service (System Management)

**Design Pattern**: Modular Architecture with Event-Driven Components

**Structure**:

```
src/
├── modules/             # Feature-specific modules (LLM, User, Team management)
├── guards/              # Custom authentication guards
├── strategies/          # Passport authentication strategies
├── event-listener/      # Event-driven update handlers
├── subscribers/         # Database event subscribers
└── decorators/          # Custom metadata decorators
```

**Authentication Strategy**: Multi-strategy approach with `AuthenticationGuard` supporting both JWT and local authentication.

### auth-service (Authentication Hub)

**Design Pattern**: OAuth2 Adapter Pattern

**Key Components**:

- OAuth2 flow management with external identity providers
- JWT token generation and validation
- Integration with HRM (Human Resource Management) systems
- Team and user synchronization

### continue-proxy (IDE Integration)

**Design Pattern**: Proxy Pattern with View Entities

**Specialization**: Uses read-only database views (`*View` entities) for optimized data access without modification capabilities.

### logger-service & analytic-service

**Design Pattern**: Repository Pattern with MongoDB

**Logger Service - Multi-Tenant Architecture:**

- **Database Strategy**: Per-tenant MongoDB databases using dynamic connection provider
- **Pattern**: `${TEAM_DB_PREFIX}_${tenantId}` database naming convention
- **Implementation**: Uses `tenantConnectionProvider` for tenant-aware database connections
- **Data Isolation**: Complete tenant isolation at the database level
- **Connection Management**: Dynamic connection switching based on `TEAM_ID_HEADER` or request data
- **Repository**: `ModelLoggerRepository` implements tenant-aware data access
- **Schema**: `ModelMessageDocument` for storing LLM interaction logs per tenant

**Analytic Service - Single Database Architecture:**

- **Database Strategy**: Single shared MongoDB database with application-level tenant filtering
- **Pattern**: Traditional CRUD service with `teamId` field for data segregation
- **Implementation**: Uses standard MongoDB connection without tenant isolation
- **Data Access**: `AnalyticRepository` filters data by `teamId` at query level
- **Schema**: `AnalyticDocument` with `teamId` field for tenant identification
- **Performance**: Optimized for analytics queries across tenant data

**Key Differences:**

- Logger Service: Complete database isolation per tenant for security and compliance
- Analytic Service: Shared database with logical separation for cross-tenant analytics capabilities

### usage-control

**Design Pattern**: Guard Pattern with Rate Limiting

**Purpose**: Centralized usage validation, rate limiting, and resource access control using database views for performance.

## Data Integration Patterns

### External Dependencies

The system integrates with multiple external services:

- **LLM Providers**: OpenAI, Azure OpenAI, self-hosted vLLM instances
- **Identity Providers**: OAuth2-compatible systems
- **HRM Systems**: Employee and team data synchronization
- **LiteLLM**: Unified LLM API abstraction layer

### Database Design

- **PostgreSQL**: Relational data with referential integrity
  - Teams, Users, Roles, LLMs, Connections, Rate Limits
  - Complex relationships with cascade deletes
  - JSONB columns for flexible configuration storage

- **MongoDB**: Document storage for time-series and analytics
  - Model interaction logs
  - Analytics events
  - Hybrid multi-tenant approach (see Multi-Tenant Strategy below)

### Multi-Tenant Data Strategy

The system implements different multi-tenancy patterns based on service requirements:

**1. Database-per-Tenant (Logger Service):**

- **Rationale**: Complete data isolation for compliance and security
- **Implementation**: Each tenant gets a dedicated MongoDB database
- **Naming Convention**: `${TEAM_DB_PREFIX}_${tenantId}` (e.g., `team_db_abc123`)
- **Connection Management**: Dynamic database switching via `tenantConnectionProvider`
- **Benefits**: Complete isolation, independent scaling, regulatory compliance
- **Use Case**: Sensitive log data requiring strict isolation

**2. Shared Database with Tenant ID (Most Services):**

- **Rationale**: Efficient resource utilization and cross-tenant analytics
- **Implementation**: Single database with `teamId` field for data segregation
- **Filtering**: Application-level queries filter by `teamId`
- **Benefits**: Resource efficiency, easier maintenance, cross-tenant reporting
- **Use Cases**: Configuration data, analytics, general business data

**3. Hybrid Approach (System-wide):**

- PostgreSQL entities use shared database with `teamId` filtering
- MongoDB logger service uses database-per-tenant for isolation
- Analytics service uses shared MongoDB for cross-tenant insights

### Caching Strategy

- **Redis**: Performance optimization for frequently accessed data
- **Application-level**: Configuration and authentication data caching

## Development Workflow

### Installation & Setup

```bash
# Install dependencies
pnpm install

# Development mode (default: model-gateway)
pnpm run start:dev

# Production mode
pnpm run start:prod
```

### Code Quality & Testing

```bash
# Type checking
pnpm run type-check

# Linting with auto-fix
pnpm run lint

# Code formatting
pnpm run format

# Testing
pnpm run test          # Unit tests
pnpm run test:e2e      # End-to-end tests
pnpm run test:cov      # Coverage report
```

### Build & Deployment

```bash
# Build all applications
pnpm run build

# Pre-commit hooks (automated)
pnpm run precommit     # Runs lint-staged
```

### Service-Specific Commands

Each service can be started individually in development mode:

```bash
pnpm run start:dev analytic-service
pnpm run start:dev auth-service
pnpm run start:dev configuration-service
# ... etc
```

### Docker Infrastructure

The system includes comprehensive Docker configuration:

- Multi-stage Dockerfiles for each service
- Docker Compose for local development
- Production-ready container configurations
- Nginx reverse proxy with SSL termination

### Configuration Management

- Environment-based configuration with Joi validation
- Centralized configuration service for dynamic settings
- LiteLLM integration for LLM provider abstraction
- Kubernetes-ready configuration structure

### Monitoring & Observability

- Structured logging with request correlation IDs
- Health check endpoints across all services
- Usage analytics and reporting
- Error tracking and performance monitoring

## Key Design Principles

1. **Separation of Concerns**: Clear boundaries between authentication, business logic, and data access
2. **Repository Pattern**: Consistent data access abstraction across SQL and NoSQL databases
3. **Event-Driven Architecture**: Loose coupling through event emitters and listeners
4. **Guard Pattern**: Centralized security and access control
5. **Provider Pattern**: Pluggable business logic for different LLM types
6. **View Pattern**: Read-optimized database views for performance-critical operations
7. **Configuration as Code**: Externalized configuration with validation
8. **Observability**: Comprehensive logging, monitoring, and analytics integration
