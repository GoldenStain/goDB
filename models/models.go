package models

import (
	"time"

	"gorm.io/gorm"
)

// 书籍
// 如果有多个关键字或者作者，那么用逗号连接他们
type Book struct {
	ID            int32  `gorm:"primaryKey"`
	BookNo        string `gorm:"unique;not null;size:255"`
	Title         string `gorm:"size:255;not null"`
	PublisherName string `gorm:"size:255"`
	Price         int32  `gorm:"not null;default:0"`
	Keywords      string `gorm:"size:1024"`
	Authors       string `gorm:"size:1024"`
	StockQuantity int32
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// 客户
type Customer struct {
	ID             int32            `gorm:"primaryKey"`
	OnlineID       string           `gorm:"unique;not null;size:255"`
	Password       string           `gorm:"size:255;not null"`
	Name           string           `gorm:"size:255;not null"`
	Address        string           `gorm:"size:512;not null"`
	AccountBalance int32            `gorm:"not null;default:0"`
	CreditLevel    int32            `gorm:"not null"`
	CustomerOrders []*CustomerOrder `gorm:"foreignKey:CustomerOnlineID"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type CustomerOrder struct {
	ID        int32  `gorm:"primaryKey"`
	OrderDate string `gorm:"not null;size:50"`
	// 客户的在线ID
	CustomerOnlineID string `gorm:"size:255"`
	BookNo           string `gorm:"not null;size:255"`
	BookCount        int32  `gorm:"not null"`
	Price            int32  `gorm:"not null;default:0"`
	Address          string `gorm:"size:512;not null"`
	Status           string `gorm:"not null;size:50;default:'未发货'"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// 缺书登记
type StockRequest struct {
	ID          int32  `gorm:"primaryKey"`
	BookNo      string `gorm:"not null;size:255"`
	Title       string `gorm:"size:255;not null"`
	Quantity    int32  `gorm:"not null"`
	RequestDate string `gorm:"not null;size:50"`
	Publisher   string `gorm:"size:255"`
	Author      string `gorm:"size:255"`
	Supplier    string `gorm:"size:255"`
	Finished    bool   `gorm:"not null;default:false"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// 采购单
type PurchaseOrder struct {
	ID        int32  `gorm:"primaryKey"`
	BookNo    string `gorm:"not null;size:255"`
	Title     string `gorm:"size:255;not null"`
	Publisher string `gorm:"size:255;not null"`
	Supplier  string `gorm:"not null;size:255"`
	Author    string `gorm:"size:255;not null"`
	Quantity  int32  `gorm:"not null"`
	OrderDate string `gorm:"not null;size:50"`
	Finished  bool   `gorm:"not null;default:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SupplyBook struct {
	ID            int32  `gorm:"primaryKey"`
	BookNo        string `gorm:"unique;not null;size:255"`
	Title         string `gorm:"size:255;not null"`
	PublisherName string `gorm:"size:255"`
	Price         int32  `gorm:"not null;default:0"`
	Quantity      int32  `gorm:"not null;default:0"`
	SupplierID    int32  `gorm:"not null;default:0"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Supplier struct {
	ID          int32         `gorm:"primaryKey"`
	Name        string        `gorm:"size:255;not null"`
	BasicInfo   string        `gorm:"size:1024"`
	SupplyInfo  string        `gorm:"size:1024"`
	SupplyBooks []*SupplyBook `gorm:"foreignKey:SupplierID"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// AutoMigrate 自动迁移函数
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&Book{},
		&Customer{},
		&StockRequest{},
		&PurchaseOrder{},
		&CustomerOrder{},
		&Supplier{},
		&SupplyBook{},
	)
}
