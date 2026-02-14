# Database Schema Design Checklist - Go (GORM) + PostgreSQL

Complete checklist for designing and reviewing database schemas with GORM and PostgreSQL.

---

## Pre-Design

- [ ] **Requirements Gathered**: Understand data entities and relationships
- [ ] **Access Patterns Identified**: Know how data will be queried
- [ ] **PostgreSQL Features Needed**: JSON/JSONB, arrays, full-text search, etc.
- [ ] **Scale Estimate**: Expected data volume and growth rate
- [ ] **Read/Write Ratio**: Understand if read-heavy or write-heavy
- [ ] **Connection Pooling Strategy**: pgbouncer or built-in pooling configuration

---

## GORM Model Design

### Model Structure

- [ ] **Embedded gorm.Model**: Using `gorm.Model` for ID, timestamps, soft delete
- [ ] **Custom Model Base**: Alternative base model if not using `gorm.Model`
- [ ] **Struct Tags Defined**: `gorm:"column:name;type:varchar(100)"`
- [ ] **JSON Tags**: `json:"field_name,omitempty"` for API responses

### Primary Keys

- [ ] **Primary Key Strategy**: Using `uint` with GORM auto-increment or UUID
- [ ] **UUID Implementation**: Using `github.com/google/uuid` if needed
- [ ] **Composite Keys**: `gorm:"primaryKey"` on multiple fields if needed

```go
type User struct {
    ID        uint           `gorm:"primaryKey"`
    // or
    ID        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid()"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}
```

### Data Types Mapping

- [ ] **String Types**: `string` → VARCHAR, TEXT
- [ ] **Numeric Types**: `int`, `int64`, `float64`, `decimal.Decimal`
- [ ] **Boolean**: `bool` → BOOLEAN
- [ ] **Time Types**: `time.Time` → TIMESTAMPTZ (with timezone)
- [ ] **JSON Types**: `datatypes.JSON` or custom types → JSONB
- [ ] **Arrays**: `pq.StringArray` or `postgres.Jsonb` → arrays
- [ ] **Enums**: PostgreSQL ENUMs or string with validation

```go
type Product struct {
    Price    decimal.Decimal `gorm:"type:decimal(10,2)"`
    Tags     pq.StringArray  `gorm:"type:text[]"`
    Metadata datatypes.JSON  `gorm:"type:jsonb"`
    Status   string          `gorm:"type:varchar(20);check:status IN ('active','inactive')"`
}
```

### Constraints & Validations

- [ ] **NOT NULL**: Using pointer types for nullable fields
- [ ] **Unique Constraints**: `gorm:"unique"` or `gorm:"uniqueIndex"`
- [ ] **Check Constraints**: `gorm:"check:price >= 0"`
- [ ] **Default Values**: `gorm:"default:value"` or PostgreSQL defaults
- [ ] **Field Validation**: Using validator tags `validate:"required,min=1"`

---

## GORM Relationships

### Association Tags

- [ ] **Foreign Key Tags**: `gorm:"foreignKey:UserID;references:ID"`
- [ ] **Constraint Tags**: `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
- [ ] **Join Table**: `gorm:"many2many:user_languages"`
- [ ] **Preloading Strategy**: Using `Preload` or `Joins` appropriately

### Relationship Types

- [ ] **Has One**: Using `has one` tag correctly
- [ ] **Has Many**: Using slice types for collections
- [ ] **Belongs To**: Foreign key field and object field defined
- [ ] **Many-to-Many**: Junction table with composite primary key
- [ ] **Polymorphic**: Using `polymorphic:Owner` tag
- [ ] **Self-Referencing**: Proper foreign key setup

```go
// One-to-Many
type User struct {
    ID      uint
    Orders  []Order `gorm:"foreignKey:UserID"`
}

type Order struct {
    ID     uint
    UserID uint
    User   User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}

// Many-to-Many
type User struct {
    ID        uint
    Languages []Language `gorm:"many2many:user_languages;"`
}

// Polymorphic
type Comment struct {
    ID            uint
    Body          string
    CommentableID uint
    CommentableType string
}
```

---

## PostgreSQL Indexing with GORM

### Index Creation

- [ ] **Index Tags**: Using `gorm:"index"` or `gorm:"uniqueIndex"`
- [ ] **Composite Indexes**: `gorm:"index:idx_name,priority:1"`
- [ ] **Partial Indexes**: Using raw SQL for conditional indexes
- [ ] **GIN/GiST Indexes**: For JSONB and full-text search
- [ ] **BRIN Indexes**: For large time-series tables

```go
type User struct {
    Email     string `gorm:"uniqueIndex"`
    Name      string `gorm:"index"`
    DeletedAt gorm.DeletedAt `gorm:"index"`

    // Composite index
    FirstName string `gorm:"index:idx_full_name,priority:1"`
    LastName  string `gorm:"index:idx_full_name,priority:2"`
}

// PostgreSQL specific indexes (in migration)
db.Exec("CREATE INDEX idx_users_email_lower ON users (LOWER(email))")
db.Exec("CREATE INDEX idx_products_metadata ON products USING GIN (metadata)")
```

### Index Strategy

- [ ] **Foreign Keys Indexed**: GORM auto-indexes foreign keys
- [ ] **WHERE Columns**: Frequently queried columns indexed
- [ ] **ORDER BY Columns**: Sort columns indexed
- [ ] **Pattern Matching**: Using `pg_trgm` for LIKE queries
- [ ] **Full-Text Search**: Using `tsvector` and GIN indexes

---

## GORM Performance Optimization

### Query Optimization

- [ ] **N+1 Prevention**: Using `Preload` or `Joins` for associations
- [ ] **Select Specific Fields**: Using `Select()` to limit columns
- [ ] **Batch Operations**: Using `CreateInBatches` for bulk inserts
- [ ] **Raw SQL**: Using `Raw()` for complex queries
- [ ] **Query Caching**: Implementing Redis or in-memory cache

```go
// Prevent N+1
db.Preload("Orders").Preload("Orders.Items").Find(&users)
db.Joins("Company").Joins("Manager").Find(&users)

// Select specific columns
db.Select("id", "name", "email").Find(&users)

// Batch insert
db.CreateInBatches(users, 100)

// Efficient counting
var count int64
db.Model(&User{}).Count(&count)
```

### PostgreSQL Specific

- [ ] **Connection Pool**: Configured `SetMaxIdleConns`, `SetMaxOpenConns`
- [ ] **Prepared Statements**: Using `PrepareStmt: true`
- [ ] **COPY Protocol**: For bulk data import
- [ ] **Partitioning**: Table partitioning for large tables
- [ ] **EXPLAIN ANALYZE**: Query plan analysis

```go
sqlDB, _ := db.DB()
sqlDB.SetMaxIdleConns(10)
sqlDB.SetMaxOpenConns(100)
sqlDB.SetConnMaxLifetime(time.Hour)
```

---

## GORM Migrations

### AutoMigrate

- [ ] **Safe Mode**: Understanding AutoMigrate limitations
- [ ] **Index Creation**: Indexes defined in struct tags
- [ ] **Foreign Keys**: Constraints properly created
- [ ] **No Column Deletion**: AutoMigrate won't drop columns

```go
// Basic auto-migration
db.AutoMigrate(&User{}, &Order{}, &Product{})

// With foreign keys
db.AutoMigrate(&User{}, &Order{})
db.Migrator().CreateConstraint(&Order{}, "User")
```

### Migration Tools

- [ ] **Migration Library**: Using golang-migrate or goose
- [ ] **Version Control**: Migration files in git
- [ ] **Rollback Scripts**: Down migrations written
- [ ] **Zero-Downtime**: Blue-green deployments considered

```sql
-- migrations/001_create_users.up.sql
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- migrations/001_create_users.down.sql
DROP TABLE IF EXISTS users;
```

---

## Security with GORM & PostgreSQL

### SQL Injection Prevention

- [ ] **GORM Methods**: Using GORM query builders, not raw strings
- [ ] **Parameterized Raw SQL**: Using `?` placeholders with `Raw()`
- [ ] **Input Validation**: Sanitizing user inputs
- [ ] **SQL Builder**: Using `Expr()` for dynamic queries safely

```go
// Safe - parameterized
db.Where("name = ?", userInput).Find(&users)
db.Raw("SELECT * FROM users WHERE email = ?", email).Scan(&result)

// Unsafe - avoid
db.Where(fmt.Sprintf("name = '%s'", userInput)).Find(&users)
```

### Data Protection

- [ ] **Password Hashing**: Using bcrypt or argon2
- [ ] **Field Encryption**: Sensitive fields encrypted at rest
- [ ] **PII Handling**: GDPR compliance, soft deletes
- [ ] **Audit Logging**: Track data access and modifications

```go
type User struct {
    ID              uint
    Email           string `gorm:"uniqueIndex"`
    PasswordHash    string `json:"-"` // Never expose in JSON
    SSN             string `gorm:"type:bytea"` // Encrypted
    DeletedAt       gorm.DeletedAt `gorm:"index"`
}
```

### PostgreSQL Security

- [ ] **Row Level Security**: RLS policies defined
- [ ] **SSL/TLS**: Encrypted connections enforced
- [ ] **Role-Based Access**: Different DB users per service
- [ ] **Connection String**: Credentials from environment variables

```go
dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=require",
    os.Getenv("DB_HOST"),
    os.Getenv("DB_USER"),
    os.Getenv("DB_PASSWORD"),
    os.Getenv("DB_NAME"),
    os.Getenv("DB_PORT"))
```

---

## Documentation & Testing

### Code Documentation

- [ ] **Model Comments**: Purpose and relationships documented
- [ ] **Tag Explanations**: Complex GORM tags explained
- [ ] **Query Examples**: Common queries documented
- [ ] **Performance Notes**: Slow query optimizations noted

### Schema Documentation

- [ ] **ERD Generated**: Using tools like dbdiagram.io
- [ ] **Migration README**: Explaining each migration
- [ ] **Data Dictionary**: Column purposes and constraints
- [ ] **API Documentation**: Model JSON representations

### Testing

- [ ] **Unit Tests**: Model methods tested
- [ ] **Integration Tests**: Database operations tested
- [ ] **Migration Tests**: Rollback/forward tested
- [ ] **Performance Tests**: Query benchmarks

```go
// Example test
func TestUserCreation(t *testing.T) {
    db := setupTestDB()
    user := User{Email: "test@example.com"}

    result := db.Create(&user)
    assert.NoError(t, result.Error)
    assert.NotZero(t, user.ID)

    var found User
    db.First(&found, user.ID)
    assert.Equal(t, user.Email, found.Email)
}
```

---

## PostgreSQL Specific Features

### Advanced Data Types

- [ ] **UUID**: Using `pgcrypto` extension
- [ ] **JSONB**: For semi-structured data
- [ ] **Arrays**: For list-like data
- [ ] **Range Types**: For time periods, numeric ranges
- [ ] **PostGIS**: For geographical data

### Extensions

- [ ] **pg_trgm**: Trigram matching for fuzzy search
- [ ] **pgcrypto**: UUID generation and encryption
- [ ] **postgres_fdw**: Foreign data wrappers
- [ ] **pg_stat_statements**: Query performance monitoring

### Performance Features

- [ ] **Materialized Views**: For complex aggregations
- [ ] **Table Partitioning**: For large tables
- [ ] **Parallel Queries**: For large data scans
- [ ] **VACUUM Strategy**: Auto-vacuum configured

```go
// Enable extensions
db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")
db.Exec("CREATE EXTENSION IF NOT EXISTS \"pg_trgm\"")

// Use PostgreSQL features
type Location struct {
    ID       uint
    Point    string         `gorm:"type:geography(POINT, 4326)"`
    Metadata datatypes.JSON `gorm:"type:jsonb"`
    Tags     pq.StringArray `gorm:"type:text[]"`
}
```
