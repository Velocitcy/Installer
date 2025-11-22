/*
 * SPDX-License-Identifier: GPL-3.0
 * Velocity Installer, a cross platform gui/cli app for installing Velocity
 * Copyright (c) 2025 Velocitcs and Velocity contributors
 */

package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"runtime"
	"time"
	"velocityinstaller/buildinfo"
)

var IsSelfOutdated = false
var SelfUpdateCheckDoneChan = make(chan bool, 1)

func init() {
	if buildinfo.InstallerTag == buildinfo.VersionUnknown {
		Log.Debug("Disabling self updater because this is not a release build")
		return
	}

	go DeleteOldExecutable()

	go func() {
		Log.Debug("Checking for Installer Updates...")

		res, err := GetGithubRelease(InstallerReleaseUrl)
		if err != nil {
			Log.Warn("Failed to check for self updates:", err)
			SelfUpdateCheckDoneChan <- false
			return
		}

		IsSelfOutdated = res.TagName != buildinfo.InstallerTag
		Log.Debug("Is self outdated?", IsSelfOutdated)
		SelfUpdateCheckDoneChan <- true
	}()
}

func GetInstallerDownloadLink() string {
	const BaseUrl = "https://github.com/Velocitcy/Installer/releases/latest/download/"
	switch runtime.GOOS {
	case "windows":
		filename := Ternary(buildinfo.UiType == buildinfo.UiTypeCli, "VelocityInstallerCli.exe", "VelocityInstaller.exe")
		return BaseUrl + filename
	case "darwin":
		return BaseUrl + "VelocityInstaller.MacOS.zip"
	case "linux":
		return BaseUrl + "VelocityInstallerCli-linux"
	default:
		return ""
	}
}

func CanUpdateSelf() bool {
	return IsSelfOutdated && runtime.GOOS != "darwin"
}

func UpdateSelf() error {
	if !CanUpdateSelf() {
		return errors.New("cannot update self. Either no update available or macos")
	}

	url := GetInstallerDownloadLink()
	if url == "" {
		return errors.New("failed to get installer download link")
	}

	Log.Debug("Updating self from", url)

	ownExePath, err := os.Executable()
	if err != nil {
		return err
	}

	ownExeDir := path.Dir(ownExePath)

	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	tmp, err := os.CreateTemp(ownExeDir, "VelocityInstallerUpdate")
	if err != nil {
		return fmt.Errorf("failed to create tempfile: %w", err)
	}
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
	}()

	if err = tmp.Chmod(0o755); err != nil {
		return fmt.Errorf("failed to chmod 755 %s: %w", tmp.Name(), err)
	}

	if _, err = io.Copy(tmp, res.Body); err != nil {
		return err
	}

	if err = tmp.Close(); err != nil {
		return err
	}

	if err = os.Remove(ownExePath); err != nil {
		if err = os.Rename(ownExePath, ownExePath+".old"); err != nil {
			return fmt.Errorf("failed to remove or rename own executable: %w", err)
		}
	}

	if err = os.Rename(tmp.Name(), ownExePath); err != nil {
		return fmt.Errorf("failed to replace self with updated executable. Please manually redownload the installer: %w", err)
	}

	return nil
}

func DeleteOldExecutable() {
	ownExePath, err := os.Executable()
	if err != nil {
		return
	}

	for attempts := 0; attempts < 10; attempts++ {
		err = os.Remove(ownExePath + ".old")

		if err == nil || errors.Is(err, os.ErrNotExist) {
			break
		}

		Log.Warn("Failed to remove old executable. Retrying in 1 second.", err)
		time.Sleep(1 * time.Second)
	}
}

func RelaunchSelf() error {
	attr := new(os.ProcAttr)
	attr.Files = []*os.File{os.Stdin, os.Stdout, os.Stderr}

	var argv []string
	if len(os.Args) > 1 {
		argv = os.Args[1:]
	} else {
		argv = []string{}
	}

	Log.Debug("Restarting self with exe", os.Args[0], "and args", argv)

	proc, err := os.StartProcess(os.Args[0], argv, attr)
	if err != nil {
		return fmt.Errorf("failed to start new process: %w", err)
	}

	if err = proc.Release(); err != nil {
		return fmt.Errorf("failed to release new process: %w", err)
	}

	os.Exit(0)
	return nil
}
