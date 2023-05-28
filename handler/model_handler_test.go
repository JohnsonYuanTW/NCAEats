package handler

import (
	"testing"

	"github.com/JohnsonYuanTW/NCAEats/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type MockRestaurantRepository struct {
	mock.Mock
	*models.RestaurantRepository
}

type MockOrderRepository struct {
	mock.Mock
	*models.OrderRepository
}

type MockOrderDetailRepository struct {
	mock.Mock
	*models.OrderDetailRepository
}

type MockMenuItemRepository struct {
	mock.Mock
	*models.MenuItemRepository
}

type MockTemplateHandler struct {
	mock.Mock
	*TemplateHandler
}

func (m *MockRestaurantRepository) GetRestaurantByName(name string) (*models.Restaurant, error) {
	args := m.Called(name)
	restaurant, ok := args.Get(0).(*models.Restaurant)
	if !ok || restaurant == nil {
		return nil, args.Get(1).(error)
	} else {
		return restaurant, nil
	}
}

func TestHandleNewOrder(t *testing.T) {
	// Instantiate mocks
	mockRestaurantRepo := new(MockRestaurantRepository)
	mockOrderRepo := new(MockOrderRepository)
	mockMenuItemRepo := new(MockMenuItemRepository)
	mockTemplateHandler := new(MockTemplateHandler)

	mockRestaurantRepo.On("GetRestaurantByName", "testRestaurant").Return(&models.Restaurant{Name: "testRestaurant", Tel: "0912345678"}, nil)
	mockRestaurantRepo.On("GetRestaurantByName", mock.Anything).Return(nil, gorm.ErrRecordNotFound)

	appHandler := &AppHandler{
		RestaurantRepo: mockRestaurantRepo,
		OrderRepo:      mockOrderRepo,
		MenuItemRepo:   mockMenuItemRepo,
		Templates:      mockTemplateHandler,
		Logger:         logrus.New(),
	}

	t.Run("should return ErrInputError when args is not valid", func(t *testing.T) {
		args := []string{"res1", "res2"}
		ID := "123"
		_, err := appHandler.handleNewOrder(args, ID)
		assert.Equal(t, ErrInputError, err)
	})

	t.Run("should return ErrRestaurantNotFound when restaurant is not found", func(t *testing.T) {
		args := []string{"restaurant1"}
		ID := "123"
		_, err := appHandler.handleNewOrder(args, ID)
		assert.Equal(t, ErrRestaurantNotFound, err)
	})
}
