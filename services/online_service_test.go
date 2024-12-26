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

func setupTestDBQuery(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	if err := db.AutoMigrate(&models.Book{}); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}
	return db
}

func TestQueryBook(t *testing.T) {
	db := setupTestDBQuery(t)
	server := NewOnlineServiceServer(db, 50)

	// 添加书籍
	books := []models.Book{
		{BookNo: "B001", Title: "Go Programming", PublisherName: "Tech Press", Keywords: "programming,go", Authors: "John Doe"},
		{BookNo: "B002", Title: "Python Programming", PublisherName: "Tech Press", Keywords: "programming,python", Authors: "Jane Doe"},
		{BookNo: "B003", Title: "Java Programming", PublisherName: "Tech Press", Keywords: "programming,java", Authors: "Jim Doe"},
	}
	for _, book := range books {
		db.Create(&book)
	}

	tests := []struct {
		input    string
		expected int
	}{
		{"B001", 1},
		{"Programming", 3},
		{"Tech Press", 3},
		{"programming,go", 1},
		{"John Doe", 1},
		{"Nonexistent", 0},
	}

	for _, test := range tests {
		req := &pb.QueryBookRequest{Input: test.input}
		resp, err := server.QueryBook(context.Background(), req)
		assert.NoError(t, err)
		assert.Equal(t, test.expected, len(resp.Books))
		if test.expected > 0 {
			assert.True(t, resp.Success)
		} else {
			assert.False(t, resp.Success)
		}
	}
}
