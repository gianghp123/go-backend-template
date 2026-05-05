# dtos/res/

**Purpose:** Response DTOs — shape output data sent to the API consumer.

**Rules:**
- Use `json:"..."` tags with camelCase naming
- Keep them flat — no nested business logic objects
- Only expose fields the client needs

## Example

```go
type ExampleRes struct {
    ID          uint      `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    OwnerID     string    `json:"ownerId"`
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
}
```

**Recommendations:**
- Timestamp fields as `time.Time` (not `string`) — let framework JSON serializer handle formatting
- Use pointers for nullable fields (e.g. `*uint` for optional IDs)
- Never embed models directly — always map through DTOs
