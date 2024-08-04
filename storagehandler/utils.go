package storagehandler

import (
	"fmt"
)

type VideoInfo struct {
	UploadedBy string `json:"uploaded_by"`
	CourseId   string `json:"course_id"`
	VideoId    string `json:"video_id"`
	Filename   string
}

func GenerateSegmentS3Key(info VideoInfo) string {
	return fmt.Sprintf("course/%s/%s/%s/video_segment/%s",
		info.UploadedBy,
		info.CourseId,
		info.VideoId,
		info.VideoId,
	)
}
