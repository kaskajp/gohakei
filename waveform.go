package main

import (
	"flag"
	"fmt"
	"image/color"
	"image/png"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/iFaceless/godub"
	"github.com/iFaceless/godub/converter"
	"github.com/xigh/go-waveform"
	"github.com/xigh/go-wavreader"
)

// Split an mp3 into multiple segments
// mp3File: The mp3 file to split
// segmentTime: The time in seconds to split the mp3 (default is 60 but can be overriden)
// Returns object with the segments and count
func splitMp3(mp3File string, segmentTime int, uuid string) error {
	if segmentTime == 0 {
		segmentTime = 60
	}

	cmd := exec.Command("ffmpeg", "-i", mp3File, "-f", "segment", "-segment_time", fmt.Sprintf("%d", segmentTime), "-c", "copy", "output/"+uuid+"/audio/%03d.mp3")
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func getFilesInDirectory(directory string) ([]string, error) {
	files, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	var fileNames []string
	for _, file := range files {
		fileNames = append(fileNames, file.Name())
	}

	return fileNames, nil
}

func convertAllMp3ToWaveformImages(uuid string) error {
	files, err := getFilesInDirectory("output/" + uuid + "/audio")
	if err != nil {
		return err
	}

	for _, file := range files {
		if strings.Contains(file, ".mp3") {
			fileName := strings.Replace(file, ".mp3", ".png", -1)

			cmd := exec.Command("ffprobe", "-i", "output/"+uuid+"/audio/"+file, "-show_entries", "format=duration", "-v", "quiet", "-of", "csv=p=0")
			d, err := cmd.Output()
			if err != nil {
				return err
			}

			d = d[:len(d)-1]
			dStr := string(d)
			dFloat, err := strconv.ParseFloat(dStr, 64)
			if err != nil {
				return err
			}

			duration := int(dFloat)

			width := duration * 100

			cmd = exec.Command("audiowaveform", "-i", "output/"+uuid+"/audio/"+file, "-o", "output/"+uuid+"/images/"+fileName, "--no-axis-labels", "--pixels-per-second", "100", "--width", fmt.Sprint(width), "--background-color", "00000000", "--waveform-color", "FFFFFFFF", "--height", "80")
			err = cmd.Run()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func convertMp3ToWav(mp3File string) string {
	filePath := mp3File
	toFilePath := strings.Replace(filePath, ".mp3", ".wav", -1)
	w, _ := os.Create(toFilePath)

	err := converter.NewConverter(w).
		WithBitRate(64000).
		WithDstFormat("wav").
		Convert(filePath)
	if err != nil {
		log.Fatal(err)
	}

	return toFilePath
}

func wavToWaveform(wavFile string) error {
	fmt.Println("wavToWaveform", wavFile)

	segment, _ := godub.NewLoader().Load(wavFile)
	fmt.Println(segment)

	r, err := os.Open(wavFile)
	if err != nil {
		return err
	}
	defer r.Close()

	w0, err := wavreader.New(r)
	if err != nil {
       	return err
	}

    margin := flag.Int("margin", 0, "margin")

    duration := w0.Duration().Seconds()
    fmt.Println("Duration:", duration)

    width := int(duration * 10)
    height := 50

	img := waveform.MinMax(w0, &waveform.Options{
		Width:   width,
		Height:  height,
		Zoom:    1,
		Half:    false,
		MarginL: *margin,
		MarginR: *margin,
		MarginT: *margin,
		MarginB: *margin,
		Front: &color.NRGBA{
			R: 255,
			G: 128,
			B: 255,
			A: 150,
		},
		Back: &color.NRGBA{
			A: 0,
		},
	})

	// Generate a uuid for the file name
	uuid := uuid.NewString()

	w, err := os.Create("output/"+uuid+".png")
	if err != nil {
		return err
	}
	defer w.Close()

	err = png.Encode(w, img)
	if err != nil {
		return err
    }
    return nil
}

func main() {
	uuid := uuid.NewString()
	// Create the output directory
	os.MkdirAll("output/"+uuid, os.ModePerm)
	os.MkdirAll("output/"+uuid+"/audio", os.ModePerm)
	os.MkdirAll("output/"+uuid+"/images", os.ModePerm)

	err := splitMp3("audio/15min.mp3", 60, uuid)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	err = convertAllMp3ToWaveformImages(uuid)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	/*wav := convertMp3ToWav("audio/15min.mp3")
	err := wavToWaveform(wav) // Capture the error
	if err != nil {
		fmt.Println("Error:", err)
		return
	}*/

	// err := wavToWaveform("audio/Half_Day.wav") // Capture the error
	// if err != nil {
	//	fmt.Println("Error:", err)
	//	return
	// }
}
