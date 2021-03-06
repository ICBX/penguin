package rest

import (
	"errors"
	"github.com/ICBX/penguin/pkg/common"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"gorm.io/gorm"
)

func (s *Server) routeVideoRemoveBlobber(ctx *fiber.Ctx) (err error) {

	videoID := utils.CopyString(ctx.Params(VideoIDKey))
	blobberID := utils.CopyString(ctx.Params(BlobberIDKey))

	// convert blobberID to int
	blobberIDU, err := convertStringToUint(blobberID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Could not process blobber id")
	}

	// find corresponding video
	// and check if it exists
	var video common.Video
	if err = s.db.Where(&common.Video{ID: videoID}).First(&video).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Did not find video")
	}

	// find corresponding blobber
	// and check if it exists
	var blobber common.BlobDownloader
	if err = s.db.Where(common.BlobDownloader{ID: blobberIDU}).First(&blobber).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Did not find blobber")
	}

	// remove blobber from all corresponding videos
	if err = s.db.Model(&video).Association("Blobbers").Delete(&blobber); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Could not remove blobber from video: "+err.Error())
	}

	// remove BlobLocation for blobberID and videoID
	if err = s.db.Where(&common.BlobLocation{
		VideoID:          videoID,
		BlobDownloaderID: blobberIDU,
	}).Delete(&common.BlobLocation{}).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
	}

	// remove queue entry if exists
	if err = s.db.Where(&common.Queue{
		VideoID:   videoID,
		BlobberID: blobberIDU,
		Action:    common.GetBlob,
	}).Delete(&common.Queue{}).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
	}

	// add video to blobber 'remove' queue
	if err = s.db.Create(&common.Queue{
		VideoID:   videoID,
		BlobberID: blobberIDU,
		Action:    common.RemoveBlob,
	}).Error; err != nil {
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}

	return ctx.Status(fiber.StatusCreated).SendString("blobber removed from video")
}
