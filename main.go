package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/fsouza/go-dockerclient"
)

var keepAtLeast int
var keepYoungerThanDays int

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
	flag.Parse()
	endpoint := "unix:///var/run/docker.sock" // localhost for now
	client, _ := docker.NewClient(endpoint)

	// Pull image info
	LocalImages := make(map[string][]LocalImage)
	images, err := client.ListImages(false)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
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
			fmt.Println("Ignoring images frpm", name, "which has less than the minimum", keepAtLeast, "image versions.")
			continue
		}
		sort.Sort(ByAge(imArr))
		saveImages := imArr[0:keepAtLeast]
		delImages := imArr[keepAtLeast:]
		for _, val := range saveImages {
			fmt.Println("Keeping ", val)
		}
		for _, val := range delImages {
			if now.Add(time.Duration(-keepYoungerThanDays*24) * time.Hour).Before(val.CreatedAt) {
				fmt.Println("Avoiding deletion of too recent image", val)
				continue
			}
			fmt.Println("Deleting", val)
			err = client.RemoveImage(val.ID)
			if err != nil {
				fmt.Println("Got err from RemoveImage", err)
			}
		}
	}

}
