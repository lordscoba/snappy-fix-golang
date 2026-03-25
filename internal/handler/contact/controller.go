package contact

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/snappy-fix-golang/internal/adapters/db"
	"github.com/snappy-fix-golang/internal/domain/entities"
	contactservice "github.com/snappy-fix-golang/internal/services/contact_service"
	"github.com/snappy-fix-golang/pkg/utils/responses"
)

type Controller struct {
	Db        *db.Database
	Validator *validator.Validate
}

//////////////////////////////////////////////////////
//// CREATE
//////////////////////////////////////////////////////

func (base *Controller) CreateMessage(c *gin.Context) {

	var req entities.CreateContactMessageRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		rd := responses.BuildErrorResponse(http.StatusBadRequest, "error", "invalid request body", err.Error(), nil)
		c.JSON(http.StatusBadRequest, rd)
		return
	}

	data, code, err := contactservice.CreateMessageService(req, c, base.Db.Postgresql.DB())
	if err != nil {
		rd := responses.BuildErrorResponse(code, "error", err.Error(), err, nil)
		c.JSON(code, rd)
		return
	}

	rd := responses.BuildSuccessResponse(code, "success", data, nil, code)
	c.JSON(code, rd)
}

//////////////////////////////////////////////////////
//// GET ALL
//////////////////////////////////////////////////////

func (base *Controller) GetAllMessages(c *gin.Context) {

	data, pagination, code, err := contactservice.GetAllMessagesService(base.Db.Postgresql.DB(), c)

	if err != nil {
		rd := responses.BuildErrorResponse(code, "error", err.Error(), err, pagination)
		c.JSON(code, rd)
		return
	}

	rd := responses.BuildSuccessResponse(http.StatusOK, "success", data, pagination, code)
	c.JSON(http.StatusOK, rd)
}

//////////////////////////////////////////////////////
//// GET BY ID
//////////////////////////////////////////////////////

func (base *Controller) GetMessageByID(c *gin.Context) {

	id := c.Param("id")

	data, code, err := contactservice.GetMessageByIDService(id, base.Db.Postgresql.DB())

	if err != nil {
		rd := responses.BuildErrorResponse(code, "error", err.Error(), err, nil)
		c.JSON(code, rd)
		return
	}

	rd := responses.BuildSuccessResponse(http.StatusOK, "success", data, nil, code)
	c.JSON(http.StatusOK, rd)
}

//////////////////////////////////////////////////////
//// MARK AS READ
//////////////////////////////////////////////////////

func (base *Controller) MarkAsRead(c *gin.Context) {

	id := c.Param("id")

	data, code, err := contactservice.MarkAsReadService(id, base.Db.Postgresql.DB())

	if err != nil {
		rd := responses.BuildErrorResponse(code, "error", err.Error(), err, nil)
		c.JSON(code, rd)
		return
	}

	rd := responses.BuildSuccessResponse(http.StatusOK, "success", data, nil, code)
	c.JSON(http.StatusOK, rd)
}

//////////////////////////////////////////////////////
//// DELETE
//////////////////////////////////////////////////////

func (base *Controller) DeleteMessage(c *gin.Context) {

	id := c.Param("id")

	data, code, err := contactservice.DeleteMessageService(id, base.Db.Postgresql.DB())

	if err != nil {
		rd := responses.BuildErrorResponse(code, "error", err.Error(), err, nil)
		c.JSON(code, rd)
		return
	}

	rd := responses.BuildSuccessResponse(http.StatusOK, "success", data, nil, code)
	c.JSON(http.StatusOK, rd)
}

//////////////////////////////////////////////////////
//// COUNT
//////////////////////////////////////////////////////

func (base *Controller) CountAllMessages(c *gin.Context) {

	data, code, err := contactservice.CountAllMessagesService(base.Db.Postgresql.DB())

	if err != nil {
		rd := responses.BuildErrorResponse(code, "error", err.Error(), err, nil)
		c.JSON(code, rd)
		return
	}

	rd := responses.BuildSuccessResponse(http.StatusOK, "success", data, nil, code)
	c.JSON(http.StatusOK, rd)
}
