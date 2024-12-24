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

func setupTestDBCustomerService(t *testing.T) *gorm.DB {
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err := models.AutoMigrate(db); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}
	return db
}

func TestCustomerService(t *testing.T) {
	db := setupTestDBCustomerService(t)
	server := NewCustomerServiceServer(db)

	// 添加一些客户
	for i := 1; i <= 5; i++ {
		req := &pb.CreateCustomerRequest{
			OnlineId:       fmt.Sprintf("online_id_%d", i),
			Password:       "password",
			Name:           fmt.Sprintf("Customer %d", i),
			Address:        fmt.Sprintf("Address %d", i),
			AccountBalance: int32(100 * i),
			CreditLevel:    int32(i),
		}
		resp, err := server.CreateCustomer(context.Background(), req)
		assert.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, "Customer created successfully", resp.Feedback)
	}

	// 验证 GetCustomer 方法
	getReq := &pb.GetCustomerRequest{
		Start: 0,
		Stop:  4,
	}
	getResp, err := server.GetCustomer(context.Background(), getReq)
	assert.NoError(t, err)
	assert.True(t, getResp.Success)
	assert.Equal(t, 5, len(getResp.Customers))

	for i, customer := range getResp.Customers {
		assert.Equal(t, fmt.Sprintf("online_id_%d", i+1), customer.OnlineId)
		assert.Equal(t, "password", customer.Password)
		assert.Equal(t, fmt.Sprintf("Customer %d", i+1), customer.Name)
		assert.Equal(t, fmt.Sprintf("Address %d", i+1), customer.Address)
		assert.Equal(t, int32(100*(i+1)), customer.AccountBalance)
		assert.Equal(t, int32(i+1), customer.CreditLevel)
	}

	// 修改客户信息
	updateReq := &pb.UpdateCustomerRequest{
		Id:             1,
		OnlineId:       "updated_online_id",
		Password:       "updated_password",
		Name:           "Updated Customer",
		Address:        "Updated Address",
		AccountBalance: 500,
		CreditLevel:    10,
	}
	updateResp, err := server.UpdateCustomer(context.Background(), updateReq)
	assert.NoError(t, err)
	assert.True(t, updateResp.Success)
	assert.Equal(t, "Customer updated successfully", updateResp.Feedback)

	// 再次验证 GetCustomer 方法
	getReq = &pb.GetCustomerRequest{
		Start: 0,
		Stop:  4,
	}
	getResp, err = server.GetCustomer(context.Background(), getReq)
	assert.NoError(t, err)
	assert.True(t, getResp.Success)
	assert.Equal(t, 5, len(getResp.Customers))

	updatedCustomer := getResp.Customers[0]
	assert.Equal(t, "updated_online_id", updatedCustomer.OnlineId)
	assert.Equal(t, "updated_password", updatedCustomer.Password)
	assert.Equal(t, "Updated Customer", updatedCustomer.Name)
	assert.Equal(t, "Updated Address", updatedCustomer.Address)
	assert.Equal(t, int32(500), updatedCustomer.AccountBalance)
	assert.Equal(t, int32(10), updatedCustomer.CreditLevel)

	// 删除客户
	deleteReq := &pb.DeleteCustomerRequest{
		Id: 1,
	}
	deleteResp, err := server.DeleteCustomer(context.Background(), deleteReq)
	assert.NoError(t, err)
	assert.True(t, deleteResp.Success)
	assert.Equal(t, "Customer deleted successfully", deleteResp.Feedback)

	// 验证客户是否删除成功
	getReq = &pb.GetCustomerRequest{
		Start: 0,
		Stop:  4,
	}
	getResp, err = server.GetCustomer(context.Background(), getReq)
	assert.NoError(t, err)
	assert.True(t, getResp.Success)
	assert.Equal(t, 4, len(getResp.Customers))
}
