# PR Pilot ✈️

**PR Pilot** is a terminal-based tool for listing, navigating, and opening GitHub pull requests across multiple repositories.

## Features

- **List PRs:** Displays open pull requests from configured repositories.
- **Navigate:** Use keyboard arrows or `j/k` to move through PRs.
- **Open PRs:** Press `Enter` to open the selected PR in your default browser.
- **Quit:** Press `q` or `Ctrl+C` to exit the application.

## Installation

1. **Clone the Repository:**

    ```bash
    git clone https://github.com/bjess9/pr-pilot.git
    cd pr-pilot
    ```

2. **Run the Application:**

    ```bash
    go run .
    ```

## Configuration

1. **Set GitHub Token:**

    Ensure your GitHub Personal Access Token is set as an environment variable:

    ```bash
    export GITHUB_TOKEN=your_personal_access_token
    ```

2. **Configure Repositories:**

    Edit `config.yaml` to include your repositories in the `owner/repo` format:

    ```yaml
    repos:
      - owner1/repo1
      - owner2/repo2
    ```

## Usage

- **Navigate PRs:** Use `↑/↓` arrow keys or `j/k` to move up and down the list.
- **Open PR:** Press `Enter` to open the selected PR in your browser.
- **Quit:** Press `q` or `Ctrl+C` to exit the application.
