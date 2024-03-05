package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

/*
Split an mp3 into multiple segments.

Parameters:
	mp3File: The path to the mp3 file to split
	segmentTime: Segment length in seconds (default is 60)
	uuid: The uuid of the request

Returns:
	error: An error if the operation fails
*/
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

/*
Get all files in a directory.

Parameters:
	directory: The directory to get the files from

Returns:
	[]string: A list of file names
	error: An error if the operation fails
*/
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

/*
Convert all mp3 files in a directory to waveform images.

Parameters:
	uuid: The uuid of the request

Returns:
	error: An error if the operation fails
*/
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

			cmd = exec.Command("audiowaveform", "-i", "output/"+uuid+"/audio/"+file, "-o", "output/"+uuid+"/images/"+fileName, "--no-axis-labels", "--pixels-per-second", "100", "--width", fmt.Sprint(width), "--background-color", "00000000", "--waveform-color", "999FAC", "--height", "80")
			err = cmd.Run()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func main() {
	uuid := uuid.NewString()

	// Create the output directory
	os.MkdirAll("output/"+uuid, os.ModePerm)
	os.MkdirAll("output/"+uuid+"/audio", os.ModePerm)
	os.MkdirAll("output/"+uuid+"/images", os.ModePerm)

	err := splitMp3("audio/long2.mp3", 60, uuid)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	err = convertAllMp3ToWaveformImages(uuid)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
}
