// Package nsfwdetector provides ONNX-based NSFW detection for images and videos.
//
// Two independent detectors are supported and can run simultaneously:
//   - OpenNSFW2  (opennsfw2.go): binary SFW/NSFW classifier
//   - NudeNet v2 (nudenetv2.go): object-level nude content detector
//
// # Quick start
//
//	// Initialize one or both detectors
//	if err := nsfwdetector.Init("opennsfw2.onnx", "", "", ""); err != nil {
//	    log.Fatal(err)
//	}
//	defer nsfwdetector.Close()
//
//	if err := nsfwdetector.InitNudeNet("nudenet.onnx", ""); err != nil {
//	    log.Fatal(err)
//	}
//	defer nsfwdetector.CloseNudeNet()
//
//	// Detect on an image or video (auto-detected by extension)
//	result, err := nsfwdetector.Detect("photo.jpg")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("NSFW=%.4f  SFW=%.4f\n", result.NSFW, result.SFW)
//	if result.IsNSFW(0.5) {
//	    fmt.Println("Content flagged as NSFW")
//	}
package nsfwdetector

import (
	"fmt"
	"strings"
)

// Result holds combined detection results from all active models.
type Result struct {
	// SFW/NSFW probabilities from OpenNSFW2 (each 0–1, softmax output so SFW+NSFW ≈ 1).
	// Both are zero when OpenNSFW2 is not initialized.
	SFW  float32
	NSFW float32

	// Detections from NudeNet v2, sorted by confidence descending.
	// Nil when NudeNet is not initialized or nothing was detected above threshold.
	Detections []NudeDetection
}

// IsNSFW returns true when the NSFW probability or any detection score meets or exceeds threshold.
func (r *Result) IsNSFW(threshold float32) bool {
	if r.NSFW >= threshold {
		return true
	}
	for _, d := range r.Detections {
		if d.Score >= threshold {
			return true
		}
	}
	return false
}

// String returns a human-readable summary of the result.
//
// Detections == nil means NudeNet was not run.
// Detections != nil (including empty slice) means NudeNet ran.
func (r *Result) String() string {
	var sb strings.Builder
	if r.SFW != 0 || r.NSFW != 0 {
		fmt.Fprintf(&sb, "SFW:  %.4f\nNSFW: %.4f\n", r.SFW, r.NSFW)
	}
	if r.Detections != nil {
		if len(r.Detections) == 0 {
			sb.WriteString("NUDENET: (none above threshold)\n")
		} else {
			sb.WriteString("NUDENET:")
			for _, d := range r.Detections {
				fmt.Fprintf(&sb, " %s:%.4f", d.Class, d.Score)
			}
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}

// IsEnabled reports whether OpenNSFW2 is initialized.
func IsEnabled() bool { return enabled }

// Detect auto-detects whether filePath is a video or image and runs all active detectors.
// Requires ffmpeg in PATH for video files. At least one detector must be initialized.
func Detect(filePath string) (*Result, error) {
	if isVideoFile(filePath) {
		return DetectVideo(filePath)
	}
	return DetectImage(filePath)
}

// DetectImage runs all active detectors on a single image file (JPEG, PNG, GIF).
// Returns an error only when detection itself fails; empty results are not errors.
func DetectImage(imagePath string) (*Result, error) {
	if !enabled && !nudeEnabled {
		return nil, fmt.Errorf("nsfw: no detector initialized; call Init() and/or InitNudeNet() first")
	}

	result := &Result{}

	if enabled {
		sfw, nsfw, err := detectOpenNSFW2(imagePath)
		if err != nil {
			return nil, fmt.Errorf("OpenNSFW2: %w", err)
		}
		result.SFW = sfw
		result.NSFW = nsfw
	}

	if nudeEnabled {
		dets, err := detectNudeNet(imagePath)
		if err != nil {
			return nil, fmt.Errorf("NudeNet v2: %w", err)
		}
		// Use empty non-nil slice to distinguish "ran but found nothing" from "not run".
		if dets == nil {
			dets = []NudeDetection{}
		}
		result.Detections = dets
	}

	return result, nil
}

// DetectVideo runs all active detectors on a video file by extracting sample frames.
// Requires ffmpeg in PATH. At least one detector must be initialized.
func DetectVideo(videoPath string) (*Result, error) {
	if !enabled && !nudeEnabled {
		return nil, fmt.Errorf("nsfw: no detector initialized; call Init() and/or InitNudeNet() first")
	}

	result := &Result{}

	if enabled {
		sfw, nsfw, err := detectVideoOpenNSFW2(videoPath)
		if err != nil {
			return nil, fmt.Errorf("OpenNSFW2 video: %w", err)
		}
		result.SFW = sfw
		result.NSFW = nsfw
	}

	if nudeEnabled {
		dets, err := detectVideoNudeNet(videoPath)
		if err != nil {
			return nil, fmt.Errorf("NudeNet v2 video: %w", err)
		}
		if dets == nil {
			dets = []NudeDetection{}
		}
		result.Detections = dets
	}

	return result, nil
}
