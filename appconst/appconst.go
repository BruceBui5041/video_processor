package appconst

const (
	VideoMaxConcurrentResolutionParse = 3
	VideoMaxConcurrentHLSProcesses    = 1
	UnprecessedVideoDir               = "unprocessed_video"
)

const (
	TopicVideoProcessed   = "video_processed"
	TopicNewVideoUploaded = "new_video_uploaded"
)

const (
	MaxConcurrentS3Push  = 50
	AWSVideoS3BuckerName = "hls-video-segment"
	AWSRegion            = "ap-southeast-1"
)
