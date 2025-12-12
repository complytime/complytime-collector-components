package server

import (
	"crypto/tls"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	middleware "github.com/oapi-codegen/gin-middleware"

	"github.com/complytime/complybeacon/compass/api"
	httpmw "github.com/complytime/complybeacon/compass/internal/middleware"
	compass "github.com/complytime/complybeacon/compass/service"
)

func NewGinServer(service *compass.Service, port string) *http.Server {
	swagger, err := api.GetSwagger()
	if err != nil {
		slog.Error("Error loading swagger spec", "err", err)
		os.Exit(1)
	}

	// Clear out the servers array in the swagger spec, that skips validating
	// that server names match. We don't know how this thing will be run.
	swagger.Servers = nil

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(requestid.New(), httpmw.AccessLogger())

	r.Use(middleware.OapiRequestValidator(swagger))

	api.RegisterHandlers(r, service)

	s := &http.Server{
		Handler:           r,
		Addr:              net.JoinHostPort("0.0.0.0", port),
		ReadHeaderTimeout: 10 * time.Second,
	}

	return s
}

func SetupTLS(server *http.Server, config Config) (string, string) {
	// TODO: Allow loosening here through configuration
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS13}
	server.TLSConfig = tlsConfig

	if config.Certificate.PublicKey == "" {
		slog.Error("Invalid certification configuration. Please add certConfig.cert to the configuration.")
		os.Exit(1)
	}

	if config.Certificate.PrivateKey == "" {
		slog.Error("Invalid certification configuration. Please add certConfig.key to the configuration.")
		os.Exit(1)
	}

	return config.Certificate.PublicKey, config.Certificate.PrivateKey
}
