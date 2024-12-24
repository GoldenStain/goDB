package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	pb "github.com/GoldenStain/goDB/bookstorepb"
	"github.com/GoldenStain/goDB/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDBOnlineService(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}
	if err := db.AutoMigrate(&models.Book{}, &models.Customer{}, &models.CustomerOrder{}); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}
	return db
}

func TestOnlineService(t *testing.T) {
	db := setupTestDBOnlineService(t)
	bookServer := NewBookServiceServer(db)
	customerServer := NewCustomerServiceServer(db)
	customerOrderServer := NewCustomerOrderServiceServer(db)
	onlineServiceServer := NewOnlineServiceServer(db)

	// 添加几本书
	for i := 1; i <= 3; i++ {
		createBookReq := &pb.CreateBookRequest{
			BookNo:        fmt.Sprintf("B%03d", i),
			Title:         fmt.Sprintf("Book Title %d", i),
			PublisherName: "Test Publisher",
			Price:         100,
			Keywords:      "Test Keywords",
			Authors:       "Test Author",
			StockQuantity: 50,
		}
		createBookResp, err := bookServer.CreateBook(context.Background(), createBookReq)
		assert.NoError(t, err)
		assert.True(t, createBookResp.Success)
		assert.Equal(t, "Book created successfully", createBookResp.Feedback)
	}

	// 添加几个客户
	for i := 1; i <= 3; i++ {
		createCustomerReq := &pb.CreateCustomerRequest{
			OnlineId:       fmt.Sprintf("customer%d", i),
			Password:       "password",
			Name:           fmt.Sprintf("Customer %d", i),
			Address:        fmt.Sprintf("Address %d", i),
			AccountBalance: 1000,
			CreditLevel:    1,
		}
		createCustomerResp, err := customerServer.CreateCustomer(context.Background(), createCustomerReq)
		assert.NoError(t, err)
		assert.True(t, createCustomerResp.Success)
		assert.Equal(t, "Customer created successfully", createCustomerResp.Feedback)
	}

	// 添加几个订单
	for i := 1; i <= 3; i++ {
		createCustomerOrderReq := &pb.CreateCustomerOrderRequest{
			OrderDate:        time.Now().Format("2006-01-02"),
			CustomerOnlineId: fmt.Sprintf("customer%d", i),
			BookNo:           fmt.Sprintf("B%03d", i),
			BookCount:        1,
			Price:            100,
			Address:          fmt.Sprintf("Address %d", i),
			Status:           "未发货",
		}
		createCustomerOrderResp, err := customerOrderServer.CreateCustomerOrder(context.Background(), createCustomerOrderReq)
		assert.NoError(t, err)
		assert.True(t, createCustomerOrderResp.Success)
		assert.Equal(t, "Customer order created successfully", createCustomerOrderResp.Feedback)
	}

	// 查询并验证书籍
	for i := 1; i <= 3; i++ {
		var book models.Book
		err := db.Where("book_no = ?", fmt.Sprintf("B%03d", i)).First(&book).Error
		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("Book Title %d", i), book.Title)
		assert.Equal(t, "Test Publisher", book.PublisherName)
		assert.Equal(t, int32(100), book.Price)
		assert.Equal(t, "Test Keywords", book.Keywords)
		assert.Equal(t, "Test Author", book.Authors)
		assert.Equal(t, int32(50), book.StockQuantity)
	}

	// 查询并验证客户
	for i := 1; i <= 3; i++ {
		var customer models.Customer
		err := db.Where("online_id = ?", fmt.Sprintf("customer%d", i)).First(&customer).Error
		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("Customer %d", i), customer.Name)
		assert.Equal(t, fmt.Sprintf("Address %d", i), customer.Address)
		assert.Equal(t, int32(1000), customer.AccountBalance)
		assert.Equal(t, int32(1), customer.CreditLevel)
	}

	// 查询并验证订单
	for i := 1; i <= 3; i++ {
		var order models.CustomerOrder
		err := db.Where("customer_online_id = ? AND book_no = ?", fmt.Sprintf("customer%d", i), fmt.Sprintf("B%03d", i)).First(&order).Error
		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("customer%d", i), order.CustomerOnlineID)
		assert.Equal(t, fmt.Sprintf("B%03d", i), order.BookNo)
		assert.Equal(t, int32(1), order.BookCount)
		assert.Equal(t, int32(100), order.Price)
		assert.Equal(t, fmt.Sprintf("Address %d", i), order.Address)
		assert.Equal(t, "未发货", order.Status)
	}

	// 测试 QueryCustomer
	queryCustomerReq := &pb.QueryCustomerRequest{
		OnlineId: "customer1",
	}
	queryCustomerResp, err := onlineServiceServer.QueryCustomer(context.Background(), queryCustomerReq)
	assert.NoError(t, err)
	assert.True(t, queryCustomerResp.Success)
	assert.Equal(t, 1, len(queryCustomerResp.Customers))
	assert.Equal(t, "customer1", queryCustomerResp.Customers[0].OnlineId)

	// 测试 QueryBook
	queryBookReq := &pb.QueryBookRequest{
		Title:          "Book Title 1",
		MatchThreshold: 50,
	}
	queryBookResp, err := onlineServiceServer.QueryBook(context.Background(), queryBookReq)
	assert.NoError(t, err)
	assert.True(t, queryBookResp.Success)
	assert.Equal(t, 1, len(queryBookResp.Books))
	assert.Equal(t, "Book Title 1", queryBookResp.Books[0].Title)
}
