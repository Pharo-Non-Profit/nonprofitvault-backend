//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"

	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/config"

	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/adapter/cache/mongodbcache"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/adapter/emailer/mailgun"
	object_storage "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/adapter/storage/object"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/adapter/templatedemailer"

	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/provider/jwt"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/provider/kmutex"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/provider/logger"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/provider/mongodb"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/provider/password"

	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/provider/time"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/provider/uuid"

	ds_howhear "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/howhear/datastore"
	ds_objectfile "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/objectfile/datastore"
	ds_shareablelink "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/shareablelink/datastore"
	ds_smartfolder "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/smartfolder/datastore"
	ds_tenant "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/tenant/datastore"
	ds_user "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/user/datastore"

	uc_gateway "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/gateway/controller"
	uc_howhear "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/howhear/controller"
	uc_objectfile "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/objectfile/controller"
	uc_shareablelink "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/shareablelink/controller"
	uc_smartfolder "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/smartfolder/controller"
	uc_tenant "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/tenant/controller"
	uc_user "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/user/controller"

	http_gate "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/gateway/httptransport"
	http_howhear "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/howhear/httptransport"
	http_objectfile "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/objectfile/httptransport"
	http_shareablelink "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/shareablelink/httptransport"
	http_smartfolder "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/smartfolder/httptransport"
	http_tenant "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/tenant/httptransport"
	http_user "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/user/httptransport"

	http "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/inputport/httptransport"
	http_middleware "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/inputport/httptransport/middleware"
)

func InitializeEvent() Application {
	// Our application is dependent on the following Golang packages. We need to
	// provide them to Google wire so it can sort out the dependency injection
	// at compile time.
	wire.Build(
		// CONFIGURATION SECTION
		config.New,

		// PROVIDERS SECTION
		logger.NewProvider,
		uuid.NewProvider,
		time.NewProvider,
		jwt.NewProvider,
		password.NewProvider,
		kmutex.NewProvider,
		mongodb.NewProvider,

		// TODO
		mailgun.NewEmailer,
		templatedemailer.NewTemplatedEmailer,
		mongodbcache.NewCache,
		object_storage.NewStorage,

		// ADAPTERS SECTION

		// DATASTORE
		ds_tenant.NewDatastore,
		ds_user.NewDatastore,
		ds_howhear.NewDatastore,
		ds_objectfile.NewDatastore,
		ds_smartfolder.NewDatastore,
		ds_shareablelink.NewDatastore,

		// USECASE
		uc_tenant.NewController,
		uc_gateway.NewController,
		uc_user.NewController,
		uc_howhear.NewController,
		uc_objectfile.NewController,
		uc_smartfolder.NewController,
		uc_shareablelink.NewController,

		// HTTP TRANSPORT SECTION
		http_tenant.NewHandler,
		http_gate.NewHandler,
		http_user.NewHandler,
		http_howhear.NewHandler,
		http_objectfile.NewHandler,
		http_smartfolder.NewHandler,
		http_shareablelink.NewHandler,

		// INPUT PORT SECTION
		http_middleware.NewMiddleware,
		http.NewInputPort,

		// APP
		NewApplication)
	return Application{}
}
