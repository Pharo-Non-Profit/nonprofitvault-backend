package httptransport

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/rs/cors"

	gateway "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/gateway/httptransport"
	howhear "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/howhear/httptransport"
	objectfile "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/objectfile/httptransport"
	tenant "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/tenant/httptransport"
	user "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/user/httptransport"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/config"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/inputport/httptransport/middleware"
)

type InputPortServer interface {
	Run()
	Shutdown()
}

type httpTransportInputPort struct {
	Config     *config.Conf
	Logger     *slog.Logger
	Server     *http.Server
	Middleware middleware.Middleware
	Tenant     *tenant.Handler
	Gateway    *gateway.Handler
	User       *user.Handler
	HowHear    *howhear.Handler
	ObjectFile *objectfile.Handler
}

func NewInputPort(
	configp *config.Conf,
	loggerp *slog.Logger,
	mid middleware.Middleware,
	org *tenant.Handler,
	gate *gateway.Handler,
	user *user.Handler,
	howhear *howhear.Handler,
	att *objectfile.Handler,
) InputPortServer {
	// Initialize the ServeMux.
	mux := http.NewServeMux()

	// cors.Default() setup the middleware with default options being
	// all origins accepted with simple methods (GET, POST). See
	// documentation via `https://github.com/rs/cors` for more options.
	handler := cors.AllowAll().Handler(mux)

	// Bind the HTTP server to the assigned address and port.
	addr := fmt.Sprintf("%s:%s", configp.AppServer.IP, configp.AppServer.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	// Create our HTTP server controller.
	p := &httpTransportInputPort{
		Config:     configp,
		Logger:     loggerp,
		Middleware: mid,
		Tenant:     org,
		Gateway:    gate,
		User:       user,
		HowHear:    howhear,
		ObjectFile: att,
		Server:     srv,
	}

	// Attach the HTTP server controller to the ServerMux.
	mux.HandleFunc("/", mid.Attach(p.HandleRequests))

	return p
}

func (port *httpTransportInputPort) Run() {
	port.Logger.Info("HTTP server running")
	if err := port.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		port.Logger.Error("listen failed", slog.Any("error", err))

		// DEVELOPERS NOTE: We terminate app here b/c dependency injection not allowed to fail, so fail here at startup of app.
		panic("failed running")
	}
}

func (port *httpTransportInputPort) Shutdown() {
	port.Logger.Info("HTTP server shutdown")
}

func (port *httpTransportInputPort) HandleRequests(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get our URL paths which are slash-seperated.
	ctx := r.Context()
	p := ctx.Value("url_split").([]string)
	n := len(p)
	port.Logger.Debug("Handling request",
		slog.Int("n", n),
		slog.String("m", r.Method),
		slog.Any("p", p),
	)

	switch {
	// --- GATEWAY & PROFILE --- //
	case n == 3 && p[1] == "v1" && p[2] == "health-check" && r.Method == http.MethodGet:
		port.Gateway.HealthCheck(w, r)
	case n == 3 && p[1] == "v1" && p[2] == "version" && r.Method == http.MethodGet:
		port.Gateway.Version(w, r)
	case n == 3 && p[1] == "v1" && p[2] == "greeting" && r.Method == http.MethodPost:
		port.Gateway.Greet(w, r)
	case n == 3 && p[1] == "v1" && p[2] == "login" && r.Method == http.MethodPost:
		port.Gateway.Login(w, r)
	case n == 3 && p[1] == "v1" && p[2] == "register" && r.Method == http.MethodPost:
		port.Gateway.UserRegister(w, r)
	case n == 3 && p[1] == "v1" && p[2] == "refresh-token" && r.Method == http.MethodPost:
		port.Gateway.RefreshToken(w, r)
	// case n == 3 && p[1] == "v1" && p[2] == "verify" && r.Method == http.MethodPost:
	// 	port.Gateway.Verify(w, r)
	case n == 3 && p[1] == "v1" && p[2] == "logout" && r.Method == http.MethodPost:
		port.Gateway.Logout(w, r)
	case n == 3 && p[1] == "v1" && p[2] == "profile" && r.Method == http.MethodGet:
		port.Gateway.Profile(w, r)
	case n == 3 && p[1] == "v1" && p[2] == "profile" && r.Method == http.MethodPut:
		port.Gateway.ProfileUpdate(w, r)
	case n == 4 && p[1] == "v1" && p[2] == "profile" && p[3] == "change-password" && r.Method == http.MethodPut:
		port.Gateway.ProfileChangePassword(w, r)
	case n == 3 && p[1] == "v1" && p[2] == "forgot-password" && r.Method == http.MethodPost:
		port.Gateway.ForgotPassword(w, r)
	case n == 3 && p[1] == "v1" && p[2] == "password-reset" && r.Method == http.MethodPost:
		port.Gateway.PasswordReset(w, r)
		// case n == 3 && p[1] == "v1" && p[2] == "profile" && r.Method == http.MethodGet:
	case n == 3 && p[1] == "v1" && p[2] == "executive-visit-tenant" && r.Method == http.MethodPost:
		port.Gateway.ExecutiveVisitsTenant(w, r)

	// // --- DASHBOARD --- //
	// case n == 3 && p[1] == "v1" && p[2] == "dashboard" && r.Method == http.MethodGet:
	// 	port.Dashboard.Dashboard(w, r)
	// 	// ...

	// --- ORGANIZATION --- //
	case n == 3 && p[1] == "v1" && p[2] == "tenants" && r.Method == http.MethodGet:
		port.Tenant.List(w, r)
	case n == 3 && p[1] == "v1" && p[2] == "tenants" && r.Method == http.MethodPost:
		port.Tenant.Create(w, r)
	case n == 4 && p[1] == "v1" && p[2] == "tenant" && r.Method == http.MethodGet:
		port.Tenant.GetByID(w, r, p[3])
	case n == 4 && p[1] == "v1" && p[2] == "tenant" && r.Method == http.MethodPut:
		port.Tenant.UpdateByID(w, r, p[3])
	case n == 4 && p[1] == "v1" && p[2] == "tenant" && r.Method == http.MethodDelete:
		port.Tenant.DeleteByID(w, r, p[3])
	case n == 5 && p[1] == "v1" && p[2] == "tenants" && p[3] == "operation" && p[4] == "create-comment" && r.Method == http.MethodPost:
		port.Tenant.OperationCreateComment(w, r)
	case n == 4 && p[1] == "v1" && p[2] == "tenants" && p[3] == "select-options" && r.Method == http.MethodGet:
		port.Tenant.ListAsSelectOptionByFilter(w, r)

	// // --- HOW HEAR --- //
	// case n == 3 && p[1] == "v1" && p[2] == "how-hear-about-us-items" && r.Method == http.MethodGet:
	// 	port.HowHear.List(w, r)
	// case n == 3 && p[1] == "v1" && p[2] == "how-hear-about-us-items" && r.Method == http.MethodPost:
	// 	port.HowHear.Create(w, r)
	// case n == 4 && p[1] == "v1" && p[2] == "how-hear-about-us-item" && r.Method == http.MethodGet:
	// 	port.HowHear.GetByID(w, r, p[3])
	// case n == 4 && p[1] == "v1" && p[2] == "how-hear-about-us-item" && r.Method == http.MethodPut:
	// 	port.HowHear.UpdateByID(w, r, p[3])
	// case n == 4 && p[1] == "v1" && p[2] == "how-hear-about-us-item" && r.Method == http.MethodDelete:
	// 	port.HowHear.DeleteByID(w, r, p[3])
	// // case n == 5 && p[1] == "v1" && p[2] == "users" && p[3] == "operation" && p[4] == "create-comment" && r.Method == http.MethodPost:
	// // 	port.Tag.OperationCreateComment(w, r)
	// case n == 4 && p[1] == "v1" && p[2] == "how-hear-about-us-items" && p[3] == "select-options" && r.Method == http.MethodGet:
	// 	port.HowHear.ListAsSelectOptions(w, r)
	case n == 4 && p[1] == "v1" && p[2] == "select-options" && p[3] == "how-hear-about-us-items" && r.Method == http.MethodGet:
		port.HowHear.PublicListAsSelectOptions(w, r)

	// --- USERS --- //
	case n == 3 && p[1] == "v1" && p[2] == "users" && r.Method == http.MethodGet:
		port.User.List(w, r)
	case n == 4 && p[1] == "v1" && p[2] == "users" && p[3] == "count" && r.Method == http.MethodGet:
		port.User.Count(w, r)
	case n == 3 && p[1] == "v1" && p[2] == "users" && r.Method == http.MethodPost:
		port.User.Create(w, r)
	case n == 4 && p[1] == "v1" && p[2] == "user" && r.Method == http.MethodGet:
		port.User.GetByID(w, r, p[3])
	case n == 4 && p[1] == "v1" && p[2] == "user" && r.Method == http.MethodPut:
		port.User.UpdateByID(w, r, p[3])
	case n == 4 && p[1] == "v1" && p[2] == "user" && r.Method == http.MethodDelete:
		port.User.DeleteByID(w, r, p[3])
	case n == 5 && p[1] == "v1" && p[2] == "users" && p[3] == "operation" && p[4] == "create-comment" && r.Method == http.MethodPost:
		port.User.OperationCreateComment(w, r)
	case n == 4 && p[1] == "v1" && p[2] == "users" && p[3] == "select-options" && r.Method == http.MethodGet:
		port.User.ListAsSelectOptions(w, r)

	// --- OBJECT FILES --- //
	case n == 3 && p[1] == "v1" && p[2] == "object-files" && r.Method == http.MethodGet:
		port.ObjectFile.List(w, r)
	case n == 3 && p[1] == "v1" && p[2] == "object-files" && r.Method == http.MethodPost:
		port.ObjectFile.Create(w, r)
	case n == 4 && p[1] == "v1" && p[2] == "object-file" && r.Method == http.MethodGet:
		port.ObjectFile.GetByID(w, r, p[3])
	case n == 4 && p[1] == "v1" && p[2] == "object-file" && r.Method == http.MethodPut:
		port.ObjectFile.UpdateByID(w, r, p[3])
	case n == 4 && p[1] == "v1" && p[2] == "object-file" && r.Method == http.MethodDelete:
		port.ObjectFile.DeleteByID(w, r, p[3])
	case n == 5 && p[1] == "v1" && p[2] == "object-file" && p[4] == "presigned-url" && r.Method == http.MethodGet:
		port.ObjectFile.GetPresignedURLByID(w, r, p[3])
	case n == 5 && p[1] == "v1" && p[2] == "object-file" && p[4] == "content" && r.Method == http.MethodGet:
		port.ObjectFile.GetContentByID(w, r, p[3])

	// --- CATCH ALL: D.N.E. ---
	default:
		http.NotFound(w, r)
	}
}
