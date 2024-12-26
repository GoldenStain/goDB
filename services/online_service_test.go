package services

import (
	"context"
	"testing"

	pb "github.com/GoldenStain/goDB/bookstorepb"
	"github.com/GoldenStain/goDB/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDBQuery(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	if err := db.AutoMigrate(&models.Book{}, &models.Customer{}, &models.CustomerOrder{}); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}
	return db
}

func TestQueryBook(t *testing.T) {
	db := setupTestDBQuery(t)
	server := NewOnlineServiceServer(db, 50)

	// 添加书籍
	books := []models.Book{
		{BookNo: "B001", Title: "Go Programming", PublisherName: "Tech Press", Keywords: "programming,go", Authors: "John Doe"},
		{BookNo: "B002", Title: "Python Programming", PublisherName: "Tech Press", Keywords: "programming,python", Authors: "Jane Doe"},
		{BookNo: "B003", Title: "Java Programming", PublisherName: "Tech Press", Keywords: "programming,java", Authors: "Jim Doe"},
	}
	for _, book := range books {
		db.Create(&book)
	}

	tests := []struct {
		input    string
		expected int
	}{
		{"B001", 1},
		{"Programming", 3},
		{"Tech Press", 3},
		{"programming,go", 1},
		{"John Doe", 1},
		{"Nonexistent", 0},
	}

	for _, test := range tests {
		req := &pb.QueryBookRequest{Input: test.input}
		resp, err := server.QueryBook(context.Background(), req)
		t.Logf("QueryBook(%s) = %v", test.input, resp)
		assert.NoError(t, err)
		assert.Equal(t, test.expected, len(resp.Books))
		if test.expected > 0 {
			assert.True(t, resp.Success)
		} else {
			assert.False(t, resp.Success)
		}
	}
}

func TestQueryCustomer(t *testing.T) {
	db := setupTestDBQuery(t)
	server := NewOnlineServiceServer(db, 50)

	// 添加客户
	customers := []models.Customer{
		{OnlineID: "C001", Name: "Alice", Address: "123 Wonderland"},
		{OnlineID: "C002", Name: "Bob", Address: "456 Nowhere"},
		{OnlineID: "C003", Name: "Charlie", Address: "789 Everywhere"},
	}
	for _, customer := range customers {
		db.Create(&customer)
	}

	// 添加订单
	orders := []models.CustomerOrder{
		{CustomerOnlineID: "C001", BookNo: "B001", BookCount: 1, Price: 100, Address: "123 Wonderland", Status: "Shipped"},
		{CustomerOnlineID: "C002", BookNo: "B002", BookCount: 2, Price: 200, Address: "456 Nowhere", Status: "Pending"},
		{CustomerOnlineID: "C003", BookNo: "B003", BookCount: 3, Price: 300, Address: "789 Everywhere", Status: "Delivered"},
	}
	for _, order := range orders {
		db.Create(&order)
	}

	tests := []struct {
		input          string
		expected       int
		expectedOrders int
	}{
		{"C001", 1, 1},
		{"Alice", 1, 1},
		{"123 Wonderland", 1, 1},
		{"C002", 1, 1},
		{"Bob", 1, 1},
		{"456 Nowhere", 1, 1},
		{"C003", 1, 1},
		{"Charlie", 1, 1},
		{"789 Everywhere", 1, 1},
		{"Nonexistent", 0, 0},
	}

	for _, test := range tests {
		req := &pb.QueryCustomerRequest{Input: test.input}
		resp, err := server.QueryCustomer(context.Background(), req)
		t.Logf("QueryCustomer(%s) = %v", test.input, resp.Customers)
		assert.NoError(t, err)
		assert.Equal(t, test.expected, len(resp.Customers))
		if test.expected > 0 {
			assert.True(t, resp.Success)
			assert.Equal(t, test.expectedOrders, len(resp.Customers[0].CustomerOrders))
		} else {
			assert.False(t, resp.Success)
		}
	}
}
