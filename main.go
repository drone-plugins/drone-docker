package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/drone/drone-plugin-go/plugin"
)

type Save struct {
	// Absolute or relative path
	File string `json:"destination"`
	// Only save specified tags (optional)
	Tags StrSlice `json:"tag"`
}

type Docker struct {
	Storage   string   `json:"storage_driver"`
	Registry  string   `json:"registry"`
	Mirror    string   `json:"mirror"`
	Insecure  bool     `json:"insecure"`
	Username  string   `json:"username"`
	Password  string   `json:"password"`
	Email     string   `json:"email"`
	Auth      string   `json:"auth"`
	Repo      string   `json:"repo"`
	ForceTag  bool     `json:"force_tag"`
	Tag       StrSlice `json:"tag"`
	File      string   `json:"file"`
	Context   string   `json:"context"`
	Bip       string   `json:"bip"`
	Dns       []string `json:"dns"`
	Load      string   `json:"load"`
	Save      Save     `json:"save"`
	BuildArgs []string `json:"build_args"`
}

func main() {
	workspace := plugin.Workspace{}
	build := plugin.Build{}
	vargs := Docker{}

	plugin.Param("workspace", &workspace)
	plugin.Param("build", &build)
	plugin.Param("vargs", &vargs)
	plugin.MustParse()

	// in case someone uses the shorthand repository name
	// with a custom registry, we should concatinate so that
	// we have the fully qualified image name.
	if strings.Count(vargs.Repo, "/") <= 1 && len(vargs.Registry) != 0 && !strings.HasPrefix(vargs.Repo, vargs.Registry) {
		vargs.Repo = fmt.Sprintf("%s/%s", vargs.Registry, vargs.Repo)
	}

	// Set the Registry value
	if len(vargs.Registry) == 0 {
		vargs.Registry = "https://index.docker.io/v1/"
	}
	// Set the Dockerfile name
	if len(vargs.File) == 0 {
		vargs.File = "Dockerfile"
	}
	// Set the Context value
	if len(vargs.Context) == 0 {
		vargs.Context = "."
	}
	// Set the Tag value
	if vargs.Tag.Len() == 0 {
		vargs.Tag = StrSlice{[]string{"latest"}}
	}
	// Get absolute path for 'save' file
	if len(vargs.Save.File) != 0 {
		if !filepath.IsAbs(vargs.Save.File) {
			vargs.Save.File = filepath.Join(workspace.Path, vargs.Save.File)
		}
	}
	// Get absolute path for 'load' file
	if len(vargs.Load) != 0 {
		if !filepath.IsAbs(vargs.Load) {
			vargs.Load = filepath.Join(workspace.Path, vargs.Load)
		}
	}

	go func() {
		args := []string{"-d", "-g", "/drone/docker"}

		if len(vargs.Storage) != 0 {
			args = append(args, "-s", vargs.Storage)
		}
		if vargs.Insecure && len(vargs.Registry) != 0 {
			args = append(args, "--insecure-registry", vargs.Registry)
		}
		if len(vargs.Mirror) != 0 {
			args = append(args, "--registry-mirror", vargs.Mirror)
		}
		if len(vargs.Bip) != 0 {
			args = append(args, "--bip", vargs.Bip)
		}

		for _, value := range vargs.Dns {
			args = append(args, "--dns", value)
		}

		cmd := exec.Command("/usr/bin/docker", args...)
		if os.Getenv("DOCKER_LAUNCH_DEBUG") == "true" {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		} else {
			cmd.Stdout = ioutil.Discard
			cmd.Stderr = ioutil.Discard
		}
		trace(cmd)
		cmd.Run()
	}()

	// ping Docker until available
	for i := 0; i < 3; i++ {
		cmd := exec.Command("/usr/bin/docker", "info")
		cmd.Stdout = ioutil.Discard
		cmd.Stderr = ioutil.Discard
		err := cmd.Run()
		if err == nil {
			break
		}
		time.Sleep(time.Second * 5)
	}

	// Login to Docker
	if len(vargs.Username) != 0 {
		cmd := exec.Command("/usr/bin/docker", "login", "-u", vargs.Username, "-p", vargs.Password, "-e", vargs.Email, vargs.Registry)
		cmd.Dir = workspace.Path
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			fmt.Println("Login failed.")
			os.Exit(1)
		}
	} else {
		fmt.Printf("A username was not specified. Assuming anoynmous publishing.\n")
	}

	// Docker environment info
	cmd := exec.Command("/usr/bin/docker", "version")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	trace(cmd)
	cmd.Run()
	cmd = exec.Command("/usr/bin/docker", "info")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	trace(cmd)
	cmd.Run()

	// Restore from tarred image repository
	if len(vargs.Load) != 0 {
		if _, err := os.Stat(vargs.Load); err != nil {
			fmt.Printf("Archive %s does not exist. Building from scratch.\n", vargs.Load)
		} else {
			cmd := exec.Command("/usr/bin/docker", "load", "-i", vargs.Load)
			cmd.Dir = workspace.Path
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			trace(cmd)
			err := cmd.Run()
			if err != nil {
				os.Exit(1)
			}
		}
	}

	// Build the container
	name := fmt.Sprintf("%s:%s", vargs.Repo, vargs.Tag.Slice()[0])
	cmd = exec.Command("/usr/bin/docker", "build", "--pull=true", "--rm=true", "-f", vargs.File, "-t", name)
	for _, value := range vargs.BuildArgs {
		cmd.Args = append(cmd.Args, "--build-arg", value)
	}
	cmd.Args = append(cmd.Args, vargs.Context)
	cmd.Dir = workspace.Path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	trace(cmd)
	err := cmd.Run()
	if err != nil {
		os.Exit(1)
	}

	// Creates image tags
	for _, tag := range vargs.Tag.Slice()[1:] {
		name_ := fmt.Sprintf("%s:%s", vargs.Repo, tag)
		cmd = exec.Command("/usr/bin/docker", "tag")
		if vargs.ForceTag {
			cmd.Args = append(cmd.Args, "--force=true")
		}
		cmd.Args = append(cmd.Args, name, name_)
		cmd.Dir = workspace.Path
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		trace(cmd)
		err = cmd.Run()
		if err != nil {
			os.Exit(1)
		}
	}

	// Push the image and tags to the registry
	for _, tag := range vargs.Tag.Slice() {
		name_ := fmt.Sprintf("%s:%s", vargs.Repo, tag)
		cmd = exec.Command("/usr/bin/docker", "push", name_)
		cmd.Dir = workspace.Path
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		trace(cmd)
		err = cmd.Run()
		if err != nil {
			os.Exit(1)
		}
	}

	// Save to tarred image repository
	if len(vargs.Save.File) != 0 {
		// if the destination directory does not exist, create it
		dir := filepath.Dir(vargs.Save.File)
		os.MkdirAll(dir, 0755)

		cmd = exec.Command("/usr/bin/docker", "save", "-o", vargs.Save.File)

		// Limit saving to the given tags
		if vargs.Save.Tags.Len() != 0 {
			for _, tag := range vargs.Save.Tags.Slice() {
				name_ := fmt.Sprintf("%s:%s", vargs.Repo, tag)
				cmd.Args = append(cmd.Args, name_)
			}
		} else {
			cmd.Args = append(cmd.Args, vargs.Repo)
		}

		cmd.Dir = workspace.Path
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		trace(cmd)
		err := cmd.Run()
		if err != nil {
			os.Exit(1)
		}
	}
}

// Trace writes each command to standard error (preceded by a ‘$ ’) before it
// is executed. Used for debugging your build.
func trace(cmd *exec.Cmd) {
	fmt.Println("$", strings.Join(cmd.Args, " "))
}

// authorize is a helper function that authorizes the Docker client
// by manually creating the Docker authentication file.
func authorize(d *Docker) error {
	var path = "/root/.dockercfg" // TODO should probably use user.Home() for good measure
	var data = fmt.Sprintf(dockerconf, d.Registry, d.Auth, d.Email)
	return ioutil.WriteFile(path, []byte(data), 0644)
}

var dockerconf = `
{
	"%s": {
		"auth": "%s",
		"email": "%s"
	}
}
`
