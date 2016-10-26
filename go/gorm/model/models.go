package model

// Customer is a model in the "customers" table.
type Customer struct {
	ID   int
	Name *string `gorm:"not null"`
}

// Order is a model in the "orders" table.
type Order struct {
	ID       int
	Subtotal float64 `gorm:"type:decimal(18,2)"`

	Customer   Customer `gorm:"ForeignKey:CustomerID"`
	CustomerID int      `json:"-"`

	Products []Product `gorm:"many2many:order_products"`
}

// Product is a model in the "products" table.
type Product struct {
	ID    int
	Name  *string `gorm:"not null;unique"`
	Price float64 `gorm:"type:decimal(18,2)"`
}
