package services

import (
	"context"
	"fmt"
	"strings"

	pb "github.com/GoldenStain/goDB/bookstorepb"
	"github.com/GoldenStain/goDB/models"
	"github.com/agnivade/levenshtein"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

type OnlineServiceServer struct {
	pb.UnimplementedOnlineServiceServer
	db *gorm.DB
}

// NewOnlineServiceServer 用于创建 OnlineServiceServer
func NewOnlineServiceServer(db *gorm.DB) *OnlineServiceServer {
	return &OnlineServiceServer{
		db: db,
	}
}

// QueryCustomer 查询客户
func (s *OnlineServiceServer) QueryCustomer(ctx context.Context, req *pb.QueryCustomerRequest) (*pb.QueryCustomerResponse, error) {
	var customers []models.Customer

	query := s.db.Model(&models.Customer{})

	if req.GetOnlineId() != "" {
		query = query.Where("online_id = ?", req.GetOnlineId())
	}
	if req.GetName() != "" {
		query = query.Where("name LIKE ?", "%"+req.GetName()+"%")
	}
	if req.GetAddress() != "" {
		query = query.Where("address LIKE ?", "%"+req.GetAddress()+"%")
	}
	if req.GetOrderId() != 0 {
		query = query.Joins("JOIN customer_orders ON customer_orders.customer_online_id = customers.online_id").
			Where("customer_orders.id = ?", req.GetOrderId())
	}

	if err := query.Find(&customers).Error; err != nil {
		return &pb.QueryCustomerResponse{
			Success:  false,
			Feedback: fmt.Sprintf("failed to query customers: %v", err),
		}, nil
	}

	var pbCustomers []*pb.Customer
	for _, customer := range customers {
		pbCustomer := &pb.Customer{
			Id:             customer.ID,
			OnlineId:       customer.OnlineID,
			Name:           customer.Name,
			Address:        customer.Address,
			AccountBalance: customer.AccountBalance,
			CreditLevel:    customer.CreditLevel,
			CreatedAt:      timestamppb.New(customer.CreatedAt),
			UpdatedAt:      timestamppb.New(customer.UpdatedAt),
		}
		pbCustomers = append(pbCustomers, pbCustomer)
	}

	return &pb.QueryCustomerResponse{
		Success:   true,
		Feedback:  "Customers retrieved successfully",
		Customers: pbCustomers,
	}, nil
}

// QueryBook 查询书籍
func (s *OnlineServiceServer) QueryBook(ctx context.Context, req *pb.QueryBookRequest) (*pb.QueryBookResponse, error) {
	var books []models.Book

	query := s.db.Model(&models.Book{})

	if req.GetBookNo() != "" {
		query = query.Where("book_no = ?", req.GetBookNo())
	}
	if req.GetTitle() != "" {
		query = query.Where("title LIKE ?", "%"+req.GetTitle()+"%")
	}
	if req.GetPublisherName() != "" {
		query = query.Where("publisher_name LIKE ?", "%"+req.GetPublisherName()+"%")
	}
	if req.GetKeywords() != "" {
		query = query.Where("keywords LIKE ?", "%"+req.GetKeywords()+"%")
	}
	if req.GetAuthors() != "" {
		query = query.Where("authors LIKE ?", "%"+req.GetAuthors()+"%")
	}

	if err := query.Find(&books).Error; err != nil {
		return &pb.QueryBookResponse{
			Success:  false,
			Feedback: fmt.Sprintf("failed to query books: %v", err),
		}, nil
	}

	var pbBooks []*pb.Book
	for _, book := range books {
		if isBookMatch(req, book) {
			pbBook := &pb.Book{
				Id:            book.ID,
				BookNo:        book.BookNo,
				Title:         book.Title,
				PublisherName: book.PublisherName,
				Price:         book.Price,
				Keywords:      book.Keywords,
				Authors:       book.Authors,
				StockQuantity: book.StockQuantity,
				CreatedAt:     timestamppb.New(book.CreatedAt),
				UpdatedAt:     timestamppb.New(book.UpdatedAt),
			}
			pbBooks = append(pbBooks, pbBook)
		}
	}

	return &pb.QueryBookResponse{
		Success:  true,
		Feedback: "Books retrieved successfully",
		Books:    pbBooks,
	}, nil
}

func isBookMatch(req *pb.QueryBookRequest, book models.Book) bool {
	matchThreshold := req.GetMatchThreshold()

	if req.GetBookNo() != "" && !isMatch(req.GetBookNo(), book.BookNo, matchThreshold) {
		return false
	}
	if req.GetTitle() != "" && !isMatch(req.GetTitle(), book.Title, matchThreshold) {
		return false
	}
	if req.GetPublisherName() != "" && !isMatch(req.GetPublisherName(), book.PublisherName, matchThreshold) {
		return false
	}
	if req.GetKeywords() != "" && !isKeywordsMatch(req.GetKeywords(), book.Keywords, matchThreshold) {
		return false
	}
	if req.GetAuthors() != "" && !isAuthorsMatch(req.GetAuthors(), book.Authors, matchThreshold) {
		return false
	}

	return true
}

func isMatch(query, target string, threshold int32) bool {
	distance := levenshtein.ComputeDistance(query, target)
	maxLen := max(len(query), len(target))
	matchScore := (1 - float64(distance)/float64(maxLen)) * 100
	return matchScore >= float64(threshold)
}

func isKeywordsMatch(query, target string, threshold int32) bool {
	queryKeywords := strings.Split(query, ",")
	targetKeywords := strings.Split(target, ",")

	for _, qk := range queryKeywords {
		matched := false
		for _, tk := range targetKeywords {
			if isMatch(qk, tk, threshold) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}

func isAuthorsMatch(query, target string, threshold int32) bool {
	queryAuthors := strings.Split(query, ",")
	targetAuthors := strings.Split(target, ",")

	for _, qa := range queryAuthors {
		matched := false
		for _, ta := range targetAuthors {
			if isMatch(qa, ta, threshold) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
