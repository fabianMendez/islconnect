package main

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
)

func unzip(src string) (*os.File, error) {
	zipReader, err := zip.OpenReader(src)
	if err != nil {
		return nil, err
	}
	defer zipReader.Close()

	zipFile := zipReader.File[0]

	rc, err := zipFile.Open()
	if err != nil {
		return nil, err
	}

	defer rc.Close()

	f, err := ioutil.TempFile("", fmt.Sprintf("*_%s", zipFile.Name))
	if err != nil {
		return nil, err
	}

	os.Chmod(f.Name(), zipFile.Mode())
	if err != nil {
		return nil, err
	}

	defer f.Close()

	_, err = io.Copy(f, rc)
	if err != nil {
		if err := os.Remove(f.Name()); err != nil {
			panic(err)
		}

		return nil, err
	}

	return f, nil
}

func detectPlatform(os, arch string) string {
	switch os {
	case "linux":
		if arch == "amd64" {
			return "linux64"
		}

		return "linux"
	case "windows":
		if arch == "amd64" {
			return "win64"
		}

		return "win32"
	case "darwin":
		if arch == "amd64" {
			return "mac"
		}

		return "macppc"

	default:
		return os
	}
}

func main() {
	if len(os.Args) == 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s <session>\n", os.Args[0])
		os.Exit(1)
	}

	session := os.Args[1]
	platform := detectPlatform(runtime.GOOS, runtime.GOARCH)

	u := fmt.Sprintf(`https://isllight.islonline.net/start/ISLLightClient?addbase=+%s&cmdline=--auto-close+--connect+%s&platform=%s`,
		session, session, platform)

	resp, err := http.Get(u)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	tmpZip, err := ioutil.TempFile("", "*.zip")
	if err != nil {
		log.Fatal(err)
	}

	defer os.Remove(tmpZip.Name())

	log.Println("[*] Downloading to:", tmpZip.Name())

	_, err = io.Copy(tmpZip, resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("[*] Unzipping...")

	tmpBin, err := unzip(tmpZip.Name())
	if err != nil {
		log.Fatal(err)
	}

	defer os.Remove(tmpBin.Name())

	log.Println("[*] Executing", tmpBin.Name())

	cmd := exec.Command(tmpBin.Name())
	// cmd := exec.Command("ls", "-l", tmpBin.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}
