# Landmark Recordings

A demo for building recordings from landmark information provided by [mediapipe](https://mediapipe.dev/).

## Requirements

* FFMPEG
* python 3.9 64-bit (mediapipe requires 64bit)
* Golang 1.17 (I'm sure most versions would work, this is just whats on my machine.)

## Processessing Recordings

### Deconstruct the Video

```bash
ffmpeg -i in.mp4 frames/frame_%04d.png -hide_banner
```

### Run the Landmark Identification Software

```bash
python pose/pose.py
```

### Rebuilding the Video

```bash
ffmpeg -y -r 29 -i frames_out/frame_%04d.png -c:v libx264 -vf fps=29 -pix_fmt yuv420p out.mp4
```

### Buiding a Recolude Recording

```bash
go run pose.go
```