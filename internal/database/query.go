package database

type Query struct {
	Filters map[string]interface{}
	Page    int
	Limit   int
	OrderBy string
}

func NewQuery() *Query {
	return &Query{
		Filters: make(map[string]interface{}),
		Page:    1,
		Limit:   10,
	}
}

func (q *Query) SetFilter(key string, value interface{}) *Query {
	q.Filters[key] = value
	return q
}

func (q *Query) SetPage(page int) *Query {
	if page > 0 {
		q.Page = page
	}
	return q
}

func (q *Query) SetLimit(limit int) *Query {
	if limit > 0 && limit <= 100 {
		q.Limit = limit
	}
	return q
}

func (q *Query) SetOrderBy(order string) *Query {
	q.OrderBy = order
	return q
}

// Count returns a query that counts total matching rows.
// Implement with your ORM:
//   GORM:  func (q *Query) Count(tx *gorm.DB) *gorm.DB { return tx.Where(q.Filters) }
func (q *Query) Count(tx interface{}) interface{} {
	return tx
}

// Apply applies filters, ordering, and pagination to a query.
// Implement with your ORM:
//   GORM:  tx.Where(q.Filters).Order(q.OrderBy).Offset((q.Page-1)*q.Limit).Limit(q.Limit)
func (q *Query) Apply(tx interface{}) interface{} {
	return tx
}
