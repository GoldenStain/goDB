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

func setupTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	models.AutoMigrate(db)
	return db
}

func TestCreateUpdateDeletePurchaseOrder(t *testing.T) {
	db := setupTestDB()
	server := NewPurchaseOrderServiceServer(db)

	// 添加 20 个采购记录
	for i := 1; i <= 20; i++ {
		req := &pb.CreatePurchaseOrderRequest{
			BookNo:    fmt.Sprintf("B%03d", i),
			Title:     fmt.Sprintf("Book Title %d", i),
			Publisher: "Test Publisher",
			Supplier:  "Test Supplier",
			Author:    "Test Author",
			Quantity:  int32(10 + i),
			OrderDate: "2023-01-01",
		}
		resp, err := server.CreatePurchaseOrder(context.Background(), req)
		assert.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, "Purchase order created successfully", resp.Feedback)
	}

	// 更新所有采购记录
	for i := 1; i <= 20; i++ {
		req := &pb.UpdatePurchaseOrderRequest{
			Id:        int32(i),
			BookNo:    fmt.Sprintf("B%03d-updated", i),
			Title:     fmt.Sprintf("Book Title %d - updated", i),
			Publisher: "Updated Publisher",
			Supplier:  "Updated Supplier",
			Author:    "Updated Author",
			Quantity:  int32(20 + i),
			OrderDate: "2023-02-01",
		}
		resp, err := server.UpdatePurchaseOrder(context.Background(), req)
		assert.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, "Purchase order updated successfully", resp.Feedback)
	}

	// 查询并验证所有采购记录的更新结果
	for i := 1; i <= 20; i++ {
		req := &pb.GetPurchaseOrderRequest{
			Start: int32(i - 1),
			Stop:  int32(i - 1),
		}
		resp, err := server.GetPurchaseOrder(context.Background(), req)
		assert.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, 1, len(resp.PurchaseOrders))

		purchaseOrder := resp.PurchaseOrders[0]
		assert.Equal(t, fmt.Sprintf("B%03d-updated", i), purchaseOrder.BookNo)
		assert.Equal(t, fmt.Sprintf("Book Title %d - updated", i), purchaseOrder.Title)
		assert.Equal(t, "Updated Publisher", purchaseOrder.Publisher)
		assert.Equal(t, "Updated Supplier", purchaseOrder.Supplier)
		assert.Equal(t, "Updated Author", purchaseOrder.Author)
		assert.Equal(t, int32(20+i), purchaseOrder.Quantity)
		assert.Equal(t, "2023-02-01", purchaseOrder.OrderDate)
	}

	// 删除所有采购记录
	for i := 1; i <= 20; i++ {
		req := &pb.DeletePurchaseOrderRequest{
			Id: int32(i),
		}
		resp, err := server.DeletePurchaseOrder(context.Background(), req)
		assert.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, "Purchase order deleted successfully", resp.Feedback)
	}

	// 验证所有采购记录是否删除成功
	for i := 1; i <= 20; i++ {
		var purchaseOrder models.PurchaseOrder
		err := db.First(&purchaseOrder, uint32(i)).Error
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	}
}
