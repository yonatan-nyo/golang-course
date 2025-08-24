package services

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryService struct {
	client *cloudinary.Cloudinary
}

func NewCloudinaryService(cloudinaryURL string) (*CloudinaryService, error) {
	if cloudinaryURL == "" {
		return nil, fmt.Errorf("cloudinary URL is not configured")
	}

	cld, err := cloudinary.NewFromURL(cloudinaryURL)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cloudinary: %v", err)
	}

	return &CloudinaryService{
		client: cld,
	}, nil
}

// UploadPDF uploads a PDF file to Cloudinary
func (cs *CloudinaryService) UploadPDF(file *multipart.FileHeader) (string, error) {
	log.Printf("CloudinaryService: Starting to upload PDF file: %s (size: %d bytes)", file.Filename, file.Size)

	// Validate file type
	contentType := file.Header.Get("Content-Type")
	log.Printf("CloudinaryService: PDF file content type: %s", contentType)
	if contentType != "application/pdf" {
		log.Printf("CloudinaryService: Invalid PDF file type: %s", contentType)
		return "", fmt.Errorf("invalid file type: only PDF files are allowed")
	}

	// Validate file size (10MB limit)
	if file.Size > 10*1024*1024 {
		log.Printf("CloudinaryService: PDF file too large: %d bytes", file.Size)
		return "", fmt.Errorf("file size too large: maximum 10MB allowed for PDF files")
	}

	// Open the file
	src, err := file.Open()
	if err != nil {
		log.Printf("CloudinaryService: Failed to open PDF file: %v", err)
		return "", fmt.Errorf("failed to open uploaded file: %v", err)
	}
	defer src.Close()

	// Generate unique public ID
	publicID := fmt.Sprintf("pdfs/%d_%s", time.Now().Unix(), strings.TrimSuffix(file.Filename, ".pdf"))

	// Upload to Cloudinary
	ctx := context.Background()
	useFilename := true
	uniqueFilename := false
	uploadResult, err := cs.client.Upload.Upload(ctx, src, uploader.UploadParams{
		PublicID:       publicID,
		ResourceType:   "raw", // Use 'raw' for non-image files like PDFs
		Folder:         "labpro/pdfs",
		UseFilename:    &useFilename,
		UniqueFilename: &uniqueFilename,
		Type:           "upload",
	})

	if err != nil {
		log.Printf("CloudinaryService: Failed to upload PDF to Cloudinary: %v", err)
		return "", fmt.Errorf("failed to upload PDF: %v", err)
	}

	log.Printf("CloudinaryService: PDF uploaded successfully. URL: %s", uploadResult.SecureURL)
	return uploadResult.SecureURL, nil
}

// UploadVideo uploads a video file to Cloudinary
func (cs *CloudinaryService) UploadVideo(file *multipart.FileHeader) (string, error) {
	log.Printf("CloudinaryService: Starting to upload video file: %s (size: %d bytes)", file.Filename, file.Size)

	// Validate file type
	contentType := file.Header.Get("Content-Type")
	log.Printf("CloudinaryService: Video file content type: %s", contentType)

	validVideoTypes := []string{
		"video/mp4",
		"video/avi",
		"video/mov",
		"video/quicktime",
		"video/x-msvideo",
		"video/webm",
		"video/ogg",
	}

	isValidType := false
	for _, validType := range validVideoTypes {
		if contentType == validType {
			isValidType = true
			break
		}
	}

	if !isValidType {
		log.Printf("CloudinaryService: Invalid video file type: %s", contentType)
		return "", fmt.Errorf("invalid file type: only video files (MP4, AVI, MOV, WebM, OGG) are allowed")
	}

	// Validate file size (100MB limit)
	if file.Size > 100*1024*1024 {
		log.Printf("CloudinaryService: Video file too large: %d bytes", file.Size)
		return "", fmt.Errorf("file size too large: maximum 100MB allowed for video files")
	}

	// Open the file
	src, err := file.Open()
	if err != nil {
		log.Printf("CloudinaryService: Failed to open video file: %v", err)
		return "", fmt.Errorf("failed to open uploaded file: %v", err)
	}
	defer src.Close()

	// Get file extension
	fileExt := ""
	parts := strings.Split(file.Filename, ".")
	if len(parts) > 1 {
		fileExt = parts[len(parts)-1]
	}

	// Generate unique public ID
	publicID := fmt.Sprintf("videos/%d_%s", time.Now().Unix(), strings.TrimSuffix(file.Filename, "."+fileExt))

	// Upload to Cloudinary
	ctx := context.Background()
	useFilename := true
	uniqueFilename := false
	uploadResult, err := cs.client.Upload.Upload(ctx, src, uploader.UploadParams{
		PublicID:       publicID,
		ResourceType:   "video",
		Folder:         "labpro/videos",
		UseFilename:    &useFilename,
		UniqueFilename: &uniqueFilename,
		Type:           "upload", // Explicitly set to 'upload' for public access
	})

	if err != nil {
		log.Printf("CloudinaryService: Failed to upload video to Cloudinary: %v", err)
		return "", fmt.Errorf("failed to upload video: %v", err)
	}

	log.Printf("CloudinaryService: Video uploaded successfully. URL: %s", uploadResult.SecureURL)
	return uploadResult.SecureURL, nil
}

// DeleteFile deletes a file from Cloudinary using its public ID
func (cs *CloudinaryService) DeleteFile(publicID string, resourceType string) error {
	ctx := context.Background()

	destroyParams := uploader.DestroyParams{
		PublicID:     publicID,
		ResourceType: resourceType,
	}

	result, err := cs.client.Upload.Destroy(ctx, destroyParams)
	if err != nil {
		log.Printf("CloudinaryService: Failed to delete file %s: %v", publicID, err)
		return fmt.Errorf("failed to delete file: %v", err)
	}

	log.Printf("CloudinaryService: File deletion result: %s", result.Result)
	return nil
}

// ExtractPublicIDFromURL extracts the public ID from a Cloudinary URL
func (cs *CloudinaryService) ExtractPublicIDFromURL(url string) string {
	// Example URL: https://res.cloudinary.com/cloud_name/resource_type/upload/v1234567890/folder/public_id.ext
	parts := strings.Split(url, "/")
	if len(parts) < 2 {
		return ""
	}

	// Get the last part (filename with extension)
	filename := parts[len(parts)-1]

	// Remove version if present (v1234567890)
	if len(parts) >= 2 && strings.HasPrefix(parts[len(parts)-2], "v") {
		// If there's a version, the public ID might include folder structure
		publicIDParts := parts[len(parts)-2:]
		if strings.HasPrefix(publicIDParts[0], "v") {
			return strings.Join(publicIDParts[1:], "/")
		}
	}

	return filename
}
