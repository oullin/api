---
name: database-schema-designer
description: Design robust, scalable database schemas for SQL and NoSQL databases. Provides normalization guidelines, indexing strategies, migration patterns, constraint design, and performance optimization. Ensures data integrity, query performance, and maintainable data models.
license: MIT
---

# Database Schema Designer

Design production-ready database schemas with best practices built-in.

---

## Quick Start

Just describe your data model:

```
design a schema for an e-commerce platform with users, products, orders
```

You'll get complete GORM models and PostgreSQL schema:

```go
// User model
type User struct {
    ID        uint           `gorm:"primaryKey"`
    Email     string         `gorm:"uniqueIndex;not null;size:255"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`

    // Relationships
    Orders []Order `gorm:"foreignKey:UserID"`
}

// Order model
type Order struct {
    ID        uint           `gorm:"primaryKey"`
    UserID    uint           `gorm:"index;not null"`
    User      User           `gorm:"constraint:OnDelete:CASCADE"`
    Total     decimal.Decimal `gorm:"type:decimal(10,2);not null"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

// Migration
db.AutoMigrate(&User{}, &Order{})
```

**What to include in your request:**
- Entities (users, products, orders)
- Key relationships (users have orders, orders have items)
- Scale hints (high-traffic, millions of records)
- GORM-specific features needed (soft deletes, hooks, custom types)

---

## Triggers

| Trigger             | Example                                   |
|---------------------|-------------------------------------------|
| `design schema`     | "design a schema for user authentication" |
| `database design`   | "database design for multi-tenant SaaS"   |
| `create tables`     | "create tables for a blog system"         |
| `schema for`        | "schema for inventory management"         |
| `model data`        | "model data for real-time analytics"      |
| `I need a database` | "I need a database for tracking orders"   |
| `design NoSQL`      | "design NoSQL schema for product catalog" |

---

## Key Terms

| Term                 | Definition                                                               |
|----------------------|--------------------------------------------------------------------------|
| **GORM**             | Go ORM library that simplifies database operations and migrations        |
| **Struct Tags**      | Go tags that define GORM behavior (`gorm:"primaryKey"`, etc.)           |
| **AutoMigrate**      | GORM feature that automatically creates/updates tables from structs      |
| **Soft Delete**      | Marks records as deleted without removing (uses `DeletedAt` field)       |
| **Preloading**       | Loading related data with `Preload()` to avoid N+1 queries               |
| **Hooks**            | GORM callbacks (BeforeCreate, AfterUpdate, etc.) for custom logic        |
| **Foreign Key (FK)** | Relationship constraint defined with `gorm:"foreignKey:FieldName"`       |
| **Index**            | Defined with tags like `gorm:"index"` or `gorm:"uniqueIndex"`            |
| **Access Pattern**   | How your app reads/writes data (queries, joins, filters)                 |

---

## Quick Reference

| Task         | GORM Approach          | PostgreSQL Consideration     |
|--------------|------------------------|------------------------------|
| New model    | Define struct with tags| Use `gorm.Model` for basics  |
| Primary keys | `gorm:"primaryKey"`    | Auto-incrementing BIGSERIAL  |
| Foreign keys | `gorm:"foreignKey"`    | Add `constraint:OnDelete`    |
| Indexes      | `gorm:"index"`         | Composite: `gorm:"index:idx_name"` |
| Unique       | `gorm:"uniqueIndex"`   | Can be composite too         |
| Migrations   | `db.AutoMigrate()`     | Use golang-migrate for production |
| Soft delete  | Include `gorm.DeletedAt` | Automatic WHERE clauses    |

---

## Process Overview

```
Your Data Requirements
    |
    v
+-----------------------------------------------------+
| Phase 1: ANALYSIS                                   |
| * Identify entities and relationships               |
| * Determine access patterns (read vs write heavy)   |
| * Choose SQL or NoSQL based on requirements         |
+-----------------------------------------------------+
    |
    v
+-----------------------------------------------------+
| Phase 2: DESIGN                                     |
| * Normalize to 3NF (SQL) or embed/reference (NoSQL) |
| * Define primary keys and foreign keys              |
| * Choose appropriate data types                     |
| * Add constraints (UNIQUE, CHECK, NOT NULL)         |
+-----------------------------------------------------+
    |
    v
+-----------------------------------------------------+
| Phase 3: OPTIMIZE                                   |
| * Plan indexing strategy                            |
| * Consider denormalization for read-heavy queries   |
| * Add timestamps (created_at, updated_at)           |
+-----------------------------------------------------+
    |
    v
+-----------------------------------------------------+
| Phase 4: MIGRATE                                    |
| * Generate migration scripts (up + down)            |
| * Ensure backward compatibility                     |
| * Plan zero-downtime deployment                     |
+-----------------------------------------------------+
    |
    v
Production-Ready Schema
```

---

## Commands

| Command                      | When to Use           | Action                      |
|------------------------------|-----------------------|-----------------------------|
| `design schema for {domain}` | Starting fresh        | Full schema generation      |
| `normalize {table}`          | Fixing existing table | Apply normalization rules   |
| `add indexes for {table}`    | Performance issues    | Generate index strategy     |
| `migration for {change}`     | Schema evolution      | Create reversible migration |
| `review schema`              | Code review           | Audit existing schema       |

**Workflow:** Start with `design schema` → iterate with `normalize` → optimize with `add indexes` → evolve with `migration`

---

## Core Principles

| Principle                   | WHY                         | Implementation                         |
|-----------------------------|-----------------------------|----------------------------------------|
| Model the Domain            | UI changes, domain doesn't  | Entity names reflect business concepts |
| Data Integrity First        | Corruption is costly to fix | Constraints at database level          |
| Optimize for Access Pattern | Can't optimize for both     | OLTP: normalized, OLAP: denormalized   |
| Plan for Scale              | Retrofitting is painful     | Index strategy + partitioning plan     |

---

## Anti-Patterns

| Avoid                           | Why                          | GORM/PostgreSQL Solution               |
|---------------------------------|------------------------------|----------------------------------------|
| Using `interface{}` for types   | Loses type safety            | Define proper Go types                 |
| Missing `gorm.Model`            | No timestamps/soft delete    | Embed `gorm.Model` or add fields       |
| `float64` for money             | Rounding errors              | Use `decimal.Decimal` type             |
| No struct tags                  | GORM guesses wrong           | Explicit tags for all constraints      |
| Missing FK indexes              | Slow JOINs                   | Add `gorm:"index"` to FK fields        |
| Raw SQL everywhere              | Loses GORM benefits          | Use GORM query builder                 |
| `SELECT *` with GORM            | Fetches unnecessary data     | Use `Select()` for specific columns    |
| No connection pooling           | Connection exhaustion        | Configure `SetMaxOpenConns()`          |
| Ignoring GORM errors            | Silent failures              | Always check `if err != nil`           |

---

## Verification Checklist

After designing a schema:

- [ ] Every table has a primary key
- [ ] All relationships have foreign key constraints
- [ ] ON DELETE strategy defined for each FK
- [ ] Indexes exist on all foreign keys
- [ ] Indexes exist on frequently queried columns
- [ ] Appropriate data types (DECIMAL for money, etc.)
- [ ] NOT NULL on required fields
- [ ] UNIQUE constraints where needed
- [ ] CHECK constraints for validation
- [ ] created_at and updated_at timestamps
- [ ] Migration scripts are reversible
- [ ] Tested on staging with production data

---

<details>
<summary><strong>Deep Dive: GORM Models & Normalization</strong></summary>

### Normal Forms in GORM

| Form    | Rule                               | GORM Implementation              |
|---------|------------------------------------|----------------------------------|
| **1NF** | Atomic values, no arrays in columns | Use separate models with relationships |
| **2NF** | No partial dependencies            | Proper struct composition        |
| **3NF** | No transitive dependencies         | Reference tables via foreign keys |

### First Normal Form (1NF)

```go
// BAD: Multiple values in single field
type Order struct {
    ID         uint
    ProductIDs string // "101,102,103" - violates 1NF
}

// GOOD: Separate model for items
type Order struct {
    gorm.Model
    CustomerID uint
    Items      []OrderItem `gorm:"foreignKey:OrderID"`
}

type OrderItem struct {
    gorm.Model
    OrderID   uint
    ProductID uint
    Quantity  int
    Price     decimal.Decimal `gorm:"type:decimal(10,2)"`
}
```

### Relationships in GORM

```go
// One-to-Many
type User struct {
    gorm.Model
    Email  string  `gorm:"uniqueIndex;not null"`
    Orders []Order `gorm:"foreignKey:UserID"`
}

// Many-to-Many
type Product struct {
    gorm.Model
    Name       string `gorm:"not null"`
    Categories []Category `gorm:"many2many:product_categories;"`
}

// Self-referencing
type Employee struct {
    gorm.Model
    Name      string
    ManagerID *uint
    Manager   *Employee  `gorm:"foreignKey:ManagerID"`
    Team      []Employee `gorm:"foreignKey:ManagerID"`
}
```

### When to Denormalize

```go
// Denormalized for performance
type Order struct {
    gorm.Model
    UserID     uint

    // Denormalized fields (cached)
    TotalAmount decimal.Decimal `gorm:"type:decimal(10,2)"`
    ItemCount   int
    UserEmail   string `gorm:"type:varchar(255)"` // Cached from User
}

// Use hooks to maintain denormalized data
func (o *Order) BeforeSave(tx *gorm.DB) error {
    // Calculate total from items
    var total decimal.Decimal
    tx.Model(&OrderItem{}).Where("order_id = ?", o.ID).
        Select("SUM(price * quantity)").Scan(&total)
    o.TotalAmount = total
    return nil
}
```

</details>

<details>
<summary><strong>Deep Dive: PostgreSQL Data Types in GORM</strong></summary>

### String Types

```go
type Product struct {
    // PostgreSQL VARCHAR
    Name        string `gorm:"type:varchar(100);not null"`

    // PostgreSQL TEXT
    Description string `gorm:"type:text"`

    // Fixed length CHAR
    CountryCode string `gorm:"type:char(2)"`

    // With size constraint
    Email       string `gorm:"size:255;uniqueIndex"`
}
```

### Numeric Types

```go
import "github.com/shopspring/decimal"

type FinancialRecord struct {
    // Integer types
    ID          uint    `gorm:"primaryKey"`           // BIGSERIAL
    Quantity    int32   `gorm:"type:integer"`         // INTEGER
    BigNumber   int64   `gorm:"type:bigint"`          // BIGINT
    SmallCount  int16   `gorm:"type:smallint"`        // SMALLINT

    // Money - ALWAYS use decimal
    Price       decimal.Decimal `gorm:"type:decimal(10,2);not null"`

    // Float - only for scientific data
    Temperature float64 `gorm:"type:real"`            // REAL
    Precision   float64 `gorm:"type:double precision"` // DOUBLE PRECISION
}
```

### Date/Time Types

```go
import (
    "time"
    "database/sql/driver"
)

type Event struct {
    gorm.Model // Includes CreatedAt, UpdatedAt, DeletedAt

    // PostgreSQL DATE
    EventDate   time.Time `gorm:"type:date"`

    // PostgreSQL TIME
    StartTime   time.Time `gorm:"type:time"`

    // PostgreSQL TIMESTAMP WITH TIME ZONE
    ScheduledAt time.Time `gorm:"type:timestamptz"`

    // Custom nullable time
    CompletedAt *time.Time `gorm:"type:timestamptz"`
}
```

### PostgreSQL-Specific Types

```go
import (
    "github.com/lib/pq"
    "github.com/google/uuid"
)

type AdvancedModel struct {
    // UUID
    ID       uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

    // Arrays
    Tags     pq.StringArray `gorm:"type:text[]"`
    Numbers  pq.Int64Array  `gorm:"type:integer[]"`

    // JSON/JSONB
    Metadata json.RawMessage `gorm:"type:jsonb"`

    // Boolean
    IsActive bool `gorm:"type:boolean;default:true"`

    // ENUM (requires type creation)
    Status   string `gorm:"type:order_status;default:'pending'"`
}
```

### Custom Types

```go
// Custom type for PostgreSQL ENUM
type OrderStatus string

const (
    StatusPending   OrderStatus = "pending"
    StatusProcessing OrderStatus = "processing"
    StatusCompleted OrderStatus = "completed"
)

func (s *OrderStatus) Scan(value interface{}) error {
    *s = OrderStatus(value.(string))
    return nil
}

func (s OrderStatus) Value() (driver.Value, error) {
    return string(s), nil
}
```

</details>

<details>
<summary><strong>Deep Dive: Indexing in GORM & PostgreSQL</strong></summary>

### Defining Indexes in GORM

```go
type User struct {
    gorm.Model

    // Single column index
    Email    string `gorm:"index;not null"`

    // Unique index
    Username string `gorm:"uniqueIndex"`

    // Composite index
    FirstName string `gorm:"index:idx_name"`
    LastName  string `gorm:"index:idx_name"`

    // Index with options
    Phone    string `gorm:"index:,unique,sort:desc"`

    // Foreign key (auto-indexed by GORM)
    TeamID   uint `gorm:"index"`
}

// Custom index creation
db.Exec("CREATE INDEX CONCURRENTLY idx_users_email_lower ON users(LOWER(email))")
```

### PostgreSQL Index Types

```go
// B-Tree (default) - ranges, equality
type Product struct {
    Price decimal.Decimal `gorm:"index;type:decimal(10,2)"`
}

// GIN index for JSONB
type Document struct {
    Data json.RawMessage `gorm:"type:jsonb;index:,type:gin"`
}

// GiST for full-text search
db.Exec("CREATE INDEX idx_search ON products USING GiST(to_tsvector('english', name || ' ' || description))")

// Partial index
db.Exec("CREATE INDEX idx_active_users ON users(email) WHERE deleted_at IS NULL")
```

### Query Optimization with GORM

```go
// Use index hints
var users []User

// Efficient: uses index
db.Where("email = ?", "user@example.com").Find(&users)

// Efficient: composite index usage
db.Where("first_name = ? AND last_name = ?", "John", "Doe").Find(&users)

// Inefficient: full table scan
db.Where("LOWER(email) = ?", "user@example.com").Find(&users)

// Better: create functional index
db.Exec("CREATE INDEX idx_email_lower ON users(LOWER(email))")
```

### Analyzing Query Performance

```go
// Enable query logging
db = db.Debug()

// Explain query plan
var result []map[string]interface{}
db.Raw("EXPLAIN ANALYZE SELECT * FROM users WHERE email = ?", "test@example.com").Scan(&result)

// Check index usage
db.Raw(`
    SELECT schemaname, tablename, indexname, idx_scan
    FROM pg_stat_user_indexes
    WHERE tablename = 'users'
`).Scan(&result)
```

### Index Best Practices for GORM

| Practice                | Implementation                                    |
|-------------------------|---------------------------------------------------|
| Index foreign keys      | Add `gorm:"index"` to FK fields                  |
| Composite for queries   | Match WHERE clause column order                  |
| Unique business keys    | Use `gorm:"uniqueIndex"`                         |
| Soft delete queries     | PostgreSQL partial index on `deleted_at IS NULL` |
| Case-insensitive search | Functional index on `LOWER(column)`              |

</details>

<details>
<summary><strong>Deep Dive: Constraints in GORM</strong></summary>

### Primary Keys

```go
// Auto-increment (default)
type User struct {
    ID uint `gorm:"primaryKey"` // BIGSERIAL in PostgreSQL
}

// UUID primary key
import "github.com/google/uuid"

type Document struct {
    ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
}

// Composite primary key
type Enrollment struct {
    StudentID uint `gorm:"primaryKey"`
    CourseID  uint `gorm:"primaryKey"`
    Grade     string
}
```

### Foreign Key Constraints

```go
// Basic foreign key
type Order struct {
    gorm.Model
    UserID uint `gorm:"not null"`
    User   User `gorm:"foreignKey:UserID"`
}

// With ON DELETE behavior
type OrderItem struct {
    gorm.Model
    OrderID uint  `gorm:"not null"`
    Order   Order `gorm:"constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
}

// Optional relationship (nullable FK)
type Post struct {
    gorm.Model
    AuthorID *uint   // Nullable
    Author   *User   `gorm:"constraint:OnDelete:SET NULL"`
}
```

### Unique Constraints

```go
type Account struct {
    gorm.Model

    // Single column unique
    Email string `gorm:"uniqueIndex;not null"`

    // Composite unique
    Domain   string `gorm:"uniqueIndex:idx_domain_username"`
    Username string `gorm:"uniqueIndex:idx_domain_username"`
}
```

### Check Constraints (PostgreSQL)

```go
// Using table comment for check constraints
type Product struct {
    gorm.Model
    Price    decimal.Decimal `gorm:"type:decimal(10,2)"`
    Discount int
}

// Add check constraints via migration
db.Exec(`
    ALTER TABLE products
    ADD CONSTRAINT check_price CHECK (price >= 0),
    ADD CONSTRAINT check_discount CHECK (discount BETWEEN 0 AND 100)
`)
```

### Custom Validation with Hooks

```go
type User struct {
    gorm.Model
    Age   int
    Email string
}

// BeforeCreate hook for validation
func (u *User) BeforeCreate(tx *gorm.DB) error {
    if u.Age < 0 || u.Age > 150 {
        return errors.New("invalid age")
    }
    if !strings.Contains(u.Email, "@") {
        return errors.New("invalid email format")
    }
    return nil
}
```

</details>

<details>
<summary><strong>Deep Dive: GORM Relationship Patterns</strong></summary>

### One-to-Many

```go
// Parent model
type Customer struct {
    gorm.Model
    Name   string
    Orders []Order `gorm:"foreignKey:CustomerID"`
}

// Child model
type Order struct {
    gorm.Model
    CustomerID uint
    Customer   Customer `gorm:"constraint:OnDelete:CASCADE"`
    Items      []OrderItem
}

// Usage with preloading
var customer Customer
db.Preload("Orders.Items").First(&customer, 1)
```

### Many-to-Many

```go
// Automatic junction table
type Student struct {
    gorm.Model
    Name    string
    Courses []Course `gorm:"many2many:enrollments;"`
}

type Course struct {
    gorm.Model
    Name     string
    Students []Student `gorm:"many2many:enrollments;"`
}

// Custom junction table with extra fields
type Enrollment struct {
    StudentID   uint      `gorm:"primaryKey"`
    CourseID    uint      `gorm:"primaryKey"`
    EnrolledAt  time.Time
    Grade       string
    CompletedAt *time.Time
}

// Override default junction table
func (Enrollment) TableName() string {
    return "student_courses"
}
```

### Self-Referencing

```go
type Category struct {
    gorm.Model
    Name       string
    ParentID   *uint
    Parent     *Category  `gorm:"foreignKey:ParentID"`
    Children   []Category `gorm:"foreignKey:ParentID"`
}

// Usage
var rootCategories []Category
db.Where("parent_id IS NULL").Preload("Children.Children").Find(&rootCategories)
```

### Polymorphic Associations

```go
// Polymorphic model
type Comment struct {
    gorm.Model
    Body            string
    CommentableID   uint
    CommentableType string
}

// Models that can have comments
type Post struct {
    gorm.Model
    Title    string
    Comments []Comment `gorm:"polymorphic:Commentable;"`
}

type Video struct {
    gorm.Model
    URL      string
    Comments []Comment `gorm:"polymorphic:Commentable;"`
}

// Usage
post := Post{Title: "Hello"}
db.Create(&post)

comment := Comment{
    Body:            "Great post!",
    CommentableID:   post.ID,
    CommentableType: "posts",
}
db.Create(&comment)
```

### Has One

```go
type User struct {
    gorm.Model
    Profile Profile
}

type Profile struct {
    gorm.Model
    UserID uint   `gorm:"uniqueIndex"`
    User   User
    Bio    string
}
```

### Eager Loading Strategies

```go
// Preload specific relations
db.Preload("Orders.Items.Product").Find(&customers)

// Conditional preloading
db.Preload("Orders", "status = ?", "pending").Find(&customers)

// Nested preloading with conditions
db.Preload("Orders", func(db *gorm.DB) *gorm.DB {
    return db.Order("created_at DESC").Limit(10)
}).Find(&customers)
```

</details>

<details>
<summary><strong>Deep Dive: GORM Hooks & Advanced Features</strong></summary>

### GORM Hooks

```go
type User struct {
    gorm.Model
    Email        string
    PasswordHash string
    LastLoginAt  *time.Time
}

// BeforeCreate - runs before INSERT
func (u *User) BeforeCreate(tx *gorm.DB) error {
    u.Email = strings.ToLower(strings.TrimSpace(u.Email))
    return nil
}

// AfterCreate - runs after INSERT
func (u *User) AfterCreate(tx *gorm.DB) error {
    // Send welcome email
    go sendWelcomeEmail(u.Email)
    return nil
}

// BeforeUpdate - runs before UPDATE
func (u *User) BeforeUpdate(tx *gorm.DB) error {
    if tx.Statement.Changed("Email") {
        u.Email = strings.ToLower(strings.TrimSpace(u.Email))
    }
    return nil
}

// BeforeSave - runs before both CREATE and UPDATE
func (u *User) BeforeSave(tx *gorm.DB) error {
    // Validation logic
    if u.Email == "" {
        return errors.New("email cannot be empty")
    }
    return nil
}
```

### Scopes for Reusable Queries

```go
// Define reusable query scopes
func Active(db *gorm.DB) *gorm.DB {
    return db.Where("deleted_at IS NULL AND is_active = ?", true)
}

func OrderByCreated(db *gorm.DB) *gorm.DB {
    return db.Order("created_at DESC")
}

func WithinDateRange(start, end time.Time) func(db *gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("created_at BETWEEN ? AND ?", start, end)
    }
}

// Usage
var users []User
db.Scopes(Active, OrderByCreated).Find(&users)
db.Scopes(WithinDateRange(startDate, endDate)).Find(&orders)
```

### Transaction Management

```go
// Manual transaction
tx := db.Begin()
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
    }
}()

if err := tx.Create(&order).Error; err != nil {
    tx.Rollback()
    return err
}

if err := tx.Create(&payment).Error; err != nil {
    tx.Rollback()
    return err
}

tx.Commit()

// Using Transaction function
err := db.Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&order).Error; err != nil {
        return err
    }
    if err := tx.Create(&payment).Error; err != nil {
        return err
    }
    return nil
})
```

### Advanced Query Techniques

```go
// Subqueries
var users []User
db.Where("id IN (?)",
    db.Table("orders").Select("user_id").Where("total > ?", 100),
).Find(&users)

// Raw SQL with scan
type Result struct {
    Date  time.Time
    Total float64
}

var results []Result
db.Raw(`
    SELECT DATE(created_at) as date, SUM(total) as total
    FROM orders
    WHERE created_at > ?
    GROUP BY DATE(created_at)
`, lastMonth).Scan(&results)

// Locking
db.Clauses(clause.Locking{Strength: "UPDATE"}).Find(&user, 1)
```

</details>

<details>
<summary><strong>Deep Dive: Migrations with GORM & golang-migrate</strong></summary>

### GORM AutoMigrate (Development)

```go
// Simple auto-migration for development
func InitDB() (*gorm.DB, error) {
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        return nil, err
    }

    // AutoMigrate creates/updates tables
    err = db.AutoMigrate(
        &User{},
        &Order{},
        &Product{},
        &OrderItem{},
    )
    return db, err
}

// Custom migration with raw SQL
func CustomMigration(db *gorm.DB) error {
    return db.Exec(`
        CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_created
        ON orders(created_at DESC) WHERE deleted_at IS NULL
    `).Error
}
```

### Production Migrations with golang-migrate

```bash
# Install golang-migrate
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Create migration files
migrate create -ext sql -dir migrations -seq create_users_table
```

```sql
-- migrations/000001_create_users_table.up.sql
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

-- migrations/000001_create_users_table.down.sql
DROP TABLE IF EXISTS users;
```

### Zero-Downtime Migration Strategy

```go
// Step 1: Add nullable column
type User struct {
    gorm.Model
    Email    string
    Phone    *string `gorm:"type:varchar(20)"` // Nullable initially
}

// Step 2: Backfill data
func BackfillPhones(db *gorm.DB) error {
    return db.Model(&User{}).
        Where("phone IS NULL").
        Update("phone", "").Error
}

// Step 3: Make non-nullable
func MakePhoneRequired(db *gorm.DB) error {
    return db.Exec(`
        ALTER TABLE users
        ALTER COLUMN phone SET NOT NULL
    `).Error
}
```

### Migration Management Code

```go
package migrations

import (
    "database/sql"
    "embed"

    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/postgres"
    "github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed *.sql
var fs embed.FS

func RunMigrations(db *sql.DB) error {
    driver, err := postgres.WithInstance(db, &postgres.Config{})
    if err != nil {
        return err
    }

    source, err := iofs.New(fs, ".")
    if err != nil {
        return err
    }

    m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
    if err != nil {
        return err
    }

    // Run migrations
    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        return err
    }

    return nil
}
```

### Best Practices

| Practice                | Implementation                                    |
|-------------------------|---------------------------------------------------|
| Version control         | Store migrations in `migrations/` directory      |
| Idempotent migrations   | Use `IF NOT EXISTS` clauses                      |
| Concurrent indexes      | `CREATE INDEX CONCURRENTLY` in PostgreSQL        |
| Test rollbacks          | Always test down migrations                      |
| Separate DDL and DML    | Schema changes separate from data changes        |
| Lock timeout            | Set `lock_timeout` to prevent long waits         |

</details>

<details>
<summary><strong>Deep Dive: Performance Optimization with GORM & PostgreSQL</strong></summary>

### Query Analysis in PostgreSQL

```go
// Enable query logging in GORM
db = db.Debug()

// Analyze query performance
var explain []map[string]interface{}
db.Raw("EXPLAIN (ANALYZE, BUFFERS) SELECT * FROM users WHERE email = ?", email).Scan(&explain)

// Check slow queries
db.Raw(`
    SELECT query, calls, mean_exec_time, max_exec_time
    FROM pg_stat_statements
    WHERE mean_exec_time > 100
    ORDER BY mean_exec_time DESC
    LIMIT 10
`).Scan(&slowQueries)
```

### Solving N+1 Query Problem

```go
// BAD: N+1 queries
var users []User
db.Find(&users)
for _, user := range users {
    var orders []Order
    db.Where("user_id = ?", user.ID).Find(&orders)
}

// GOOD: Eager loading with Preload
var users []User
db.Preload("Orders").Find(&users)

// BETTER: Selective preloading
db.Preload("Orders", "status = ?", "pending").
   Preload("Orders.Items").
   Find(&users)

// BEST: Joins for read-only data
var results []struct {
    UserName  string
    OrderID   uint
    Total     decimal.Decimal
}
db.Table("users").
    Select("users.name as user_name, orders.id as order_id, orders.total").
    Joins("LEFT JOIN orders ON orders.user_id = users.id").
    Scan(&results)
```

### Connection Pool Optimization

```go
sqlDB, err := db.DB()

// SetMaxIdleConns sets the maximum number of connections in the idle pool
sqlDB.SetMaxIdleConns(10)

// SetMaxOpenConns sets the maximum number of open connections
sqlDB.SetMaxOpenConns(100)

// SetConnMaxLifetime sets the maximum amount of time a connection may be reused
sqlDB.SetConnMaxLifetime(time.Hour)

// SetConnMaxIdleTime sets the maximum amount of time a connection may be idle
sqlDB.SetConnMaxIdleTime(10 * time.Minute)
```

### Query Optimization Techniques

```go
// 1. Select only needed columns
db.Select("id", "email", "name").Find(&users)

// 2. Use pagination
db.Offset(offset).Limit(pageSize).Find(&users)

// 3. Batch operations
db.CreateInBatches(users, 100)

// 4. Use prepared statements
stmt := db.Session(&gorm.Session{PrepareStmt: true})
stmt.Where("email = ?", email).Find(&user)

// 5. Disable default transaction for read queries
db.Session(&gorm.Session{SkipDefaultTransaction: true}).Find(&users)
```

### PostgreSQL-Specific Optimizations

```go
// Partial indexes for soft deletes
db.Exec(`
    CREATE INDEX idx_active_users
    ON users(email)
    WHERE deleted_at IS NULL
`)

// Use BRIN indexes for time-series data
db.Exec(`
    CREATE INDEX idx_orders_created_brin
    ON orders USING BRIN(created_at)
`)

// Table partitioning for large tables
db.Exec(`
    CREATE TABLE orders_2024 PARTITION OF orders
    FOR VALUES FROM ('2024-01-01') TO ('2025-01-01')
`)

// Materialized views for complex aggregations
db.Exec(`
    CREATE MATERIALIZED VIEW daily_sales AS
    SELECT DATE(created_at) as date,
           COUNT(*) as order_count,
           SUM(total) as total_sales
    FROM orders
    GROUP BY DATE(created_at)
`)
```

### Caching Strategy

```go
import "github.com/go-redis/redis/v8"

type CachedDB struct {
    db    *gorm.DB
    cache *redis.Client
}

func (c *CachedDB) GetUser(id uint) (*User, error) {
    key := fmt.Sprintf("user:%d", id)

    // Check cache
    val, err := c.cache.Get(ctx, key).Result()
    if err == nil {
        var user User
        json.Unmarshal([]byte(val), &user)
        return &user, nil
    }

    // Fetch from DB
    var user User
    if err := c.db.First(&user, id).Error; err != nil {
        return nil, err
    }

    // Cache for 1 hour
    data, _ := json.Marshal(user)
    c.cache.Set(ctx, key, data, time.Hour)

    return &user, nil
}
```

</details>

---

## Extension Points

1. **PostgreSQL Advanced Features:** JSONB operations, full-text search, PostGIS for geospatial
2. **Advanced Patterns:** Time-series with TimescaleDB, event sourcing, CQRS, multi-tenancy with RLS
3. **GORM Plugins:** Soft delete, optimistic locking, sharding, tracing
4. **Performance Monitoring:** pg_stat_statements, pgBadger, GORM prometheus metrics
5. **Testing Strategies:** Testcontainers for Go, fixture management, transaction rollback tests
