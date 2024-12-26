package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	pb "github.com/GoldenStain/goDB/bookstorepb"
	"github.com/GoldenStain/goDB/models"
	"github.com/agnivade/levenshtein"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

// OnlineServiceServer 定义服务
type OnlineServiceServer struct {
	pb.UnimplementedOnlineServiceServer
	db             *gorm.DB
	matchThreshold int32
}

// NewOnlineServiceServer 用于创建 OnlineServiceServer
func NewOnlineServiceServer(db *gorm.DB, matchThreshole_in int32) *OnlineServiceServer {
	return &OnlineServiceServer{
		db:             db,
		matchThreshold: matchThreshole_in,
	}
}

// QueryCustomer 查询客户
func (s *OnlineServiceServer) QueryCustomer(ctx context.Context, req *pb.QueryCustomerRequest) (*pb.QueryCustomerResponse, error) {
	var customers []models.Customer
	input := req.GetInput()
	var feedbacks []string
	customerMap := make(map[int32]models.Customer)

	if input != "" {
		// 尝试匹配 online_id
		var onlineIDCustomers []models.Customer
		if err := s.db.Where("online_id LIKE ?", "%"+input+"%").Find(&onlineIDCustomers).Error; err == nil && len(onlineIDCustomers) > 0 {
			for _, customer := range onlineIDCustomers {
				customerMap[customer.ID] = customer
			}
			feedbacks = append(feedbacks, "Matched by online_id")
		}

		// 尝试匹配 name
		var nameCustomers []models.Customer
		if err := s.db.Where("name LIKE ?", "%"+input+"%").Find(&nameCustomers).Error; err == nil && len(nameCustomers) > 0 {
			for _, customer := range nameCustomers {
				customerMap[customer.ID] = customer
			}
			feedbacks = append(feedbacks, "Matched by name")
		}

		// 尝试匹配 address
		var addressCustomers []models.Customer
		if err := s.db.Where("address LIKE ?", "%"+input+"%").Find(&addressCustomers).Error; err == nil && len(addressCustomers) > 0 {
			for _, customer := range addressCustomers {
				customerMap[customer.ID] = customer
			}
			feedbacks = append(feedbacks, "Matched by address")
		}

		// 尝试将输入转换为整数并匹配 CustomerOrder ID
		if orderId, err := strconv.Atoi(input); err == nil {
			var customerOrder models.CustomerOrder
			if err := s.db.First(&customerOrder, orderId).Error; err == nil {
				var customer models.Customer
				if err := s.db.Where("online_id = ?", customerOrder.CustomerOnlineID).First(&customer).Error; err == nil {
					customerMap[customer.ID] = customer
					feedbacks = append(feedbacks, "Matched by CustomerOrder ID")
				}
			}
		}
	}

	if len(customerMap) == 0 {
		return &pb.QueryCustomerResponse{
			Success:  false,
			Feedback: "No customers found",
		}, nil
	}

	for _, customer := range customerMap {
		customers = append(customers, customer)
	}

	var pbCustomers []*pb.Customer
	for _, customer := range customers {
		pbCustomer := &pb.Customer{
			OnlineId: customer.OnlineID,
			Name:     customer.Name,
			Address:  customer.Address,
		}
		pbCustomers = append(pbCustomers, pbCustomer)
	}

	return &pb.QueryCustomerResponse{
		Success:   true,
		Feedback:  fmt.Sprintf("Matched by: %v", feedbacks),
		Customers: pbCustomers,
	}, nil
}

// QueryBook 查询书籍
func (s *OnlineServiceServer) QueryBook(ctx context.Context, req *pb.QueryBookRequest) (*pb.QueryBookResponse, error) {
	var books []models.Book
	input := req.GetInput()
	var feedbacks []string
	bookMap := make(map[int32]models.Book)

	if input != "" {
		// 尝试匹配 book_no
		var bookNoBooks []models.Book
		if err := s.db.Where("book_no LIKE ?", "%"+input+"%").Find(&bookNoBooks).Error; err == nil && len(bookNoBooks) > 0 {
			for _, book := range bookNoBooks {
				if isMatch(input, book.BookNo, s.matchThreshold) {
					bookMap[book.ID] = book
					feedbacks = append(feedbacks, "Matched by book_no")
				}
			}
		}

		// 尝试匹配 title
		var titleBooks []models.Book
		if err := s.db.Where("title LIKE ?", "%"+input+"%").Find(&titleBooks).Error; err == nil && len(titleBooks) > 0 {
			for _, book := range titleBooks {
				if isMatch(input, book.Title, s.matchThreshold) {
					bookMap[book.ID] = book
					feedbacks = append(feedbacks, "Matched by title")
				}
			}
		}

		// 尝试匹配 publisher_name
		var publisherNameBooks []models.Book
		if err := s.db.Where("publisher_name LIKE ?", "%"+input+"%").Find(&publisherNameBooks).Error; err == nil && len(publisherNameBooks) > 0 {
			for _, book := range publisherNameBooks {
				if isMatch(input, book.PublisherName, s.matchThreshold) {
					bookMap[book.ID] = book
					feedbacks = append(feedbacks, "Matched by publisher_name")
				}
			}
		}

		// 尝试匹配 keywords
		var keywordsBooks []models.Book
		if err := s.db.Where("keywords LIKE ?", "%"+input+"%").Find(&keywordsBooks).Error; err == nil && len(keywordsBooks) > 0 {
			for _, book := range keywordsBooks {
				if isKeywordsMatch(input, book.Keywords, s.matchThreshold) {
					bookMap[book.ID] = book
					feedbacks = append(feedbacks, "Matched by keywords")
				}
			}
		}

		// 尝试匹配 authors
		var authorsBooks []models.Book
		if err := s.db.Where("authors LIKE ?", "%"+input+"%").Find(&authorsBooks).Error; err == nil && len(authorsBooks) > 0 {
			for _, book := range authorsBooks {
				if isAuthorsMatch(input, book.Authors, s.matchThreshold) {
					bookMap[book.ID] = book
					feedbacks = append(feedbacks, "Matched by authors")
				}
			}
		}
	}

	if len(bookMap) == 0 {
		return &pb.QueryBookResponse{
			Success:  false,
			Feedback: "No books found",
		}, nil
	}

	for _, book := range bookMap {
		books = append(books, book)
	}

	var pbBooks []*pb.Book
	for _, book := range books {
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

	return &pb.QueryBookResponse{
		Success:  true,
		Feedback: fmt.Sprintf("Matched by: %v", feedbacks),
		Books:    pbBooks,
	}, nil
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
