package docker

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"github.com/drone/drone-go/drone"
	"github.com/inhies/go-bytesize"
)

func (p Plugin) writeCard() error {
	cmd := exec.Command("docker", "inspect", p.Build.Name)
	data, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	out := Inspect{}
	if err := json.Unmarshal(data, &out); err != nil {
		return err
	}

	inspect := out[0]
	inspect.SizeString = fmt.Sprint(bytesize.New(float64(inspect.Size)))
	inspect.VirtualSizeString = fmt.Sprint(bytesize.New(float64(inspect.VirtualSize)))
	inspect.Time = fmt.Sprint(inspect.Metadata.LastTagTime.Format(time.RFC3339))
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
