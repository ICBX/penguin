package rest

import (
	"errors"
	"github.com/ICBX/penguin/pkg/common"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"gorm.io/gorm"
	"strconv"
)

func (s *Server) routeBlobberPull(ctx *fiber.Ctx) (err error) {
	// get blobber id from route
	blobberID := utils.CopyString(ctx.Params("id"))
	if blobberID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Blobber ID is required")
	}
	var blobberIDUint uint
	if blobberIDUint, err = convertStringToUint(blobberID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	// get blobber secret from headers
	blobberSecret := ctx.Get("Blobber-Secret")
	if blobberSecret == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "no blobberID specified")
	}

	// check if blobber id exists and secret is correct
	if err = s.db.Where(&common.BlobDownloader{
		ID:     blobberIDUint,
		Secret: blobberSecret,
	}).First(&common.BlobDownloader{}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid blobberID or secret")
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// return a list of videos to download
	var queues []*common.Queue
	if err = s.db.Where(&common.Queue{
		BlobberID: blobberIDUint,
	}).Find(&queues).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// collect video ids
	var videoIDS = make([]string, len(queues))
	for i, q := range queues {
		videoIDS[i] = q.VideoID
	}

	return ctx.Status(fiber.StatusOK).JSON(videoIDS)
}

// TODO: move to util
func convertStringToUint(s string) (uint, error) {
	u, err := strconv.ParseUint(s, 10, 0)
	if err != nil {
		return 0, err
	}
	return uint(u), nil
}
