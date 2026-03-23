// NudeNet v2 object detector example.
//
// Detects 18 anatomical classes and prints each detection with its confidence score.
//
// Usage:
//
//	go run . <nudenet.onnx> <image_or_video>
//	go run . <nudenet.onnx> <image_or_video> [ort_lib_path]
//
// Example:
//
//	go run . /models/detector_v2_default_checkpoint.onnx photo.jpg
//	go run . /models/detector_v2_default_checkpoint.onnx clip.mp4
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/kinkist/nsfw-detector"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <nudenet.onnx> <image_or_video> [ort_lib_path]\n", os.Args[0])
		os.Exit(1)
	}

	modelPath := os.Args[1]
	filePath := os.Args[2]
	libPath := ""
	if len(os.Args) >= 4 {
		libPath = os.Args[3]
	}

	// Initialize NudeNet v2 detector
	if err := nsfwdetector.InitNudeNet(modelPath, libPath); err != nil {
		log.Fatalf("InitNudeNet failed: %v", err)
	}
	defer nsfwdetector.CloseNudeNet()

	// Run detection (auto-detects image vs video by extension)
	result, err := nsfwdetector.Detect(filePath)
	if err != nil {
		log.Fatalf("Detection failed: %v", err)
	}

	fmt.Printf("File: %s\n", filePath)
	fmt.Println("---")

	if len(result.Detections) == 0 {
		fmt.Println("No detections above threshold.")
		return
	}

	fmt.Printf("Detections (%d):\n", len(result.Detections))
	for _, d := range result.Detections {
		bar := ""
		n := int(d.Score * 20)
		for i := 0; i < n; i++ {
			bar += "█"
		}
		fmt.Printf("  %-35s  %.4f  %s\n", d.Class, d.Score, bar)
	}

	if result.IsNSFW(0.5) {
		fmt.Println("\nResult: NSFW ⚠️")
	} else {
		fmt.Println("\nResult: SFW ✅")
	}
}
