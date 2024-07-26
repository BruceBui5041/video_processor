package grpcserver

import (
	"context"
	"video_processor/messagemodel"
	pb "video_processor/proto/video_service/video_service" // import the generated protobuf package
	"video_processor/watermill"

	"google.golang.org/grpc/codes"
)

type VideoServiceServer struct {
	pb.UnimplementedVideoServiceServer
}

func (s *VideoServiceServer) ProcessNewVideoRequest(ctx context.Context, req *pb.VideoInfo) (*pb.ProcessNewVideoResponse, error) {

	videoInfo := messagemodel.VideoInfo{
		VideoID:     req.VideoId,
		Title:       req.Title,
		Description: req.Description,
		UploadedBy:  req.UploadedBy,
		S3Key:       req.S3Key,
		Timestamp:   req.Timestamp,
	}

	go watermill.PublishVideoUploadedEvent(&videoInfo)
	return &pb.ProcessNewVideoResponse{Status: codes.OK.String()}, nil
}
