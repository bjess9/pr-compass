# PR Pilot ✈️

**PR Pilot** is a terminal-based tool that helps you stay on top of incoming pull requests across multiple GitHub repositories by listing them in order from newest to oldest. Perfect for keeping an eye on your team's PR activity, PR Pilot makes it easy to navigate and open pull requests from a single interface.

## Features

- **Unified PR View:** See open pull requests from all configured repositories, sorted with the newest PRs at the top, allowing you to quickly review recent activity across multiple projects.
- **Navigate PRs:** Use keyboard arrows or `j/k` to move through PRs.
- **Open PRs:** Press `Enter` to open the selected PR in your default browser.
- **Quit:** Press `q` or `Ctrl+C` to exit the application.

## Getting Started

1. **Clone the Repository**

    ```bash
    git clone https://github.com/bjess9/pr-pilot.git
    cd pr-pilot
    ```

2. **Configure Repositories**

    Edit `config.yaml` to include the repositories you want to track in the `owner/repo` format:

    ```yaml
    repos:
      - owner1/repo1
      - owner2/repo2
    ```

3. **Run the Application**

    To start the application, use:

    ```bash
    go run cmd/prpilot/main.go
    ```

4. **Authentication**

    PR Pilot uses GitHub’s OAuth Device Flow to authenticate. When you run the app for the first time, you’ll be prompted to authenticate:

    - **Follow the Authentication Prompt:**  
      The application will display a URL and a one-time code. Open the URL in your browser, enter the provided code, and authorize the application.

    - **Token Storage:**  
      After authentication, your access token will be securely saved in a file located in your home directory (`~/.prpilot_token`). This token will be used for subsequent runs, so you won’t need to re-authenticate each time.

## Usage

- **Track New PRs:** See incoming PRs in order of newest to oldest, so you can review recent work across repositories.
- **Navigate PRs:** Use `↑/↓` arrow keys or `j/k` to move up and down the list.
- **Open PR:** Press `Enter` to open the selected PR in your browser.
- **Quit:** Press `q` or `Ctrl+C` to exit the application.
