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

func setupTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	models.AutoMigrate(db)
	return db
}

func TestCreateUpdateDeletePurchaseOrder(t *testing.T) {
	db := setupTestDB()
	bookServer := NewBookServiceServer(db)
	stockRequestServer := NewStockRequestServiceServer(db)
	purchaseOrderServer := NewPurchaseOrderServiceServer(db)

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

	// 添加 20 个采购记录
	for i := 1; i <= 20; i++ {
		req := &pb.CreatePurchaseOrderRequest{
			BookNo:    fmt.Sprintf("B%03d", (i%3)+1),
			Title:     fmt.Sprintf("Book Title %d", (i%3)+1),
			Publisher: "Test Publisher",
			Supplier:  "Test Supplier",
			Author:    "Test Author",
			Quantity:  int32(10 + i),
			OrderDate: "2023-01-01",
		}
		resp, err := purchaseOrderServer.CreatePurchaseOrder(context.Background(), req)
		assert.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, "Purchase order created successfully", resp.Feedback)
	}

	// 更新所有采购记录并设置 Finished 为 true
	for i := 1; i <= 20; i++ {
		req := &pb.UpdatePurchaseOrderRequest{
			Id:        int32(i),
			Title:     fmt.Sprintf("Book Title %d - updated", (i%3)+1),
			Publisher: "Updated Publisher",
			Supplier:  "Updated Supplier",
			Author:    "Updated Author",
			Quantity:  int32(10),
			OrderDate: "2023-02-01",
			Finished:  true,
		}
		resp, err := purchaseOrderServer.UpdatePurchaseOrder(context.Background(), req)
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
		resp, err := purchaseOrderServer.GetPurchaseOrder(context.Background(), req)
		assert.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, 1, len(resp.PurchaseOrders))

		purchaseOrder := resp.PurchaseOrders[0]
		assert.Equal(t, fmt.Sprintf("B%03d", (i%3)+1), purchaseOrder.BookNo)
		assert.Equal(t, fmt.Sprintf("Book Title %d - updated", (i%3)+1), purchaseOrder.Title)
		assert.Equal(t, "Updated Publisher", purchaseOrder.Publisher)
		assert.Equal(t, "Updated Supplier", purchaseOrder.Supplier)
		assert.Equal(t, "Updated Author", purchaseOrder.Author)
		assert.Equal(t, int32(10), purchaseOrder.Quantity)
		assert.Equal(t, "2023-02-01", purchaseOrder.OrderDate)
	}

	// 验证书籍库存数量是否正确更新
	for i := 1; i <= 3; i++ {
		var book models.Book
		err := db.Where("book_no = ?", fmt.Sprintf("B%03d", i)).First(&book).Error
		assert.NoError(t, err)
		expectedQuantity := int32(50 + 7*10) // 初始库存 + 每本书的采购数量总和
		assert.Equal(t, expectedQuantity, book.StockQuantity)
	}

	// 删除所有采购记录
	for i := 1; i <= 20; i++ {
		req := &pb.DeletePurchaseOrderRequest{
			Id: int32(i),
		}
		resp, err := purchaseOrderServer.DeletePurchaseOrder(context.Background(), req)
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

	// 添加几个缺书登记
	for i := 1; i <= 3; i++ {
		createStockRequestReq := &pb.CreateStockRequestRequest{
			BookNo:      fmt.Sprintf("B%03d", i),
			Title:       fmt.Sprintf("Book Title %d", i),
			Quantity:    10,
			RequestDate: time.Now().Format("2006-01-02"),
			Publisher:   "Test Publisher",
			Author:      "Test Author",
			Supplier:    "Test Supplier",
		}
		createStockRequestResp, err := stockRequestServer.CreateStockRequest(context.Background(), createStockRequestReq)
		assert.NoError(t, err)
		assert.True(t, createStockRequestResp.Success)
		assert.Equal(t, "Stock request created successfully", createStockRequestResp.Feedback)
	}

	// 生成采购单
	generatePurchaseOrdersReq := &pb.GeneratePurchaseOrdersRequest{}
	generatePurchaseOrdersResp, err := purchaseOrderServer.GeneratePurchaseOrdersFromStockRequests(context.Background(), generatePurchaseOrdersReq)
	assert.NoError(t, err)
	assert.True(t, generatePurchaseOrdersResp.Success)
	assert.Equal(t, "Purchase orders generated successfully", generatePurchaseOrdersResp.Feedback)

	// 查询并验证采购单
	for i := 1; i <= 3; i++ {
		var purchaseOrder models.PurchaseOrder
		err := db.Where("book_no = ?", fmt.Sprintf("B%03d", i)).First(&purchaseOrder).Error
		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("Book Title %d", i), purchaseOrder.Title)
		assert.Equal(t, "Test Publisher", purchaseOrder.Publisher)
		assert.Equal(t, "Test Supplier", purchaseOrder.Supplier)
		assert.Equal(t, "Test Author", purchaseOrder.Author)
		assert.Equal(t, int32(10), purchaseOrder.Quantity)
	}

	// 验证缺书登记的 Finished 字段是否更新
	for i := 1; i <= 3; i++ {
		var stockRequest models.StockRequest
		err := db.Where("book_no = ?", fmt.Sprintf("B%03d", i)).First(&stockRequest).Error
		assert.NoError(t, err)
		assert.True(t, stockRequest.Finished)
	}
}
