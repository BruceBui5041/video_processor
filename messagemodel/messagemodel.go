package messagemodel

// VideoInfo represents the information about a newly uploaded video
type VideoInfo struct {
	VideoID     string `json:"video_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	UploadedBy  string `json:"uploaded_by"`
	Timestamp   int64  `json:"timestamp"`
	S3Key       string `json:"s3key"`
}

type ProcessedSegmentsInfo struct {
	VideoName      string `json:"video_name"`
	LocalOutputDir string `json:"local_output_dir"`
}
