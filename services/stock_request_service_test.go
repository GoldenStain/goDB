package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	pb "github.com/GoldenStain/goDB/bookstorepb"
	"github.com/GoldenStain/goDB/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDBStockRequest(t *testing.T) *gorm.DB {
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err := models.AutoMigrate(db); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}
	return db
}

func TestCreateUpdateGetDeleteStockRequest(t *testing.T) {
	db := setupTestDBStockRequest(t)
	server := NewStockRequestServiceServer(db)

	// 添加客户
	customers := []models.Customer{
		{OnlineID: "customer1", Password: "password", Name: "Customer 1", Address: "Address 1", AccountBalance: 1000, CreditLevel: 1},
		{OnlineID: "customer2", Password: "password", Name: "Customer 2", Address: "Address 2", AccountBalance: 2000, CreditLevel: 2},
	}
	for _, customer := range customers {
		db.Create(&customer)
	}

	// 添加客户订单并获取生成的订单 ID
	var orderIDs []int32
	orders := []models.CustomerOrder{
		{OrderDate: "2023-01-01", CustomerOnlineID: "customer1", BookNo: "B001", BookCount: 1, Price: 100, Address: "Address 1", Status: "未发货"},
		{OrderDate: "2023-01-01", CustomerOnlineID: "customer2", BookNo: "B002", BookCount: 1, Price: 200, Address: "Address 2", Status: "未发货"},
	}
	for i := range orders {
		db.Create(&orders[i])
		orderIDs = append(orderIDs, orders[i].ID)
	}

	// 添加 10 个缺书登记
	for i := 1; i <= 10; i++ {
		req := &pb.CreateStockRequestRequest{
			BookNo:      fmt.Sprintf("B%03d", i),
			Title:       fmt.Sprintf("Book Title %d", i),
			Publisher:   "Test Publisher",
			Supplier:    "Test Supplier",
			Author:      "Test Author",
			Quantity:    int32(10 + i),
			RequestDate: "2023-01-01",
		}
		resp, err := server.CreateStockRequest(context.Background(), req)
		assert.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, "Stock request created successfully", resp.Feedback)
	}

	// 验证所有缺书登记的插入结果
	for i := 1; i <= 10; i++ {
		var stockRequest models.StockRequest
		err := db.First(&stockRequest, int32(i)).Error
		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("B%03d", i), stockRequest.BookNo)
		assert.Equal(t, fmt.Sprintf("Book Title %d", i), stockRequest.Title)
		assert.Equal(t, "Test Publisher", stockRequest.Publisher)
		assert.Equal(t, "Test Supplier", stockRequest.Supplier)
		assert.Equal(t, "Test Author", stockRequest.Author)
		assert.Equal(t, int32(10+i), stockRequest.Quantity)
		assert.Equal(t, "2023-01-01", stockRequest.RequestDate)
		assert.False(t, stockRequest.Finished)
	}

	// 更新所有缺书登记为 finished = true，并验证电子邮件发送情况
	for i := 1; i <= 10; i++ {
		req := &pb.UpdateStockRequestRequest{
			Id:       int32(i),
			Finished: true,
		}
		resp, err := server.UpdateStockRequest(context.Background(), req)
		assert.NoError(t, err)
		assert.True(t, resp.Success)
		if i == 1 || i == 2 {
			assert.Equal(t, "已暂存电子邮件通知客户", resp.Feedback)
		} else {
			assert.Equal(t, "没有客户创建了书目"+fmt.Sprintf("Book Title %d", i)+"的相关订单，无须发送邮件", resp.Feedback)
		}
	}

	// 查询并验证所有缺书登记的 finished 字段是否为 true
	for i := 1; i <= 10; i++ {
		var stockRequest models.StockRequest
		err := db.First(&stockRequest, int32(i)).Error
		assert.NoError(t, err)
		assert.True(t, stockRequest.Finished)
	}

	// 获取缺书登记
	getReq := &pb.GetStockRequestRequest{
		Start: 0,
		Stop:  9,
	}
	getResp, err := server.GetStockRequest(context.Background(), getReq)
	assert.NoError(t, err)
	assert.True(t, getResp.Success)
	assert.Equal(t, 10, len(getResp.StockRequests))

	for i, stockRequest := range getResp.StockRequests {
		assert.Equal(t, fmt.Sprintf("B%03d", i+1), stockRequest.BookNo)
		assert.Equal(t, fmt.Sprintf("Book Title %d", i+1), stockRequest.Title)
		assert.Equal(t, "Test Publisher", stockRequest.Publisher)
		assert.Equal(t, "Test Supplier", stockRequest.Supplier)
		assert.Equal(t, "Test Author", stockRequest.Author)
		assert.Equal(t, int32(10+i+1), stockRequest.Quantity)
		assert.Equal(t, "2023-01-01", stockRequest.RequestDate)
		assert.True(t, stockRequest.Finished)
	}

	// 删除缺书登记
	for i := 1; i <= 10; i++ {
		deleteReq := &pb.DeleteStockRequestRequest{
			StockRequestId: int32(i),
		}
		deleteResp, err := server.DeleteStockRequest(context.Background(), deleteReq)
		assert.NoError(t, err)
		assert.True(t, deleteResp.Success)
		assert.Equal(t, "Stock request deleted successfully", deleteResp.Feedback)
	}

	// 验证所有缺书登记是否删除成功
	for i := 1; i <= 10; i++ {
		var stockRequest models.StockRequest
		err := db.First(&stockRequest, int32(i)).Error
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	}

	// 验证生成的假电子邮件
	dir := ".\\fake_emails"
	for _, order := range orders {
		today := time.Now().Format("2006-01-02")
		emailFile := filepath.Join(dir, fmt.Sprintf("email_%d_%s.txt", order.ID, today))
		_, err := os.Stat(emailFile)
		if order.BookNo == "B001" || order.BookNo == "B002" {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}

	// 添加一个没有相应订单的缺书登记
	req := &pb.CreateStockRequestRequest{
		BookNo:      "B011",
		Title:       "Book Title 11",
		Publisher:   "Test Publisher",
		Supplier:    "Test Supplier",
		Author:      "Test Author",
		Quantity:    11,
		RequestDate: "2023-01-01",
	}
	resp, err := server.CreateStockRequest(context.Background(), req)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "Stock request created successfully", resp.Feedback)

	// 更新缺书登记为 finished = true，并验证电子邮件发送情况
	reqUpdate := &pb.UpdateStockRequestRequest{
		Id:       11,
		Finished: true,
	}
	respUpdate, err := server.UpdateStockRequest(context.Background(), reqUpdate)
	assert.NoError(t, err)
	assert.True(t, respUpdate.Success)
	assert.Equal(t, "没有客户创建了书目Book Title 11的相关订单，无须发送邮件", respUpdate.Feedback)
}
