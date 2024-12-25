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

func setupTestDBCustomerOrder(t *testing.T) *gorm.DB {
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err := models.AutoMigrate(db); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}
	return db
}

func TestCustomerOrderService(t *testing.T) {
	db := setupTestDBCustomerOrder(t)
	server := NewCustomerOrderServiceServer(db)

	// 添加一些客户
	customers := []models.Customer{
		{OnlineID: "online_id_1", Password: "password1", Name: "Customer 1", Address: "Address 1", AccountBalance: 1000, CreditLevel: 1},
		{OnlineID: "online_id_2", Password: "password2", Name: "Customer 2", Address: "Address 2", AccountBalance: 2000, CreditLevel: 2},
		{OnlineID: "online_id_3", Password: "password3", Name: "Customer 3", Address: "Address 3", AccountBalance: 3000, CreditLevel: 3},
	}
	for _, customer := range customers {
		db.Create(&customer)
	}

	// 添加一些客户订单
	for i := 1; i <= 3; i++ {
		req := &pb.CreateCustomerOrderRequest{
			OrderDate:        "2023-01-01",
			CustomerOnlineId: fmt.Sprintf("online_id_%d", i),
			BookNo:           fmt.Sprintf("B%03d", i),
			BookCount:        int32(10 + i),
			Price:            int32(100 * i),
			Address:          fmt.Sprintf("Address %d", i),
			Status:           "未发货",
		}
		resp, err := server.CreateCustomerOrder(context.Background(), req)
		assert.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, "Customer order created successfully", resp.Feedback)

		// 检查客户余额是否正确变动
		var customer models.Customer
		db.Where("online_id = ?", fmt.Sprintf("online_id_%d", i)).First(&customer)
		expectedBalance := int32(1000*i) - req.Price
		assert.Equal(t, expectedBalance, customer.AccountBalance)
	}

	// 验证 GetCustomerOrder 方法
	getReq := &pb.GetCustomerOrderRequest{
		Start: 0,
		Stop:  2,
	}
	getResp, err := server.GetCustomerOrder(context.Background(), getReq)
	assert.NoError(t, err)
	assert.True(t, getResp.Success)
	assert.Equal(t, 3, len(getResp.CustomerOrders))

	for i, order := range getResp.CustomerOrders {
		assert.Equal(t, "2023-01-01", order.OrderDate)
		assert.Equal(t, fmt.Sprintf("online_id_%d", i+1), order.CustomerOnlineId)
		assert.Equal(t, fmt.Sprintf("B%03d", i+1), order.BookNo)
		assert.Equal(t, int32(10+i+1), order.BookCount)
		assert.Equal(t, int32(100*(i+1)), order.Price)
		assert.Equal(t, fmt.Sprintf("Address %d", i+1), order.Address)
		assert.Equal(t, "未发货", order.Status)
	}

	// 修改客户订单信息
	updateReq := &pb.UpdateCustomerOrderRequest{
		Id:        1,
		OrderDate: "2023-02-01",
		BookNo:    "B001-updated",
		BookCount: 20,
		Price:     200,
		Address:   "Updated Address",
		Status:    "已发货",
	}
	updateResp, err := server.UpdateCustomerOrder(context.Background(), updateReq)
	assert.NoError(t, err)
	assert.True(t, updateResp.Success)
	assert.Equal(t, "Customer order updated successfully", updateResp.Feedback)

	// 再次验证 GetCustomerOrder 方法
	getReq = &pb.GetCustomerOrderRequest{
		Start: 0,
		Stop:  2,
	}
	getResp, err = server.GetCustomerOrder(context.Background(), getReq)
	assert.NoError(t, err)
	assert.True(t, getResp.Success)
	assert.Equal(t, 3, len(getResp.CustomerOrders))

	updatedOrder := getResp.CustomerOrders[0]
	assert.Equal(t, "2023-02-01", updatedOrder.OrderDate)
	assert.Equal(t, "online_id_1", updatedOrder.CustomerOnlineId)
	assert.Equal(t, "B001-updated", updatedOrder.BookNo)
	assert.Equal(t, int32(20), updatedOrder.BookCount)
	assert.Equal(t, int32(200), updatedOrder.Price)
	assert.Equal(t, "Updated Address", updatedOrder.Address)
	assert.Equal(t, "已发货", updatedOrder.Status)

	// 删除客户订单
	deleteReq := &pb.DeleteCustomerOrderRequest{
		Id: 1,
	}
	deleteResp, err := server.DeleteCustomerOrder(context.Background(), deleteReq)
	assert.NoError(t, err)
	assert.True(t, deleteResp.Success)
	assert.Equal(t, "Customer order deleted successfully", deleteResp.Feedback)

	// 验证客户订单是否删除成功
	getReq = &pb.GetCustomerOrderRequest{
		Start: 0,
		Stop:  2,
	}
	getResp, err = server.GetCustomerOrder(context.Background(), getReq)
	assert.NoError(t, err)
	assert.True(t, getResp.Success)
	assert.Equal(t, 2, len(getResp.CustomerOrders))
}
