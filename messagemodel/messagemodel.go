package messagemodel

// VideoInfo represents the information about a newly uploaded video
type VideoInfo struct {
	Timestamp   int64  `json:"timestamp"`
	RawVidS3Key string `json:"s3key"`
	CourseSlug  string `json:"course_slug"`
	VideoSlug   string `json:"video_slug"`
	UserEmail   string `json:"user_email"`
}

type ProcessedSegmentsInfo struct {
	CourseSlug     string `json:"course_slug"`
	VideoSlug      string `json:"video_slug"`
	UserEmail      string `json:"user_email"`
	LocalOutputDir string `json:"local_output_dir"`
}
