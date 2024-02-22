// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/adapter/cache/mongodbcache"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/adapter/emailer/mailgun"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/adapter/storage/object"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/adapter/templatedemailer"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/gateway/controller"
	httptransport2 "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/gateway/httptransport"
	controller4 "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/howhear/controller"
	datastore3 "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/howhear/datastore"
	httptransport4 "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/howhear/httptransport"
	controller5 "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/objectfile/controller"
	datastore4 "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/objectfile/datastore"
	httptransport5 "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/objectfile/httptransport"
	controller6 "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/smartfolder/controller"
	datastore5 "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/smartfolder/datastore"
	httptransport6 "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/smartfolder/httptransport"
	controller2 "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/tenant/controller"
	datastore2 "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/tenant/datastore"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/tenant/httptransport"
	controller3 "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/user/controller"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/user/datastore"
	httptransport3 "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/user/httptransport"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/config"
	httptransport7 "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/inputport/httptransport"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/inputport/httptransport/middleware"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/provider/jwt"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/provider/kmutex"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/provider/logger"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/provider/mongodb"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/provider/password"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/provider/time"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/provider/uuid"
)

import (
	_ "go.uber.org/automaxprocs"
	_ "time/tzdata"
)

// Injectors from wire.go:

func InitializeEvent() Application {
	slogLogger := logger.NewProvider()
	conf := config.New()
	provider := uuid.NewProvider()
	timeProvider := time.NewProvider()
	jwtProvider := jwt.NewProvider(conf)
	passwordProvider := password.NewProvider()
	kmutexProvider := kmutex.NewProvider()
	client := mongodb.NewProvider(conf, slogLogger)
	cacher := mongodbcache.NewCache(conf, slogLogger, client)
	emailer := mailgun.NewEmailer(conf, slogLogger, provider)
	templatedEmailer := templatedemailer.NewTemplatedEmailer(conf, slogLogger, provider, emailer)
	userStorer := datastore.NewDatastore(conf, slogLogger, client)
	tenantStorer := datastore2.NewDatastore(conf, slogLogger, client)
	howHearAboutUsItemStorer := datastore3.NewDatastore(conf, slogLogger, client)
	gatewayController := controller.NewController(conf, slogLogger, provider, jwtProvider, passwordProvider, kmutexProvider, cacher, templatedEmailer, client, userStorer, tenantStorer, howHearAboutUsItemStorer)
	middlewareMiddleware := middleware.NewMiddleware(conf, slogLogger, provider, timeProvider, jwtProvider, gatewayController)
	objectStorager := object.NewStorage(conf, slogLogger, provider)
	tenantController := controller2.NewController(conf, slogLogger, provider, kmutexProvider, objectStorager, emailer, client, tenantStorer)
	handler := httptransport.NewHandler(slogLogger, tenantController)
	httptransportHandler := httptransport2.NewHandler(slogLogger, gatewayController)
	userController := controller3.NewController(conf, slogLogger, provider, passwordProvider, kmutexProvider, client, tenantStorer, userStorer, templatedEmailer)
	handler2 := httptransport3.NewHandler(slogLogger, userController)
	howHearAboutUsItemController := controller4.NewController(conf, slogLogger, provider, objectStorager, passwordProvider, kmutexProvider, templatedEmailer, client, userStorer, howHearAboutUsItemStorer)
	handler3 := httptransport4.NewHandler(slogLogger, howHearAboutUsItemController)
	objectFileStorer := datastore4.NewDatastore(conf, slogLogger, client)
	objectFileController := controller5.NewController(conf, slogLogger, provider, objectStorager, client, emailer, objectFileStorer, userStorer)
	handler4 := httptransport5.NewHandler(slogLogger, objectFileController)
	smartFolderStorer := datastore5.NewDatastore(conf, slogLogger, client)
	smartFolderController := controller6.NewController(conf, slogLogger, provider, objectStorager, passwordProvider, kmutexProvider, templatedEmailer, client, userStorer, smartFolderStorer)
	handler5 := httptransport6.NewHandler(slogLogger, smartFolderController)
	inputPortServer := httptransport7.NewInputPort(conf, slogLogger, middlewareMiddleware, handler, httptransportHandler, handler2, handler3, handler4, handler5)
	application := NewApplication(slogLogger, inputPortServer)
	return application
}
