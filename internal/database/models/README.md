# models/

**Purpose:** Domain entity/struct definitions — one file per database table.

**What it can contain:**
- Struct definitions with ORM tags (`gorm`, `db`, `json`, etc.)
- `TableName()` methods for custom table mapping
- Composite types, enums, or value objects tied to entities

**Recommendation:**
- Use GORM tags (`gorm:"column:name;type:varchar(255)"`) if using GORM
- Use `db` struct tags if using SQLx
- Keep models pure — no business logic in this folder

**Conventions:**
- File name: `{entity}.model.go` (e.g. `user.model.go`)
- Struct name: singular PascalCase
- Define `TableName()` if table differs from the pluralized struct name

**Example:**

```go
type User struct {
    ID        string    `gorm:"primaryKey"`
    Email     string    `gorm:"uniqueIndex;not null"`
    Name      string    `gorm:"type:varchar(255)"`
    CreatedAt time.Time
}

func (User) TableName() string { return "users" }
```

This folder can also hold embedded structs, enums, or shared value objects related to your entities.
