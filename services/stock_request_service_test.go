package services

import (
	"context"
	"fmt"
	"testing"

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

func TestCreateAndUpdateStockRequest(t *testing.T) {
	db := setupTestDBStockRequest(t)
	server := NewStockRequestServiceServer(db)

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

	// 更新所有缺书登记为 finished = true
	for i := 1; i <= 10; i++ {
		req := &pb.UpdateStockRequestRequest{
			Id:       int32(i),
			Finished: true,
		}
		resp, err := server.UpdateStockRequest(context.Background(), req)
		assert.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, "Stock request updated successfully", resp.Feedback)
	}

	// 查询并验证所有缺书登记的 finished 字段是否为 true
	for i := 1; i <= 10; i++ {
		var stockRequest models.StockRequest
		err := db.First(&stockRequest, int32(i)).Error
		assert.NoError(t, err)
		assert.True(t, stockRequest.Finished)
	}
}
