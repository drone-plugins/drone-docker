package docker

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/drone/drone-go/drone"

	"github.com/inhies/go-bytesize"
)

// writeCard maintains backward compatibility by using TempTag
func (p Plugin) writeCard() error {
	return p.writeCardForImage(p.Build.TempTag)
}

// writeCardForImage generates card for any image reference
func (p Plugin) writeCardForImage(imageRef string) error {
	cmd := exec.Command(dockerExe, "inspect", imageRef)
	data, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	out := Card{}
	if err := json.Unmarshal(data, &out); err != nil {
		return err
	}

	inspect := out[0]
	inspect.SizeString = fmt.Sprint(bytesize.New(float64(inspect.Size)))
	inspect.VirtualSizeString = fmt.Sprint(bytesize.New(float64(inspect.VirtualSize)))
	inspect.Time = fmt.Sprint(inspect.Metadata.LastTagTime.Format(time.RFC3339))
	// change slice of tags to slice of TagStruct
	var sliceTagStruct []TagStruct
	for _, tag := range inspect.RepoTags {
		sliceTagStruct = append(sliceTagStruct, TagStruct{Tag: tag})
	}
	if len(sliceTagStruct) > 1 {
		inspect.ParsedRepoTags = sliceTagStruct[1:] // remove the first tag which is always "hash:latest"
	} else {
		inspect.ParsedRepoTags = sliceTagStruct
	}
	// create the url from repo and registry
	inspect.URL = mapRegistryToURL(p.Daemon.Registry, p.Build.Repo)
	cardData, _ := json.Marshal(inspect)

	card := drone.CardInput{
		Schema: "https://drone-plugins.github.io/drone-docker/card.json",
		Data:   cardData,
	}

	writeCard(p.CardPath, &card)
	return nil
}

func writeCard(path string, card interface{}) {
	data, _ := json.Marshal(card)
	switch {
	case path == "/dev/stdout":
		writeCardTo(os.Stdout, data)
	case path == "/dev/stderr":
		writeCardTo(os.Stderr, data)
	case path != "":
		ioutil.WriteFile(path, data, 0644)
	}
}

func writeCardTo(out io.Writer, data []byte) {
	encoded := base64.StdEncoding.EncodeToString(data)
	io.WriteString(out, "\u001B]1338;")
	io.WriteString(out, encoded)
	io.WriteString(out, "\u001B]0m")
	io.WriteString(out, "\n")
}

func mapRegistryToURL(registry, repo string) (url string) {
	url = "https://"
	var domain string
	if strings.Contains(registry, "amazonaws.com") {
		domain = "gallery.ecr.aws/"
	} else if strings.Contains(registry, "gcr.io") {
		domain = "console.cloud.google.com/gcr/images"
	} else {
		// default to docker hub
		domain = "hub.docker.com/r/"
	}
	url = path.Join(url, domain, repo)
	return url
}
