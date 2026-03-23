# nsfw-detector

ONNX-based NSFW content detection library for Go.
Two independent detectors can run simultaneously.

| Detector | Model file | Description |
|---|---|---|
| **OpenNSFW2** | `nsfw_model.onnx` | Binary SFW / NSFW image classifier |
| **NudeNet v2** | `detector_v2_default_checkpoint.onnx` | Object-level detector for 18 anatomical classes |

---

## Installation

```bash
go get github.com/kinkist/nsfw-detector
```

### Requirements

| Item | Details |
|---|---|
| Go | 1.21 or later |
| ONNX Runtime | [onnxruntime_go v1.27.0](https://github.com/yalue/onnxruntime_go) — shared library required |
| ffmpeg / ffprobe | Required for video processing (`brew install ffmpeg` / `apt install ffmpeg`) |

The ONNX Runtime shared library (`libonnxruntime.so` / `.dylib`) must be downloaded separately.
→ [ONNX Runtime GitHub Releases](https://github.com/microsoft/onnxruntime/releases)

---

## Quick Start

```go
import "github.com/kinkist/nsfw-detector"
```

### OpenNSFW2 only (binary classification)

```go
// Initialize (tensor names are auto-detected from the model)
if err := nsfwdetector.Init("nsfw_model.onnx", "", "", ""); err != nil {
    log.Fatal(err)
}
defer nsfwdetector.Close()

// Detect on an image
result, err := nsfwdetector.DetectImage("photo.jpg")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("SFW=%.4f  NSFW=%.4f\n", result.SFW, result.NSFW)
```

### NudeNet v2 only (object detection)

```go
if err := nsfwdetector.InitNudeNet("detector_v2_default_checkpoint.onnx", ""); err != nil {
    log.Fatal(err)
}
defer nsfwdetector.CloseNudeNet()

result, err := nsfwdetector.DetectImage("photo.jpg")
if err != nil {
    log.Fatal(err)
}
for _, d := range result.Detections {
    fmt.Printf("%-35s %.4f\n", d.Class, d.Score)
}
```

### Both models simultaneously

```go
nsfwdetector.Init("nsfw_model.onnx", "", "", "")
defer nsfwdetector.Close()

nsfwdetector.InitNudeNet("detector_v2_default_checkpoint.onnx", "")
defer nsfwdetector.CloseNudeNet()

// Auto-detects image vs video by file extension
result, err := nsfwdetector.Detect("photo.jpg")

fmt.Printf("NSFW=%.4f\n", result.NSFW)
for _, d := range result.Detections {
    fmt.Printf("  %s: %.4f\n", d.Class, d.Score)
}

if result.IsNSFW(0.5) {
    fmt.Println("NSFW content detected")
}
```

### Video processing

```go
result, err := nsfwdetector.DetectVideo("video.mp4")
// Extracts up to 5 frames via ffmpeg and runs detection on each.
// OpenNSFW2 : returns the highest NSFW score across all frames
// NudeNet v2: returns the highest score per class across all frames
```

---

## API Reference

### Initialization / Cleanup

| Function | Description |
|---|---|
| `Init(modelPath, libPath, inName, outName string) error` | Initialize OpenNSFW2. Empty strings for tensor names → auto-detect from model |
| `InitNudeNet(modelPath, libPath string) error` | Initialize NudeNet v2 |
| `Close()` | Release OpenNSFW2 resources |
| `CloseNudeNet()` | Release NudeNet v2 resources |
| `IsEnabled() bool` | Reports whether OpenNSFW2 is initialized |
| `IsNudeNetEnabled() bool` | Reports whether NudeNet v2 is initialized |

### Detection

| Function | Description |
|---|---|
| `Detect(filePath string) (*Result, error)` | Auto-detects image vs video by extension |
| `DetectImage(imagePath string) (*Result, error)` | Run detection on an image file (JPEG, PNG, GIF) |
| `DetectVideo(videoPath string) (*Result, error)` | Run detection on a video file (requires ffmpeg) |

### Result type

```go
type Result struct {
    SFW        float32          // OpenNSFW2 SFW probability  (0–1)
    NSFW       float32          // OpenNSFW2 NSFW probability (0–1)
    Detections []NudeDetection  // NudeNet v2 detections, sorted by score descending
}

func (r *Result) IsNSFW(threshold float32) bool
func (r *Result) String() string
```

### NudeDetection type

```go
type NudeDetection struct {
    Class string   // Detected class name
    Score float32  // Confidence score (0–1)
}
```

### NudeNet v2 detection classes (18)

```
FEMALE_GENITALIA_COVERED    FEMALE_GENITALIA_EXPOSED
FEMALE_BREAST_COVERED       FEMALE_BREAST_EXPOSED
MALE_BREAST_EXPOSED         MALE_GENITALIA_EXPOSED
BUTTOCKS_COVERED            BUTTOCKS_EXPOSED
ANUS_COVERED                ANUS_EXPOSED
ARMPITS_COVERED             ARMPITS_EXPOSED
FEET_COVERED                FEET_EXPOSED
BELLY_COVERED               BELLY_EXPOSED
FACE_COVERED                FACE_FEMALE
```

---

## Debug Logging

```go
import "github.com/kinkist/nsfw-detector/logger"

logger.Enabled = true  // Print detailed inference logs to stdout
```

---

## GPU Support

| Platform | Accelerator | Notes |
|---|---|---|
| macOS | CoreML (Apple Neural Engine) | Only one model can use CoreML at a time; the second falls back to CPU automatically |
| Linux / Windows | CUDA (NVIDIA GPU) | Both models can use CUDA simultaneously |
| Fallback | CPU | Automatic fallback when CUDA / CoreML is unavailable |

---

## Examples

```
examples/
├── opennsfw2/   # OpenNSFW2 standalone
├── nudenet/     # NudeNet v2 standalone
└── combined/    # Both models together
```

```bash
# Run OpenNSFW2 example
go run ./examples/opennsfw2 /path/to/nsfw_model.onnx photo.jpg

# Run NudeNet example
go run ./examples/nudenet /path/to/detector_v2_default_checkpoint.onnx photo.jpg

# Run combined example
go run ./examples/combined /path/to/nsfw_model.onnx /path/to/detector_v2_default_checkpoint.onnx photo.jpg
```

---

## Project Structure

```
nsfw-detector/
├── nsfw-detector.go  # Public API (Result, Detect, DetectImage, DetectVideo)
├── opennsfw2.go      # OpenNSFW2 binary classifier
├── nudenetv2.go      # NudeNet v2 object detector
├── video.go          # ffmpeg-based video frame extraction
├── logger/
│   └── logger.go     # Debug logger subpackage
├── examples/
│   ├── opennsfw2/    # OpenNSFW2 standalone example
│   ├── nudenet/      # NudeNet v2 standalone example
│   └── combined/     # Combined example
├── go.mod
└── go.sum
```

---

## Model Download & Preparation

### 1. `detector_v2_default_checkpoint.onnx` (NudeNet v2)

#### Option A — Direct download (simplest)

```bash
wget https://github.com/notAI-tech/NudeNet/releases/download/v3.4/detector_v2_default_checkpoint.onnx
# or
curl -L -o detector_v2_default_checkpoint.onnx \
  https://github.com/notAI-tech/NudeNet/releases/download/v3.4/detector_v2_default_checkpoint.onnx
```

#### Option B — Via Python nudenet package

Installing the `nudenet` package automatically downloads the model on first run.

```bash
pip install nudenet
```

```python
import nudenet, os, shutil

detector = nudenet.NudeDetector()   # triggers automatic model download

# Find and copy the .onnx file from the package directory
pkg_dir = os.path.dirname(nudenet.__file__)
for root, _, files in os.walk(pkg_dir):
    for f in files:
        if f.endswith(".onnx"):
            src = os.path.join(root, f)
            shutil.copy(src, "detector_v2_default_checkpoint.onnx")
            print("Copied:", src)
```

---

### 2. `nsfw_model.onnx` (OpenNSFW2)

Yahoo OpenNSFW architecture exported to ONNX format.

#### Option A — Download from Hugging Face

```bash
pip install huggingface_hub

python -c "
from huggingface_hub import hf_hub_download
import shutil
path = hf_hub_download(
    repo_id='bumble-tech/open-nsfw',
    filename='model/saved_model.onnx',
    local_dir='.'
)
shutil.copy(path, 'nsfw_model.onnx')
print('Saved: nsfw_model.onnx')
"
```

#### Option B — Convert via Python opennsfw2 package (Keras → ONNX)

```bash
pip install opennsfw2 tf2onnx tensorflow
```

```python
import opennsfw2 as n2
import os

# 1. Load the Keras model
model = n2.make_open_nsfw_model()

# 2. Save as SavedModel
model.save("opennsfw2_saved")

# 3. Convert to ONNX
os.system(
    "python -m tf2onnx.convert "
    "--saved-model opennsfw2_saved "
    "--output nsfw_model.onnx "
    "--opset 13"
)
print("Done: nsfw_model.onnx")
```

#### Option C — One-liner

```bash
pip install opennsfw2 tf2onnx tensorflow && \
python -c "
import opennsfw2 as n2, os, shutil
n2.make_open_nsfw_model().save('_tmp_nsfw')
os.system('python -m tf2onnx.convert --saved-model _tmp_nsfw --output nsfw_model.onnx --opset 13')
shutil.rmtree('_tmp_nsfw')
print('Done: nsfw_model.onnx')
"
```

---

### Inspect model tensor names

Use this script to verify the input/output tensor names after conversion.
Not required in most cases — `Init()` auto-detects tensor names when `inName`/`outName` are empty strings.

```python
import onnxruntime as ort

def show_model_info(path):
    sess = ort.InferenceSession(path)
    print(f"=== {path} ===")
    for i in sess.get_inputs():
        print(f"  INPUT  {i.name!r:40s} {i.shape}")
    for o in sess.get_outputs():
        print(f"  OUTPUT {o.name!r:40s} {o.shape}")

show_model_info("nsfw_model.onnx")
show_model_info("detector_v2_default_checkpoint.onnx")
```

```bash
pip install onnxruntime
python check_model.py
```

---

## License

MIT
