package dto

type FileUploadV1ResponseDTO struct {
	FileName string `json:"fileName"`
	Message  string `json:"message"`
}

type FileDeleteV1RequestDTO struct {
	FileName   string `json:"fileName" binding:"required"`
	BucketName string `json:"bucketName" binding:"required"`
}

type FileDeleteV1ResponseDTO struct {
	Message string `json:"message"`
}

type FileContentV1RequestDTO struct {
	FileName   string `json:"fileName" binding:"required"`
	BucketName string `json:"bucketName" binding:"required"`
}

type FileContentV1ResponseDTO struct {
	Content string `json:"content"`
}

type FileURLV1RequestDTO struct {
	FileName   string `json:"fileName" binding:"required"`
	BucketName string `json:"bucketName" binding:"required"`
}

type FileURLV1ResponseDTO struct {
	URL string `json:"url"`
}

type FileGetV1RequestDTO struct {
	FileName   string `json:"fileName" binding:"required"`
	BucketName string `json:"bucketName" binding:"required"`
}
