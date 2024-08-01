package grpcserver

import (
	"context"
	"video_processor/logger"
	"video_processor/messagemodel"
	pb "video_processor/proto/video_service/video_service" // import the generated protobuf package
	"video_processor/watermill"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
)

type VideoServiceServer struct {
	pb.UnimplementedVideoProcessingServiceServer
}

func (s *VideoServiceServer) ProcessNewVideoRequest(ctx context.Context, req *pb.VideoInfo) (*pb.ProcessNewVideoResponse, error) {

	videoInfo := messagemodel.VideoInfo{
		RawVidS3Key: req.S3Key,
		Timestamp:   req.Timestamp,
		CourseSlug:  req.CourseSlug,
		VideoSlug:   req.VideoSlug,
		UserEmail:   req.UserEmail,
	}

	logger.AppLogger.Info("videoInfo", zap.Any("videoInfo", videoInfo))

	go watermill.PublishVideoUploadedEvent(&videoInfo)
	return &pb.ProcessNewVideoResponse{Status: codes.OK.String()}, nil
}
