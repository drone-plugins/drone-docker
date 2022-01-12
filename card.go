package docker

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

func (p Plugin) writeCard() error {
	cmd := exec.Command("docker", "inspect", p.Build.Name)
	data, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	out := map[string]interface{}{} // replace with docker inspect struct
	if err := json.Unmarshal(data, &out); err != nil {
		return err
	}

	card := map[string]interface{}{} // replace with card struct, populate with docker inspect output

	writeCard( /*p.CardPath*/ "", &card)
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
