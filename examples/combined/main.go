// Combined NSFW detector example — runs both OpenNSFW2 and NudeNet v2 simultaneously.
//
// Usage:
//
//	go run . <opennsfw2.onnx> <nudenet.onnx> <image_or_video>
//	go run . <opennsfw2.onnx> <nudenet.onnx> <image_or_video> [ort_lib_path]
//
// Example:
//
//	go run . /models/opennsfw2.onnx /models/detector_v2_default_checkpoint.onnx photo.jpg
//	go run . /models/opennsfw2.onnx /models/detector_v2_default_checkpoint.onnx clip.mp4
//
// Notes:
//   - On macOS only one model can use CoreML at a time; the second model falls back to CPU automatically.
//   - On CUDA-capable Linux/Windows both models run on GPU simultaneously.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/kinkist/nsfw-detector"
	"github.com/kinkist/nsfw-detector/logger"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr,
			"Usage: %s <opennsfw2.onnx> <nudenet.onnx> <image_or_video> [ort_lib_path]\n",
			os.Args[0])
		os.Exit(1)
	}

	openModel := os.Args[1]
	nudeModel := os.Args[2]
	filePath := os.Args[3]
	libPath := ""
	if len(os.Args) >= 5 {
		libPath = os.Args[4]
	}

	// Enable debug logging (optional — comment out to silence)
	logger.Enabled = false

	// Initialize OpenNSFW2
	if err := nsfwdetector.Init(openModel, libPath, "", ""); err != nil {
		log.Fatalf("Init (OpenNSFW2) failed: %v", err)
	}
	defer nsfwdetector.Close()

	// Initialize NudeNet v2 (safe to call after Init — ONNX environment reuse)
	if err := nsfwdetector.InitNudeNet(nudeModel, libPath); err != nil {
		log.Fatalf("InitNudeNet failed: %v", err)
	}
	defer nsfwdetector.CloseNudeNet()

	// Run both detectors in a single call
	result, err := nsfwdetector.Detect(filePath)
	if err != nil {
		log.Fatalf("Detection failed: %v", err)
	}

	fmt.Printf("File : %s\n", filePath)
	fmt.Println("---")

	// OpenNSFW2 results
	fmt.Printf("[OpenNSFW2]\n")
	fmt.Printf("  SFW  : %.4f\n", result.SFW)
	fmt.Printf("  NSFW : %.4f\n", result.NSFW)

	// NudeNet v2 results
	fmt.Printf("[NudeNet v2]\n")
	if len(result.Detections) == 0 {
		fmt.Println("  No detections above threshold.")
	} else {
		for _, d := range result.Detections {
			fmt.Printf("  %-35s  %.4f\n", d.Class, d.Score)
		}
	}

	fmt.Println("---")
	if result.IsNSFW(0.5) {
		fmt.Println("Final result: NSFW ⚠️")
	} else {
		fmt.Println("Final result: SFW ✅")
	}
}
