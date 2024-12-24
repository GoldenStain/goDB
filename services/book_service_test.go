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

func setupTestDBBookService() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	models.AutoMigrate(db)
	return db
}

func TestCreateAndGetBook(t *testing.T) {
	db := setupTestDBBookService()
	server := NewBookServiceServer(db)

	// 添加 20 本书
	for i := 0; i < 20; i++ {
		req := &pb.CreateBookRequest{
			BookNo:        fmt.Sprintf("B%03d", i),
			Title:         fmt.Sprintf("Book Title %d", i),
			PublisherName: "Test Publisher",
			Price:         int32(10 + i),
			StockQuantity: int32(100 + i),
			Authors:       "Author1,Author2",
			Keywords:      "Keyword1,Keyword2",
		}
		resp, err := server.CreateBook(context.Background(), req)
		assert.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, "Book created successfully", resp.Feedback)
	}

	// 查询书籍
	getReq := &pb.GetBookRequest{
		Start: 0,
		Stop:  19,
	}
	getResp, err := server.GetBook(context.Background(), getReq)
	assert.NoError(t, err)
	assert.True(t, getResp.Success)
	assert.Equal(t, "Books retrieved successfully", getResp.Feedback)
	assert.Equal(t, 20, len(getResp.Books))

	// 验证查询结果
	for i, book := range getResp.Books {
		t.Logf("Checking book:%v\n", book.Id)
		assert.Equal(t, fmt.Sprintf("B%03d", i), book.BookNo)
		assert.Equal(t, fmt.Sprintf("Book Title %d", i), book.Title)
		assert.Equal(t, "Test Publisher", book.PublisherName)
		assert.Equal(t, float64(10+i), book.Price)
		assert.Equal(t, int32(100+i), book.StockQuantity)
		assert.Equal(t, "Author1,Author2", book.Authors)
		assert.Equal(t, "Keyword1,Keyword2", book.Keywords)
	}
}
