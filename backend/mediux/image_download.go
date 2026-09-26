package mediux

import (
	"aura/config"
	"aura/logging"
	"aura/utils"
	"context"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func GetImage(ctx context.Context, assetID string, formatDate string, imageQuality ImageQuality) (imageData []byte, imageType string, Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("MediUX: Getting %s Image for Asset ID '%s'",
		cases.Title(language.English).String(string(imageQuality)), assetID), logging.LevelDebug)
	defer logAction.Complete()

	imageData = nil
	imageType = ""
	Err = logging.LogErrorInfo{}

	// Determine the folder path based on qualityParam
	var folderPath string
	switch imageQuality {
	case ImageQualityOriginal:
		folderPath = FullTempImageFolder
	case ImageQualityOptimized, ImageQualityThumb:
		folderPath = ThumbsTempImageFolder
	}

	// File Name Base (extension is determined by the actual image content type,
	// since MediUX images may be JPEG, PNG, WEBP, or GIF)
	fileNameBase := fmt.Sprintf("%s_%s", assetID, formatDate)
	isNewDownload := false

	// Run this after the function completes
	// Handle Cache Images Setting
	defer func() {
		// If there is an error or no image data, do not attempt to cache
		if Err.Message != "" || len(imageData) == 0 {
			return
		}

		if !isNewDownload {
			// If the image was not newly downloaded, no need to cache
			return
		}

		// Check if caching is enabled
		if config.Current.Images.CacheImages.Enabled {
			// Create a new logging data for this goroutine
			ctx, ld := logging.CreateLoggingContext(context.Background(), "Caching - MediUX Image")
			logAction := ld.AddAction("Caching MediUX Image", logging.LevelDebug)
			ctx = logging.WithCurrentAction(ctx, logAction)
			// Check if the folder exists
			Err = utils.CreateFolderIfNotExists(ctx, folderPath)
			if Err.Message != "" {
				return
			}

			writeToFileAction := logAction.AddSubAction("Write Image to Temp Folder", logging.LevelTrace)
			fileName := fileNameBase + utils.GetExtensionFromContentType(imageType)
			filePath := path.Join(folderPath, fileName)
			err := os.WriteFile(filePath, imageData, 0644)
			if err != nil {
				logAction.SetError("Failed to write image to MediUX thumbs temp image folder",
					"Ensure the application has write permissions to the temp image folder.",
					map[string]any{
						"filePath": filePath,
						"error":    err.Error(),
					})

				return
			}
			logAction.AppendResult("filePath", filePath)
			writeToFileAction.Complete()
			logAction.Complete()
			ld.Log()
		}
	}()

	// Now we check to see if config.Images.CacheImages.Enabled is true
	// If it is, we check the temporary or full image folder for the image based on qualityParam
	// If the image exists there, we serve it from disk
	// If not, we fetch it from MediUX and save it to the temp/full folder based on qualityParam
	// If config.Images.CacheImages.Enabled is false, we always fetch from MediUX

	if config.Current.Images.CacheImages.Enabled {
		// Check if the folder exists, if not create it
		Err := utils.CreateFolderIfNotExists(ctx, folderPath)
		if Err.Message != "" {
			// If the folder creation fails, we log the error but continue to fetch the image from MediUX
			logging.LOGGER.Warn().Timestamp().Str("folder_path", folderPath).Str("error", Err.Message).Msg("Failed to create MediUX image folder, will attempt to fetch image from MediUX")
		} else {
			// Check if a cached file exists for this asset, regardless of its extension,
			// since the extension depends on the actual image content type (jpg/png/webp/gif)
			matches, _ := filepath.Glob(path.Join(folderPath, fileNameBase+".*"))
			if len(matches) > 0 {
				filePath := matches[0]
				// Read the image data from disk
				imageData, err := os.ReadFile(filePath)
				if err != nil {
					logging.LOGGER.Warn().Timestamp().Str("file_path", filePath).Str("error", err.Error()).Msg("Failed to read cached MediUX image, will attempt to fetch image from MediUX")
				} else {
					// Successfully read the image from disk
					imageType = "image/jpeg" // Assuming JPEG by default
					switch path.Ext(filePath) {
					case ".png", ".PNG":
						imageType = "image/png"
					case ".gif", ".GIF":
						imageType = "image/gif"
					case ".webp", ".WEBP":
						imageType = "image/webp"
					}
					logAction.AppendResult("filePath", filePath)
					logAction.AppendResult("size", len(imageData))
					logAction.AppendResult("imageType", imageType)
					logAction.AppendResult("source", "cache")
					return imageData, imageType, Err
				}
			}
		}
	}

	// If we reach here, we need to fetch the image from MediUX
	// It may be because caching is disabled, the file doesn't exist, or reading the cached file failed

	// Construct the MediUX URL
	mediuxURL, Err := ConstructImageUrl(ctx, assetID, formatDate, imageQuality)
	if Err.Message != "" {
		return imageData, imageType, Err
	}

	// Make the HTTP Request to MediUX
	resp, respBody, Err := makeRequest(ctx, mediuxURL, "GET", nil, "", false)
	if Err.Message != "" {
		logAction.SetErrorFromInfo(Err)
		return imageData, imageType, *logAction.Error
	}
	defer resp.Body.Close()

	// Check if response has image content
	if len(respBody) == 0 {
		logAction.SetError("MediUX returned an empty image response",
			"Ensure the asset ID is correct and the image exists on the MediUX server.",
			map[string]any{
				"URL": mediuxURL,
			})
		return imageData, imageType, *logAction.Error
	}

	// Get the Content-Type from response headers
	respContentType := resp.Header.Get("Content-Type")
	if respContentType == "" {
		logAction.SetError("MediUX response missing Content-Type header",
			"Ensure MediUX is returning valid image data.",
			map[string]any{
				"URL": mediuxURL,
			})
		return imageData, imageType, *logAction.Error
	}
	imageType = respContentType
	imageData = respBody

	isNewDownload = true
	logAction.AppendResult("size", len(imageData))
	logAction.AppendResult("imageType", respContentType)
	logAction.AppendResult("source", "MediUX")

	return imageData, imageType, Err
}

func GetAvatarImage(ctx context.Context, avatarID string) (imageData []byte, imageType string, Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("MediUX: Getting Avatar Image for Avatar ID '%s'", avatarID), logging.LevelDebug)
	defer logAction.Complete()

	imageData = nil
	imageType = ""
	Err = logging.LogErrorInfo{}

	// Construct the URL for the MediUX API request
	u, err := url.Parse(MediuxApiURL)
	if err != nil {
		logAction.SetError("Failed to parse base URL", "Ensure the URL is valid", map[string]any{"error": err.Error()})
		return imageData, imageType, *logAction.Error
	}
	u.Path = path.Join(u.Path, "assets", avatarID)
	URL := u.String()

	// Make the HTTP Request to MediUX
	resp, respBody, Err := makeRequest(ctx, URL, "GET", nil, "", false)
	if Err.Message != "" {
		logAction.SetErrorFromInfo(Err)
		return imageData, imageType, *logAction.Error
	}
	defer resp.Body.Close()

	// Check if response has image content
	if len(respBody) == 0 {
		logAction.SetError("MediUX returned an empty avatar image response",
			"Ensure the avatar ID is correct and the image exists on the MediUX server.",
			map[string]any{
				"URL": URL,
			})
		return imageData, imageType, *logAction.Error
	}

	// Get the Content-Type from response headers
	respContentType := resp.Header.Get("Content-Type")
	if respContentType == "" {
		logAction.SetError("MediUX avatar response missing Content-Type header",
			"Ensure MediUX is returning valid image data.",
			map[string]any{
				"URL": URL,
			})
		return imageData, imageType, *logAction.Error
	}
	imageType = respContentType
	imageData = respBody

	logAction.AppendResult("size", len(imageData))
	logAction.AppendResult("imageType", respContentType)

	return imageData, imageType, Err
}
