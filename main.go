package main

import (
	"log"
	"net"
	"video_processor/grpcserver"
	pb "video_processor/proto/video_service/video_service"
	redishander "video_processor/redishandler"
	"video_processor/watermill"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	// fileName := "test.mp4"
	// outputDir := "segments"

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Start the subscriber in a goroutine
	go watermill.SubscribeToTopics()

	go redishander.StartRedisSubscribers(redishander.RedisClient)

	// go hlssegmenter.StartSegmentProcess(fileName, outputDir)

	// clean up
	// err = utils.DeleteLocalFile(unprecessedVideoPath)
	// if err != nil {
	// 	fmt.Print(err.Error())
	// }

	// err = utils.DeleteDirContents(desireOutputPath)
	// if err != nil {
	// 	fmt.Print(err.Error())
	// }

	// Start gRPC server
	startGRPCServer()
}

func startGRPCServer() {
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()

	// Register your gRPC services here
	// For example:
	pb.RegisterVideoServiceServer(s, &grpcserver.VideoServiceServer{})

	log.Println("Starting gRPC server on :50052")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
