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

func setupTestDBSupplierService(t *testing.T) *gorm.DB {
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err := models.AutoMigrate(db); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}
	return db
}

func TestSupplierService(t *testing.T) {
	db := setupTestDBSupplierService(t)
	supplierServer := NewSupplierServiceServer(db)
	supplyBookServer := NewSupplyBookServiceServer(db)

	// 创建供应商及其供书记录
	for i := 1; i <= 3; i++ {
		createSupplierReq := &pb.CreateSupplierRequest{
			Name:       fmt.Sprintf("Supplier %d", i),
			BasicInfo:  fmt.Sprintf("Basic Info %d", i),
			SupplyInfo: fmt.Sprintf("Supply Info %d", i),
			SupplyBooks: []*pb.SupplyBook{
				{
					BookNo:        fmt.Sprintf("B%03d", i),
					Title:         fmt.Sprintf("Book Title %d", i),
					PublisherName: fmt.Sprintf("Publisher %d", i),
					Price:         int32(100 * i),
					Quantity:      int32(10 * i),
				},
			},
		}
		createSupplierResp, err := supplierServer.CreateSupplier(context.Background(), createSupplierReq)
		assert.NoError(t, err)
		assert.True(t, createSupplierResp.Success)
		assert.Equal(t, "Supplier created successfully", createSupplierResp.Feedback)
	}

	// 查询供应商及其供书记录
	getSupplierReq := &pb.GetSupplierRequest{
		Start: 0,
		Stop:  2,
	}
	getSupplierResp, err := supplierServer.GetSupplier(context.Background(), getSupplierReq)
	assert.NoError(t, err)
	assert.True(t, getSupplierResp.Success)
	assert.Equal(t, 3, len(getSupplierResp.Suppliers))

	for i, supplier := range getSupplierResp.Suppliers {
		assert.Equal(t, fmt.Sprintf("Supplier %d", i+1), supplier.Name)
		assert.Equal(t, fmt.Sprintf("Basic Info %d", i+1), supplier.BasicInfo)
		assert.Equal(t, fmt.Sprintf("Supply Info %d", i+1), supplier.SupplyInfo)
		assert.Equal(t, 1, len(supplier.SupplyBooks))
		assert.Equal(t, fmt.Sprintf("B%03d", i+1), supplier.SupplyBooks[0].BookNo)
		assert.Equal(t, fmt.Sprintf("Book Title %d", i+1), supplier.SupplyBooks[0].Title)
		assert.Equal(t, fmt.Sprintf("Publisher %d", i+1), supplier.SupplyBooks[0].PublisherName)
		assert.Equal(t, int32(100*(i+1)), supplier.SupplyBooks[0].Price)
		assert.Equal(t, int32(10*(i+1)), supplier.SupplyBooks[0].Quantity)
	}

	// 更新供应商及其供书记录
	updateSupplierReq := &pb.UpdateSupplierRequest{
		Id:         getSupplierResp.Suppliers[0].Id,
		Name:       "Updated Supplier",
		BasicInfo:  "Updated Basic Info",
		SupplyInfo: "Updated Supply Info",
		SupplyBooks: []*pb.SupplyBook{
			{
				BookNo:        "B001-updated",
				Title:         "Updated Book Title",
				PublisherName: "Updated Publisher",
				Price:         200,
				Quantity:      20,
			},
		},
	}
	updateSupplierResp, err := supplierServer.UpdateSupplier(context.Background(), updateSupplierReq)
	assert.NoError(t, err)
	assert.True(t, updateSupplierResp.Success)
	assert.Equal(t, "Supplier updated successfully", updateSupplierResp.Feedback)

	// 再次查询供应商及其供书记录
	getSupplierResp, err = supplierServer.GetSupplier(context.Background(), getSupplierReq)
	assert.NoError(t, err)
	assert.True(t, getSupplierResp.Success)
	assert.Equal(t, 3, len(getSupplierResp.Suppliers))

	updatedSupplier := getSupplierResp.Suppliers[0]
	assert.Equal(t, "Updated Supplier", updatedSupplier.Name)
	assert.Equal(t, "Updated Basic Info", updatedSupplier.BasicInfo)
	assert.Equal(t, "Updated Supply Info", updatedSupplier.SupplyInfo)
	assert.Equal(t, 1, len(updatedSupplier.SupplyBooks))
	assert.Equal(t, "B001-updated", updatedSupplier.SupplyBooks[0].BookNo)
	assert.Equal(t, "Updated Book Title", updatedSupplier.SupplyBooks[0].Title)
	assert.Equal(t, "Updated Publisher", updatedSupplier.SupplyBooks[0].PublisherName)
	assert.Equal(t, int32(200), updatedSupplier.SupplyBooks[0].Price)
	assert.Equal(t, int32(20), updatedSupplier.SupplyBooks[0].Quantity)

	// 单独更新供书记录
	updateSupplyBookReq := &pb.UpdateSupplyBookRequest{
		Id:            updatedSupplier.SupplyBooks[0].Id,
		BookNo:        "B001-updated-again",
		Title:         "Updated Book Title Again",
		PublisherName: "Updated Publisher Again",
		Price:         300,
		Quantity:      30,
		SupplierId:    updatedSupplier.Id,
	}
	updateSupplyBookResp, err := supplyBookServer.UpdateSupplyBook(context.Background(), updateSupplyBookReq)
	assert.NoError(t, err)
	assert.True(t, updateSupplyBookResp.Success)
	assert.Equal(t, "Supply book updated successfully", updateSupplyBookResp.Feedback)

	// 查询供书记录通过ID
	getSupplyBookByIDReq := &pb.GetSupplyBookByIDRequest{
		Id: updatedSupplier.SupplyBooks[0].Id,
	}
	getSupplyBookByIDResp, err := supplyBookServer.GetSupplyBookByID(context.Background(), getSupplyBookByIDReq)
	assert.NoError(t, err)
	assert.True(t, getSupplyBookByIDResp.Success)
	assert.Equal(t, "Supply book retrieved successfully", getSupplyBookByIDResp.Feedback)

	retrievedSupplyBook := getSupplyBookByIDResp.SupplyBook
	assert.Equal(t, "B001-updated-again", retrievedSupplyBook.BookNo)
	assert.Equal(t, "Updated Book Title Again", retrievedSupplyBook.Title)
	assert.Equal(t, "Updated Publisher Again", retrievedSupplyBook.PublisherName)
	assert.Equal(t, int32(300), retrievedSupplyBook.Price)
	assert.Equal(t, int32(30), retrievedSupplyBook.Quantity)

	// 查询供书记录通过供应商ID
	getSupplyBooksReq := &pb.GetSupplyBooksBySupplierRequest{
		SupplierId: updatedSupplier.Id,
	}
	getSupplyBooksResp, err := supplyBookServer.GetSupplyBooksBySupplier(context.Background(), getSupplyBooksReq)
	assert.NoError(t, err)
	assert.True(t, getSupplyBooksResp.Success)
	assert.Equal(t, 1, len(getSupplyBooksResp.SupplyBooks))

	updatedSupplyBook := getSupplyBooksResp.SupplyBooks[0]
	assert.Equal(t, "B001-updated-again", updatedSupplyBook.BookNo)
	assert.Equal(t, "Updated Book Title Again", updatedSupplyBook.Title)
	assert.Equal(t, "Updated Publisher Again", updatedSupplyBook.PublisherName)
	assert.Equal(t, int32(300), updatedSupplyBook.Price)
	assert.Equal(t, int32(30), updatedSupplyBook.Quantity)

	// 删除供书记录
	deleteSupplyBookReq := &pb.DeleteSupplyBookRequest{
		Id: updatedSupplyBook.Id,
	}
	deleteSupplyBookResp, err := supplyBookServer.DeleteSupplyBook(context.Background(), deleteSupplyBookReq)
	assert.NoError(t, err)
	assert.True(t, deleteSupplyBookResp.Success)
	assert.Equal(t, "Supply book deleted successfully", deleteSupplyBookResp.Feedback)

	// 验证供书记录是否删除成功
	getSupplyBooksResp, err = supplyBookServer.GetSupplyBooksBySupplier(context.Background(), getSupplyBooksReq)
	assert.NoError(t, err)
	assert.True(t, getSupplyBooksResp.Success)
	assert.Equal(t, 0, len(getSupplyBooksResp.SupplyBooks))

	// 删除供应商
	deleteSupplierReq := &pb.DeleteSupplierRequest{
		Id: updatedSupplier.Id,
	}
	deleteSupplierResp, err := supplierServer.DeleteSupplier(context.Background(), deleteSupplierReq)
	assert.NoError(t, err)
	assert.True(t, deleteSupplierResp.Success)
	assert.Equal(t, "Supplier deleted successfully", deleteSupplierResp.Feedback)

	// 验证供应商是否删除成功
	getSupplierResp, err = supplierServer.GetSupplier(context.Background(), getSupplierReq)
	assert.NoError(t, err)
	assert.True(t, getSupplierResp.Success)
	assert.Equal(t, 2, len(getSupplierResp.Suppliers))
}
