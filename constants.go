/*
 * SPDX-License-Identifier: GPL-3.0
 * Velocity Installer, a cross platform gui/cli app for installing Velocity
 * Copyright (c) 2025 Velocitcs and Velocity contributors
 */

package main

import (
    "image/color"
    "velocityinstaller/buildinfo"
)

const ReleaseUrl = "https://api.github.com/repos/Velocitcs/Velocity/releases/latest"
const InstallerReleaseUrl = "https://api.github.com/repos/Velocitcy/Installer/releases/latest"

var UserAgent = "VelocityInstaller/" + buildinfo.InstallerGitHash + " (https://github.com/Velocitcy/Installer)"

var (
	DiscordGreen  = color.RGBA{R: 0x2D, G: 0x7C, B: 0x46, A: 0xFF}
	DiscordRed    = color.RGBA{R: 0xEC, G: 0x41, B: 0x44, A: 0xFF}
	DiscordBlue   = color.RGBA{R: 0x58, G: 0x65, B: 0xF2, A: 0xFF}
	DiscordYellow = color.RGBA{R: 0xfe, G: 0xe7, B: 0x5c, A: 0xff}
)

var LinuxDiscordNames = []string{
	"Discord",
	"DiscordPTB",
	"DiscordCanary",
	"DiscordDevelopment",
	"discord",
	"discordptb",
	"discordcanary",
	"discorddevelopment",
	"discord-ptb",
	"discord-canary",
	"discord-development",
	// Flatpak
	"com.discordapp.Discord",
	"com.discordapp.DiscordPTB",
	"com.discordapp.DiscordCanary",
	"com.discordapp.DiscordDevelopment",
}
