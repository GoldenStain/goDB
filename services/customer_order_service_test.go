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

func setupTestDBCustomerOrder(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	if err := models.AutoMigrate(db); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}
	return db
}

func TestCreateCustomerOrder(t *testing.T) {
	db := setupTestDBCustomerOrder(t)
	server := NewCustomerOrderServiceServer(db)

	// 添加客户
	customer := models.Customer{
		OnlineID:       "customer1",
		Password:       "password",
		Name:           "Customer 1",
		Address:        "Address 1",
		AccountBalance: 1000,
		CreditLevel:    1,
	}
	db.Create(&customer)

	// 添加书籍
	book := models.Book{
		BookNo:        "B001",
		Title:         "Book Title 1",
		PublisherName: "Test Publisher",
		Authors:       "Test Author",
		StockQuantity: 10,
	}
	db.Create(&book)

	// 创建客户订单
	req := &pb.CreateCustomerOrderRequest{
		OrderDate:        "2023-01-01",
		CustomerOnlineId: "customer1",
		BookNo:           "B001",
		BookCount:        5,
		Price:            100,
		Address:          "Address 1",
		Status:           "未发货",
	}
	resp, err := server.CreateCustomerOrder(context.Background(), req)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "Customer order created successfully", resp.Feedback)

	// 验证库存更新
	var updatedBook models.Book
	db.First(&updatedBook, "book_no = ?", "B001")
	assert.Equal(t, int32(5), updatedBook.StockQuantity)

	// 验证客户余额更新
	var updatedCustomer models.Customer
	db.First(&updatedCustomer, "online_id = ?", "customer1")
	assert.Equal(t, int32(900), updatedCustomer.AccountBalance)
}

func TestCreateCustomerOrderInsufficientStock(t *testing.T) {
	db := setupTestDBCustomerOrder(t)
	server := NewCustomerOrderServiceServer(db)

	// 添加客户
	customer := models.Customer{
		OnlineID:       "customer1",
		Password:       "password",
		Name:           "Customer 1",
		Address:        "Address 1",
		AccountBalance: 1000,
		CreditLevel:    1,
	}
	db.Create(&customer)

	// 添加书籍
	book := models.Book{
		BookNo:        "B001",
		Title:         "Book Title 1",
		PublisherName: "Test Publisher",
		Authors:       "Test Author",
		StockQuantity: 2,
	}
	db.Create(&book)

	// 创建客户订单
	req := &pb.CreateCustomerOrderRequest{
		OrderDate:        "2023-01-01",
		CustomerOnlineId: "customer1",
		BookNo:           "B001",
		BookCount:        5,
		Price:            100,
		Address:          "Address 1",
		Status:           "未发货",
	}
	resp, err := server.CreateCustomerOrder(context.Background(), req)
	assert.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, "Insufficient stock, stock request created", resp.Feedback)

	// 验证缺书记录创建
	var stockRequest models.StockRequest
	db.First(&stockRequest, "book_no = ?", "B001")
	assert.Equal(t, int32(3), stockRequest.Quantity)
}

func TestGetCustomerOrder(t *testing.T) {
	db := setupTestDBCustomerOrder(t)
	server := NewCustomerOrderServiceServer(db)

	// 添加客户订单
	order := models.CustomerOrder{
		OrderDate:        "2023-01-01",
		CustomerOnlineID: "customer1",
		BookNo:           "B001",
		BookCount:        5,
		Price:            100,
		Address:          "Address 1",
		Status:           "未发货",
	}
	db.Create(&order)

	// 获取客户订单
	req := &pb.GetCustomerOrderRequest{
		Start: 0,
		Stop:  0,
	}
	resp, err := server.GetCustomerOrder(context.Background(), req)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, 1, len(resp.CustomerOrders))
	assert.Equal(t, "B001", resp.CustomerOrders[0].BookNo)
}

func TestUpdateCustomerOrder(t *testing.T) {
	db := setupTestDBCustomerOrder(t)
	server := NewCustomerOrderServiceServer(db)

	// 添加客户订单
	order := models.CustomerOrder{
		OrderDate:        "2023-01-01",
		CustomerOnlineID: "customer1",
		BookNo:           "B001",
		BookCount:        5,
		Price:            100,
		Address:          "Address 1",
		Status:           "未发货",
	}
	db.Create(&order)

	// 更新客户订单
	req := &pb.UpdateCustomerOrderRequest{
		Id:        order.ID,
		BookCount: 10,
		Price:     200,
		Status:    "已发货",
	}
	resp, err := server.UpdateCustomerOrder(context.Background(), req)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "Customer order updated successfully", resp.Feedback)

	// 验证订单更新
	var updatedOrder models.CustomerOrder
	db.First(&updatedOrder, order.ID)
	assert.Equal(t, int32(10), updatedOrder.BookCount)
	assert.Equal(t, int32(200), updatedOrder.Price)
	assert.Equal(t, "已发货", updatedOrder.Status)
}

func TestDeleteCustomerOrder(t *testing.T) {
	db := setupTestDBCustomerOrder(t)
	server := NewCustomerOrderServiceServer(db)

	// 添加客户订单
	order := models.CustomerOrder{
		OrderDate:        "2023-01-01",
		CustomerOnlineID: "customer1",
		BookNo:           "B001",
		BookCount:        5,
		Price:            100,
		Address:          "Address 1",
		Status:           "未发货",
	}
	db.Create(&order)

	// 删除客户订单
	req := &pb.DeleteCustomerOrderRequest{
		Id: order.ID,
	}
	resp, err := server.DeleteCustomerOrder(context.Background(), req)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "Customer order deleted successfully", resp.Feedback)

	// 验证订单删除
	var deletedOrder models.CustomerOrder
	err = db.First(&deletedOrder, order.ID).Error
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}
