package messagemodel

// VideoInfo represents the information about a newly uploaded video
type VideoInfo struct {
	Timestamp   int64  `json:"timestamp"`
	RawVidS3Key string `json:"s3key"`
	UploadedBy  string `json:"uploaded_by"`
	CourseId    string `json:"course_id"`
	VideoId     string `json:"video_id"`
}
