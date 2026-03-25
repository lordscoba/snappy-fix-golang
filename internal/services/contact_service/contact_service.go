package contactservice

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/snappy-fix-golang/internal/adapters/repository"
	"github.com/snappy-fix-golang/internal/domain/entities"
	"github.com/snappy-fix-golang/internal/inst"
	"gorm.io/gorm"
)

//////////////////////////////////////////////////////
//// CREATE MESSAGE
//////////////////////////////////////////////////////

func CreateMessageService(req entities.CreateContactMessageRequest, c *gin.Context, db *gorm.DB) (gin.H, int, error) {

	pdb := inst.InitDB(db)

	message := entities.ContactMessage{
		Category:  req.Category,
		Name:      req.Name,
		Email:     req.Email,
		Subject:   req.Subject,
		Message:   req.Message,
		IP:        c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
	}

	if err := message.Create(pdb); err != nil {
		return gin.H{}, http.StatusInternalServerError, err
	}

	return gin.H{"message": message}, http.StatusCreated, nil
}

//////////////////////////////////////////////////////
//// GET ALL
//////////////////////////////////////////////////////

func GetAllMessagesService(db *gorm.DB, c *gin.Context) (gin.H, repository.PaginationResponse, int, error) {

	pdb := inst.InitDB(db)

	var msg entities.ContactMessage

	list, pagination, err := msg.GetAll(pdb, c)
	if err != nil {
		return gin.H{}, pagination, http.StatusBadRequest, err
	}

	return gin.H{"messages": list}, pagination, http.StatusOK, nil
}

//////////////////////////////////////////////////////
//// GET BY ID
//////////////////////////////////////////////////////

func GetMessageByIDService(id string, db *gorm.DB) (gin.H, int, error) {

	pdb := inst.InitDB(db)

	var msg entities.ContactMessage

	out, err := msg.GetByID(pdb, id)
	if err != nil {
		return gin.H{}, http.StatusNotFound, err
	}

	return gin.H{"message": out}, http.StatusOK, nil
}

//////////////////////////////////////////////////////
//// MARK AS READ
//////////////////////////////////////////////////////

func MarkAsReadService(id string, db *gorm.DB) (gin.H, int, error) {

	pdb := inst.InitDB(db)

	var msg entities.ContactMessage

	if err := msg.MarkAsRead(pdb, id); err != nil {
		return gin.H{}, http.StatusBadRequest, err
	}

	return gin.H{"message": "marked as read"}, http.StatusOK, nil
}

//////////////////////////////////////////////////////
//// DELETE
//////////////////////////////////////////////////////

func DeleteMessageService(id string, db *gorm.DB) (gin.H, int, error) {

	pdb := inst.InitDB(db)

	var msg entities.ContactMessage

	existing, err := msg.GetByID(pdb, id)
	if err != nil {
		return gin.H{}, http.StatusNotFound, err
	}

	if err := existing.Delete(pdb); err != nil {
		return gin.H{}, http.StatusInternalServerError, err
	}

	return gin.H{"message": "deleted successfully"}, http.StatusOK, nil
}

//////////////////////////////////////////////////////
//// COUNT ALL
//////////////////////////////////////////////////////

func CountAllMessagesService(db *gorm.DB) (gin.H, int, error) {

	pdb := inst.InitDB(db)

	var msg entities.ContactMessage

	count, err := msg.CountAll(pdb)
	if err != nil {
		return gin.H{}, http.StatusInternalServerError, err
	}

	return gin.H{"total": count}, http.StatusOK, nil
}
