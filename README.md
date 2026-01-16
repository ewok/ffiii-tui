# ffiii-tui

[![Tests](https://github.com/ewok/ffiii-tui/actions/workflows/test.yml/badge.svg)](https://github.com/ewok/ffiii-tui/actions/workflows/test.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/ewok/ffiii-tui)](https://golang.org/)
[![License](https://img.shields.io/github/license/ewok/ffiii-tui)](https://github.com/ewok/ffiii-tui/blob/main/LICENSE)
[![Release](https://img.shields.io/github/v/release/ewok/ffiii-tui)](https://github.com/ewok/ffiii-tui/releases)
[![Contributors](https://img.shields.io/github/contributors/ewok/ffiii-tui)](https://github.com/ewok/ffiii-tui/graphs/contributors)
[![Last Commit](https://img.shields.io/github/last-commit/ewok/ffiii-tui)](https://github.com/ewok/ffiii-tui/commits/main)
[![Issues](https://img.shields.io/github/issues/ewok/ffiii-tui)](https://github.com/ewok/ffiii-tui/issues)
[![Pull Requests](https://img.shields.io/github/issues-pr/ewok/ffiii-tui)](https://github.com/ewok/ffiii-tui/pulls)

> **⚠️ Warning**: This project is in early development and may not be fully functional.

![Transactions](images/transactions.png)

A terminal user interface (TUI) for [Firefly III](https://www.firefly-iii.org/) personal finance manager. Manage your finances directly from the terminal with an intuitive interface.

## ✨ Features

- **📊 View and manage** transactions, assets, categories, expenses, and revenue accounts
- **🔍 Search and filter** transactions
- **💰 Real-time insights** with account balances and spending analysis
- **📝 Create transactions** directly from the terminal interface
- **🎨 Clean TUI** built with Charm's Bubble Tea framework

<img src="images/assets.png" alt="Assets" width="200" /> <img src="images/categories.png" alt="Categories" width="200" /> <img src="images/expenses.png" alt="Expenses" width="200" /> <img src="images/revenues.png" alt="Revenues" width="200" />

### Transaction Management

- Create new transactions with guided forms
- View transaction details and splits
- Navigate between different time periods
- Filter by account, category, or search terms

<img src="images/new_transaction.png" alt="New Transaction Form" width="600" />

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or higher
- Access to a Firefly III instance
- Firefly III API key ([How to get an API key](https://docs.firefly-iii.org/how-to/firefly-iii/features/api/#personal-access-tokens))

### Installation

1. **Clone and build:**

   ```bash
   git clone https://github.com/yourusername/ffiii-tui
   cd ffiii-tui
   go mod tidy
   go build
   ```

2. **Initialize configuration:**

   ```bash
   ./ffiii-tui init-config -k YOUR_API_KEY -u https://your-firefly-instance.com/api/v1
   ```

3. **Run the application:**
   ```bash
   ./ffiii-tui
   ```

## 📋 Usage

### Basic Commands

```bash
# Start with default config
./ffiii-tui

# Use custom config file
./ffiii-tui --config path/to/your/config.yaml

# Pass API credentials directly
./ffiii-tui -k YOUR_API_KEY -u https://your-firefly-instance.com/api/v1

# Initialize config file
./ffiii-tui init-config
```

## ⚙️ Configuration

The application uses a YAML configuration file. Generate one with:

```bash
./ffiii-tui init-config
```

### Configuration Options

```yaml
# Required Firefly III API settings
firefly:
  api_key: YOUR_API_KEY # Your Firefly III API token
  api_url: https://your-instance.com/api/v1 # API endpoint URL

# Optional UI settings
ui:
  full_view: false # Full-width transaction view

# Optional logging
logging:
  file: "ffiii-tui.log" # Log file path
```

## 🏗️ Development

### Project Structure

```
ffiii-tui/
├── cmd/                # CLI commands
├── internal/
│   ├── firefly/        # Firefly III API client
│   ├── ui/             # TUI components
│   └── logging/        # Logging utilities
├── config.yaml         # Configuration file
└── main.go             # Entry point
```

### Building from Source

```bash
# Install dependencies
go mod tidy

# Run tests
go test ./...

# Build
go build -o ffiii-tui

# Run with debug logging
./ffiii-tui --debug
```

## 🤝 Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### TODO/Roadmap

- [x] Create new transactions
- [x] Edit existing transactions
- [x] Delete transactions
- [x] Create accounts, categories
- [ ] Edit accounts, categories
- [ ] Delete accounts, categories
- [x] Advanced filtering and search
  - [ ] Not only in current period
- [ ] Budget management
- [ ] Piggy banks
- [ ] Subscriptions
- [ ] Reporting and charts

## 📦 Dependencies

| Package                                                  | Purpose                  |
| -------------------------------------------------------- | ------------------------ |
| [Bubble Tea](https://github.com/charmbracelet/bubbletea) | TUI framework            |
| [Bubbles](https://github.com/charmbracelet/bubbles)      | TUI components           |
| [Lip Gloss](https://github.com/charmbracelet/lipgloss)   | Styling                  |
| [Cobra](https://github.com/spf13/cobra)                  | CLI framework            |
| [Viper](https://github.com/spf13/viper)                  | Configuration management |
| [Zap](https://go.uber.org/zap)                           | Logging                  |

## 📄 License

This project is licensed under the Apache-2.0 License - see the [LICENSE](LICENSE) file for details.
