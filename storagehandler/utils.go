package storagehandler

import (
	"fmt"
)

type VideoInfo struct {
	Useremail  string
	CourseSlug string
	VideoSlug  string
	Filename   string
}

func GenerateSegmentS3Key(info VideoInfo) string {
	return fmt.Sprintf("course/%s/%s/%s/segments/%s",
		info.Useremail,
		info.CourseSlug,
		info.VideoSlug,
		info.Filename,
	)
}
