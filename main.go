package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/fsouza/go-dockerclient"
)

var keepAtLeast int
var keepYoungerThanDays int
var dryRun bool
var logPath string
var logFile *os.File
var logger *log.Logger

type LocalImage struct {
	ID        string
	Name      string
	Tag       string
	CreatedAt time.Time
}

func (s LocalImage) String() string {
	f := bytes.NewBufferString("")
	f.WriteString(s.ID[0:12])
	f.WriteString(" ")
	f.WriteString(s.Name)
	f.WriteString(":")
	f.WriteString(s.Tag)
	f.WriteString(" ")
	f.WriteString(s.CreatedAt.Format(time.RFC822))
	return f.String()
}

// Image sorting by age
type ByAge []LocalImage

func (im ByAge) Len() int           { return len(im) }
func (im ByAge) Swap(i, j int)      { im[i], im[j] = im[j], im[i] }
func (im ByAge) Less(i, j int) bool { return im[i].CreatedAt.After(im[j].CreatedAt) }

func main() {
	flag.IntVar(&keepAtLeast, "k", 5, "Number of image-versions to keep")
	flag.IntVar(&keepYoungerThanDays, "a", 14, "Minimum age (in days) an image must be before deletion")
	flag.BoolVar(&dryRun, "n", false, "Dry-run, will only list images to delete, without performing actual deletion")
	flag.StringVar(&logPath, "l", "/var/log/docker-image-cleanup.log", "Path to log file")
	flag.Parse()
	endpoint := "unix:///var/run/docker.sock" // localhost for now
	client, _ := docker.NewClient(endpoint)

	// Setup logger
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Unable to open logfile: ", logPath, " Got err ", err)
		os.Exit(1)
	}
	logger = log.New(io.MultiWriter(logFile, os.Stdout), "docker-image-cleanup - ", log.Flags())

	// Pull image info
	LocalImages := make(map[string][]LocalImage)
	images, err := client.ListImages(false)
	if err != nil {
		logger.Fatal(err)
	}
	for _, img := range images {
		nameParts := strings.Split(img.RepoTags[0], ":")
		ii := LocalImage{ID: img.ID, Name: nameParts[0], Tag: nameParts[1], CreatedAt: time.Unix(img.Created, 0)}
		LocalImages[nameParts[0]] = append(LocalImages[nameParts[0]], ii)
	}

	// Find old/extraneous image versions and remove
	now := time.Now()
	for name, imArr := range LocalImages {
		if len(imArr) <= keepAtLeast {
			logger.Println("Ignoring images from", name, "which has less than the minimum", keepAtLeast, "image versions.")
			continue
		}
		sort.Sort(ByAge(imArr))
		saveImages := imArr[0:keepAtLeast]
		delImages := imArr[keepAtLeast:]
		for _, val := range saveImages {
			logger.Println("Keeping ", val)
		}
		for _, val := range delImages {
			if now.Add(time.Duration(-keepYoungerThanDays*24) * time.Hour).Before(val.CreatedAt) {
				logger.Println("Avoiding deletion of too recent image", val)
				continue
			}
			if dryRun == false {
				logger.Println("Deleting", val)
				err = client.RemoveImage(val.ID)
				if err != nil {
					logger.Println("Got err from RemoveImage", err)
				}
			} else {
				logger.Println("Would delete", val)
			}
		}
	}
}
