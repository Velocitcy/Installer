/*
 * SPDX-License-Identifier: GPL-3.0
 * Velocity Installer, a cross platform gui/cli app for installing Velocity
 * Copyright (c) 2025 Velocitcs and Velocity contributors
 */

package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	path "path/filepath"
	"regexp"
	"strings"
)

type GithubRelease struct {
	Name    string `json:"name"`
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name        string `json:"name"`
		DownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

var ReleaseData GithubRelease
var GithubError error
var GithubDoneChan chan bool

var InstalledHash = "None"
var LatestHash = "Unknown"
var IsDevInstall bool

func GetGithubRelease(url string) (*GithubRelease, error) {
	Log.Debug("Fetching", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		Log.Error("Failed to create Request", err)
		return nil, err
	}

	req.Header.Set("User-Agent", UserAgent)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		Log.Error("Failed to send Request", err)
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode >= 300 {
		err = errors.New(res.Status)
		Log.Error(url, "returned Non-OK status", err)
		return nil, err
	}

	var data GithubRelease

	if err = json.NewDecoder(res.Body).Decode(&data); err != nil {
		Log.Error("Failed to decode GitHub JSON Response", err)
		return nil, err
	}

	return &data, nil
}

func InitGithubDownloader() {
	GithubDoneChan = make(chan bool, 1)

	IsDevInstall = os.Getenv("VELOCITY_DEV_INSTALL") == "1"
	Log.Debug("Is Dev Install:", IsDevInstall)

	if IsDevInstall {
		GithubDoneChan <- true
		return
	}

	go func() {
		defer func() {
			GithubDoneChan <- GithubError == nil
		}()

		data, err := GetGithubRelease(ReleaseUrl)
		if err != nil {
			GithubError = err
			return
		}

		ReleaseData = *data

		i := strings.LastIndex(data.Name, " ") + 1
		LatestHash = data.Name[i:]

		Log.Debug("Finished fetching GitHub Data")
		Log.Debug("Latest hash is", LatestHash, "Local Install is", Ternary(LatestHash == InstalledHash, "up to date!", "outdated!"))
	}()

	VelocityFile := VelocityDirectory

	stat, err := os.Stat(VelocityFile)
	if err != nil {
		return
	}

	if stat.IsDir() {
		VelocityFile = path.Join(VelocityFile, "main.js")
	}

	b, err := os.ReadFile(VelocityFile)
	if err != nil {
		return
	}

	Log.Debug("Found existing Velocity Install. Checking for hash...")

	re := regexp.MustCompile(`// Velocity (\w+)`)
	match := re.FindSubmatch(b)
	if match != nil {
		InstalledHash = string(match[1])
		Log.Debug("Existing hash is", InstalledHash)
	} else {
		Log.Debug("Did not find hash")
	}
}

func installLatestBuilds() (retErr error) {
	Log.Debug("Installing latest builds...")

	if IsDevInstall {
		Log.Debug("Skipping due to dev install")
		return
	}

	// Create parent directories
	if err := os.MkdirAll(path.Dir(VelocityDirectory), 0755); err != nil {
		Log.Error("Failed to create directories:", err)
		retErr = err
		return
	}

	distDir := path.Dir(VelocityDirectory)

	for _, ass := range ReleaseData.Assets {
		lowerName := strings.ToLower(ass.Name)

		// Skip .LEGAL.txt files
		if strings.HasSuffix(lowerName, ".legal.txt") {
			continue
		}

		// Only download: preload, velocityDesktopPreload, patcher, velocityDesktopMain, velocityDesktopRenderer, renderer
		if strings.Contains(lowerName, "preload") ||
			strings.Contains(lowerName, "patcher") ||
			strings.Contains(lowerName, "velocitydesktopMain") ||
			strings.Contains(lowerName, "velocitydesktoprenderer") ||
			(strings.Contains(lowerName, "renderer") && !strings.Contains(lowerName, "velocitydesktop")) {

			Log.Debug("Downloading", ass.Name)

			res, err := http.Get(ass.DownloadURL)
			if err == nil && res.StatusCode >= 300 {
				err = errors.New(res.Status)
			}
			if err != nil {
				Log.Error("Failed to download "+ass.Name+":", err)
				retErr = err
				return
			}

			filePath := path.Join(distDir, ass.Name)
			out, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				Log.Error("Failed to create "+ass.Name+":", err)
				retErr = err
				return
			}

			if _, err = io.Copy(out, res.Body); err != nil {
				out.Close()
				Log.Error("Failed to write "+ass.Name+":", err)
				retErr = err
				return
			}
			out.Close()

			_ = FixOwnership(filePath)
		}
	}

	InstalledHash = LatestHash
	return
}
