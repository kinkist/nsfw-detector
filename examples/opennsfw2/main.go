// OpenNSFW2 image/video classifier example.
//
// Usage:
//
//	go run . <opennsfw2.onnx> <image_or_video>
//	go run . <opennsfw2.onnx> <image_or_video> [ort_lib_path]
//
// Example:
//
//	go run . /models/opennsfw2.onnx photo.jpg
//	go run . /models/opennsfw2.onnx video.mp4
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/kinkist/nsfw-detector"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <opennsfw2.onnx> <image_or_video> [ort_lib_path]\n", os.Args[0])
		os.Exit(1)
	}

	modelPath := os.Args[1]
	filePath := os.Args[2]
	libPath := ""
	if len(os.Args) >= 4 {
		libPath = os.Args[3]
	}

	// Initialize OpenNSFW2 (tensor names are auto-detected from model)
	if err := nsfwdetector.Init(modelPath, libPath, "", ""); err != nil {
		log.Fatalf("Init failed: %v", err)
	}
	defer nsfwdetector.Close()

	// Run detection (auto-detects image vs video by extension)
	result, err := nsfwdetector.Detect(filePath)
	if err != nil {
		log.Fatalf("Detection failed: %v", err)
	}

	fmt.Printf("File : %s\n", filePath)
	fmt.Printf("SFW  : %.4f\n", result.SFW)
	fmt.Printf("NSFW : %.4f\n", result.NSFW)
	fmt.Println("---")
	if result.IsNSFW(0.5) {
		fmt.Println("Result: NSFW ⚠️")
	} else {
		fmt.Println("Result: SFW ✅")
	}
}
