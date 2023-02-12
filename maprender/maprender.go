package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/hsmtkk/qiita-cloud-run-grpc-acl/proto"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	portStr := os.Getenv("PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("failed to parse string as int; %s; %v", portStr, err)
	}

	googleMapAPIKey := os.Getenv("GOOGLE_MAP_API_KEY")
	locationProviderURI := os.Getenv("LOCATION_PROVIDER_URI")
	parsed, err := url.Parse(locationProviderURI)
	if err != nil {
		log.Fatalf("failed to parse location provider URI; %s; %v", locationProviderURI, err)
	}
	locationProviderAddress := fmt.Sprintf("%s:%d", parsed.Host, 443)

	gRPCConn, err := gRPCConnect(locationProviderAddress)
	if err != nil {
		log.Fatal(err)
	}
	defer gRPCConn.Close()

	locationClient := proto.NewLocationServiceClient(gRPCConn)

	handler := newHandler(locationClient, googleMapAPIKey)

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/", handler.Handle)

	// Start server
	if err := e.Start(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatal(err)
	}
}

func gRPCConnect(locationProviderAddress string) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{grpc.WithAuthority(locationProviderAddress)}
	systemRoots, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("failed to get system cert; %w", err)
	}
	cred := credentials.NewTLS(&tls.Config{
		RootCAs: systemRoots,
	})
	opts = append(opts, grpc.WithTransportCredentials(cred))
	gRPCConn, err := grpc.Dial(locationProviderAddress, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect location provider with gRPC; %w", err)
	}
	return gRPCConn, nil
}

type handler struct {
	locationClient  proto.LocationServiceClient
	googleMAPAPIKey string
}

func newHandler(locationClient proto.LocationServiceClient, googleMapAPIKey string) *handler {
	return &handler{locationClient, googleMapAPIKey}
}

func (h *handler) Handle(ectx echo.Context) error {
	resp, err := h.locationClient.GetLocation(ectx.Request().Context(), &proto.LocationRequest{})
	if err != nil {
		return fmt.Errorf("gRPC request failed; %w", err)
	}
	longitude := resp.GetLongitude()
	latitude := resp.GetLatitude()
	html := fmt.Sprintf(htmlTemplate, longitude, latitude)
	return ectx.HTML(http.StatusOK, html)
}

const htmlTemplate = `<iframe
width="600"
height="600"
frameborder="0" style="border:0"
referrerpolicy="no-referrer-when-downgrade"
src="https://www.google.com/maps/embed/v1/view?key=YOUR_API_KEY&center=%d,%d"
allowfullscreen>
</iframe>
`
