# lazydevops

A terminal UI for browsing and managing Azure DevOps pipelines, built with [Bubble Tea](https://github.com/charmbracelet/bubbletea). Navigate projects, pipelines, and runs with vim-style keybindings — no browser required.

## Features

- **Vim-style navigation** — `j`/`k` movement, `gg`/`G` jumps, `ctrl+d`/`ctrl+u` half-page scroll
- **Dual input modes** — normal mode for navigation, insert mode for search
- **Real-time search** — filter projects, pipelines, and runs as you type
- **Advanced run filtering** — filter by result status (`result:failed api`), quick-filter with `alt+f`
- **Run details viewer** — stages, jobs, tasks, and logs in a multi-panel layout
- **Command palette** — `:` to access commands (`:q`, `:set`, `:help`, `:refresh`, etc.)
- **Setup wizard** — guided first-run setup with prerequisite checks
- **Fully configurable** — colors, status icons, input mode, display options
- **CLI command preview** — status bar shows the equivalent `az devops` command for current context

## Prerequisites

- **Go 1.25+** (for building from source)
- **Azure CLI** installed and authenticated (`az login`)
- **Azure DevOps CLI extension** (`az extension add --name azure-devops`)
- **Default organization configured** (`az devops configure --defaults organization=https://dev.azure.com/your-org`)

## Installation

### From GitHub Releases

Download the latest binary for your platform from the [Releases](https://github.com/sudominds/lazydevops/releases) page.

### From source

```bash
git clone https://github.com/sudominds/lazydevops.git
cd lazydevops
go build -o lazydevops .
./lazydevops
```

> **Note:** `go install` from GitHub won't work because the module path is `lazydevops`, not a fully qualified GitHub URL. Use clone + build instead.

## Configuration

Copy the example config and edit it:

```bash
mkdir -p ~/.config/lazydevops
cp config.example.json ~/.config/lazydevops/config.json
```

| Setting | Description |
|---------|-------------|
| `organization` | Your Azure DevOps organization URL |
| `main_layout.default_input_mode` | Start in `normal` or `insert` mode |
| `main_layout.show_path_in_title` | Show breadcrumb path in the title bar |
| `main_layout.list_highlight_mode` | Highlight style for list items |
| `main_layout.default_project` | Skip project selection and jump to this project |
| `search.match_highlight_mode` | How search matches are highlighted (`accent` or `off`) |
| `log_rendering.mode` | Log rendering mode (`auto`) |
| `palette.*` | Full color customization for all UI elements |
| `run_status_icons.*` | Customize status icons and display mode |
| `run_status_colors.*` | Customize colors per pipeline status |

You can also edit settings interactively with the `:set` command.

## Key Bindings

### Normal Mode

| Key | Action |
|-----|--------|
| `j` / `k` | Move down / up |
| `enter` / `l` | Select / enter |
| `esc` / `b` | Go back |
| `r` | Refresh |
| `s` / `i` / `/` | Open search (switch to insert mode) |
| `?` | Show help / keymaps |
| `:` | Open command palette |
| `alt+f` | Filter runs by result status |
| `ctrl+c` | Quit |

### Insert Mode (Search)

| Key | Action |
|-----|--------|
| Type | Filter the current list |
| `esc` / `jk` | Return to normal mode |
| `enter` | Select highlighted item |

### Run Details

| Key | Action |
|-----|--------|
| `j` / `k` | Scroll up / down |
| `tab` / `h` / `l` | Switch between sections (info, tree, log) |
| `0` / `1` / `2` | Jump to section |
| `ctrl+d` / `ctrl+u` | Half-page scroll down / up |
| `gg` / `G` | Jump to top / bottom |
| `b` / `esc` | Go back |

### Command Palette

| Command | Action |
|---------|--------|
| `:q` / `:quit` / `:exit` | Quit |
| `:r` / `:refresh` | Refresh current view |
| `:set` | Open configuration wizard |
| `:help` / `:?` | Show help |
| `:i` / `:insert` | Switch to insert mode |
| `:n` / `:normal` | Switch to normal mode |
| `:b` / `:back` | Go back |

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b my-feature`)
3. Make your changes and test (`go build ./...` && `go vet ./...`)
4. Commit and push
5. Open a pull request

## License

TBD
