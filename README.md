# SFTUI

> [!IMPORTANT]
> This is a work in progress. The project is currently under development and may not be fully functional yet.
> It is also a highly experimental project, so expect changes and improvements over time.

A TUI wrapper of [Silverfin CLI](https://github.com/silverfin/silverfin-cli)

[![Demo](https://github.com/user-attachments/assets/e8a2c07a-f16e-48a5-9bc6-30f664d19fe8)](https://github.com/user-attachments/assets/7bb39656-d1eb-4de1-a1c0-70f89c2af702)

## Why a TUI?

The main reason for creating `sftui` was actually because I wanted to build a TUI, and I wanted to experiment with Go and the Bubble Tea freamework. But why a TUI for Silverfin templates? Because the Silverfin CLI is a project which I'm familiar with, and I though that it could benefit from the interactive that a TUI can provide.

## Overview

`sftui` provides an intuitive interface for browsing and managing Silverfin templates in your local repository. It displays templates organized by category and allows you to view detailed configuration information, select multiple templates, and search through your template collection.

## Features

### Template Management
- **Multi-selection**: Select multiple templates using the space key
- **Bulk Operations**: Deselect all templates at once with backspace
- **Action System**: Trigger actions on selected templates with enter key
- **Template Details**: View and modify complete configuration and metadata for each template

### Search & Navigation
- **Fuzzy Search**: Press `/` to search templates by name, category, or path
- **Smart Filtering**: Real-time filtering as you type
- **Vim-like Navigation**: Use `h/j/k/l` or arrow keys for navigation
- **Section Navigation**: Navigate between different UI sections using Tab/Shift+Tab or Shift+Arrow keys

### Configuration Integration
- **Silverfin Config**: Automatically loads firm and host information from Silverfin CLI configuration files.
- **Repository Detection**: Detects current repository and associated firm information.
- **Template Discovery**: Scans repository structure for templates.

