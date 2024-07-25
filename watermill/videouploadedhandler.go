package watermill

import (
	"fmt"

	"github.com/ThreeDotsLabs/watermill/message"
)

func HandleNewVideoUploadEvent(msg *message.Message) {
	s3Key := string(msg.Payload)
	fmt.Printf("New video uploaded: %s\n", s3Key)
	// Add your processing logic here
	// For example, you could start a video transcoding job
	msg.Ack()
}
