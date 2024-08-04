package messagemodel

type ProcessedSegmentsInfo struct {
	UploadedBy     string `json:"uploaded_by"`
	CourseId       string `json:"course_id"`
	VideoId        string `json:"video_id"`
	LocalOutputDir string `json:"local_output_dir"`
}
