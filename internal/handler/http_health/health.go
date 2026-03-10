package httphealth

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/snappy-fix-golang/external/request"
	"github.com/snappy-fix-golang/internal/adapters/db"
	"github.com/snappy-fix-golang/internal/domain/entities"
	pingservices "github.com/snappy-fix-golang/internal/services/ping_services"
	logutil "github.com/snappy-fix-golang/pkg/logger"
	"github.com/snappy-fix-golang/pkg/utils/responses"
)

type Controller struct {
	Db        *db.Database
	Validator *validator.Validate
	Logger    *logutil.Logger
	ExtReq    request.ExternalRequest
}

// Post godoc
//
//	@Summary		Ping the health endpoint
//	@Description	Responds with a success message
//	@Tags			health
//	@Accept			json
//	@Produce		json
//	@Param			request	body		entities.Ping	true	"Ping Request"
//	@Success		200			{object}	map[string]interface{}	"ping successful"
//	@Failure		400			{object}	map[string]interface{}	"Bad Request"
//	@Failure		422			{object}	map[string]interface{}	"Validation Failed"
//	@Router			/health [post]
//	@securityDefinitions.apikey  ApiKeyAuth
//	@example			Ping Request Example
//
//	{
//		"message": "Hello World"
//	}
func (base *Controller) Post(c *gin.Context) {
	var (
		req = entities.Ping{}
	)

	err := c.ShouldBind(&req)
	if err != nil {
		rd := responses.BuildErrorResponse(http.StatusBadRequest, "error", "Failed to parse request body", err, nil)
		c.JSON(http.StatusBadRequest, rd)
		return
	}

	err = base.Validator.Struct(&req)
	if err != nil {
		rd := responses.BuildErrorResponse(http.StatusBadRequest, "error", "Validation failed", responses.ValidationResponse(err, base.Validator), nil)
		c.JSON(http.StatusBadRequest, rd)
		return
	}

	if !pingservices.ReturnTrue() {
		rd := responses.BuildErrorResponse(http.StatusInternalServerError, "error", "ping failed", fmt.Errorf("ping failed"), nil)
		c.JSON(http.StatusBadRequest, rd)
		return
	}

	base.Logger.Info("ping successfull")
	rd := responses.BuildSuccessResponse(http.StatusOK, req.Message, nil)

	c.JSON(http.StatusOK, rd)

}

// Get godoc
//
//	@Summary		Get health status
//	@Description	Returns the server health status
//	@Tags			health
//	@Accept			json
//	@Produce		json
//	@Success		200		{object}	entities.Ping	"ping successful"
//	@Failure		500		{object}	map[string]interface{}	"Internal Server Error"
//	@Router			/health [get]
//	@securityDefinitions.apikey  ApiKeyAuth
//	@example			response
//
//	{
//		"message": "Hello World"
//	}
func (base *Controller) Get(c *gin.Context) {
	if !pingservices.ReturnTrue() {
		rd := responses.BuildErrorResponse(http.StatusInternalServerError, "error", "ping failed", fmt.Errorf("ping failed"), nil)
		c.JSON(http.StatusInternalServerError, rd)
		return
	}

	rd := responses.BuildSuccessResponse(http.StatusOK, "ping successful", entities.Ping{Message: "Hello World"})

	// if data, err := json.MarshalIndent(rd.Data, "", "  "); err == nil {
	// 	fmt.Println(string(data))
	// } else {
	// 	fmt.Printf("failed to marshal response: %v\n", err)
	// }

	c.JSON(http.StatusOK, rd.Data)
}

// GetIPGeo godoc
//
//	@Summary		Resolve IP Geolocation
//	@Description	Resolves geolocation details (country, region, city, lat/lon) for a given IP address.
//	@Tags			health
//	@Accept			json
//	@Produce		json
//	@Param			ip	query		string	false	"IP address to resolve (if not provided, the caller's IP is used)"	example(8.8.8.8)
//	@Success		200	{object}	map[string]interface{}	"ip geolocation resolved"
//	@Failure		400	{object}	map[string]interface{}	"Invalid IP address provided"
//	@Failure		502	{object}	map[string]interface{}	"Failed to resolve IP geolocation"
//	@Router			/health/ip-geo [get]
//	@securityDefinitions.apikey  ApiKeyAuth
//	@example response
//	{
//		"status": "success",
//		"message": "ip geolocation resolved",
//		"data": {
//			"ip": "8.8.8.8",
//			"type": "ipv4",
//			"continent_code": "NA",
//			"continent_name": "North America",
//			"country_code": "US",
//			"country_name": "United States",
//			"region_code": "CA",
//			"region_name": "California",
//			"city": "Mountain View",
//			"zip": "94043",
//			"latitude": 37.4056,
//			"longitude": -122.0775
//		}
//	}
func (base *Controller) GetIPGeo(c *gin.Context) {
	// Prefer explicit ip in query; else use client's IP from gin.
	ip := c.Query("ip")
	if ip == "" {
		ip = c.ClientIP()
	}

	svc := pingservices.Service{ExtReq: base.ExtReq}
	info, err := svc.ResolveIP(ip)
	if err != nil {
		base.Logger.Error("ip geo resolve failed", "ip", ip, "err", err.Error())
		rd := responses.BuildErrorResponse(http.StatusBadGateway, "error", "failed to resolve IP geolocation", err, nil)
		c.JSON(http.StatusBadGateway, rd)
		return
	}

	rd := responses.BuildSuccessResponse(http.StatusOK, "ip geolocation resolved", info)
	c.JSON(http.StatusOK, rd)
}
