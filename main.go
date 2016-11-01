package main

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fsouza/go-dockerclient"
)

func mountedPath(path string) string {
	return filepath.Join("/host", path)
}

func main() {
	dockerCli, err := docker.NewClient("unix:///host/var/run/docker.sock")
	if err != nil {
		panic(err)
	}

	fileToDownload := "/etc/hosts"
	tarStream := getTarStream(dockerCli, fileToDownload)
	defer tarStream.Close()

	fmt.Println("Downloading", fileToDownload, "as tar")

	tarReader := tar.NewReader(tarStream)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			panic(err)
		}

		fileInfo := header.FileInfo()
		// Should also check if file is a link in real life
		if fileInfo.IsDir() {
			fmt.Println("Got dir", header.Name)
			continue
		}

		fmt.Println("Got file", header.Name, "with size", header.Size)
		fmt.Println("=========Contents=========")
		if _, err = io.CopyN(os.Stdout, tarReader, header.Size); err != nil {
			panic(err)
		}
		fmt.Println("==========================")
	}
	fmt.Println("Download complete")
}

func getTarStream(cli *docker.Client, filename string) io.ReadCloser {
	container := createContainer(cli, filename)
	fmt.Println("Created container", container.ID)
	defer removeContainer(cli, container)

	startContainer(cli, container)
	fmt.Println("Started container")

	preader, pwriter := io.Pipe()
	opts := docker.DownloadFromContainerOptions{
		Path:         mountedPath(filename),
		OutputStream: pwriter,
	}

	// Let docker asynchronously write into the pipe while we are reading it on the other end
	go func() {
		defer pwriter.Close()
		fmt.Println("Requesting file", opts.Path)
		if err := cli.DownloadFromContainer(container.ID, opts); err != nil {
			panic(err)
		}
	}()

	return preader
}

func createContainer(cli *docker.Client, filename string) *docker.Container {
	createOpts := docker.CreateContainerOptions{
		Config: &docker.Config{
			// any image that has the command specified below can be used
			Image: "dockercp",
			// "cat" with no arguments will simply block indefinitely ensuring that the container does not terminate.
			Cmd:        []string{"cat"},
			Entrypoint: []string{},
			OpenStdin:  true,
		},
		HostConfig: &docker.HostConfig{
			Binds: []string{fmt.Sprintf("%s:%s", filename, mountedPath(filename))},
		},
	}

	container, err := cli.CreateContainer(createOpts)
	if err != nil {
		panic(err)
	}
	return container
}

func startContainer(cli *docker.Client, container *docker.Container) {
	if err := cli.StartContainer(container.ID, nil); err != nil {
		panic(err)
	}
}

func removeContainer(cli *docker.Client, container *docker.Container) {
	removeOpts := docker.RemoveContainerOptions{
		ID:            container.ID,
		RemoveVolumes: true,
		Force:         true,
	}

	err := cli.RemoveContainer(removeOpts)
	if err != nil {
		if _, ok := err.(*docker.NoSuchContainer); ok {
			return
		}
		panic(err)
	}
}
