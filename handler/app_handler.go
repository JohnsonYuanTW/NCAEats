package handler

import (
	"reflect"

	"github.com/JohnsonYuanTW/NCAEats/config"
	"github.com/JohnsonYuanTW/NCAEats/models"
	"github.com/line/line-bot-sdk-go/v7/linebot"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type DB struct {
	Connection *gorm.DB
}

type AppHandler struct {
	Logger          *logrus.Logger
	Templates       TemplateHandlerInterface
	Config          *config.Config
	Bot             *linebot.Client
	MenuItemRepo    models.MenuItemRepositoryInterface
	OrderRepo       models.OrderRepositoryInterface
	OrderDetailRepo models.OrderDetailRepositoryInterface
	RestaurantRepo  models.RestaurantRepositoryInterface
}

func NewAppHandler(log *logrus.Logger, templates *TemplateHandler, config *config.Config, bot *linebot.Client, db *gorm.DB) (*AppHandler, error) {
	// AppHandler Creation
	baseRepo := &models.BaseRepository{
		DB: db,
	}
	appHandler := &AppHandler{
		Logger:    log,
		Templates: templates,
		Config:    config,
		Bot:       bot,
		MenuItemRepo: &models.MenuItemRepository{
			BaseRepository: baseRepo,
		},
		OrderRepo: &models.OrderRepository{
			BaseRepository: baseRepo,
		},
		OrderDetailRepo: &models.OrderDetailRepository{
			BaseRepository: baseRepo,
		},
		RestaurantRepo: &models.RestaurantRepository{
			BaseRepository: baseRepo,
		},
	}

	if err := appHandler.initRepository(); err != nil {
		return nil, err
	}

	return appHandler, nil
}

func (a *AppHandler) initRepository() error {
	initRepos := []interface {
		Init() error
	}{
		a.MenuItemRepo,
		a.OrderRepo,
		a.OrderDetailRepo,
		a.RestaurantRepo,
	}

	for _, initRepo := range initRepos {
		if err := initRepo.Init(); err != nil {
			a.Logger.WithError(err).Fatalf("Failed to initialize %s", reflect.TypeOf(initRepo).Elem().Name())
			return err
		}
	}
	return nil
}
