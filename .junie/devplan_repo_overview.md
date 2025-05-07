# High-Level Architectural Overview of the Devplan CLI Repository

## 1. Technologies and Functional Components

- **Programming Language**:  
  - **Go (Golang)**: Core language for implementation, leveraging its strengths in CLI tools, concurrency, and static compilation.

- **CLI Framework & Utilities**:  
  - **Cobra**: For command-line interface structure, command parsing, and subcommands.  
  - **Bubbletea & Lipgloss**: For terminal-based UI components, such as IDE selection prompts.

- **Configuration & Environment Management**:  
  - **Viper**: For handling configuration files, environment variables, and user preferences.

- **Versioning & Self-Update**:  
  - **Semantic Versioning (semver)** via `Masterminds/semver`: For version comparison and update logic.  
  - **Custom updater** (`internal/utils/updater`) manages self-updates, fetching binaries from DigitalOcean Spaces.

- **Protobuf & gRPC**:  
  - **Protobuf**: For defining structured data (e.g., version info, configs).  
  - **Generated code** (`internal/pb/...`) for serialization/deserialization.

- **Repository & Git Utilities**:  
  - **go-git**: For managing local repositories, cloning, updating, and inspecting remotes.

- **HTTP & API Communication**:  
  - **net/http**: For REST API interactions with Devplan backend services.

- **UI & User Interaction**:  
  - **Charmbracelet Bubble**: For terminal UI prompts, selection menus, and interactive workflows.

- **Build & Dependency Management**:  
  - **Makefile**: For orchestrating build, test, generate protobuf, and release processes.  
  - **Go Modules**: For dependency management (`go.mod`).

- **Scripting & Automation**:  
  - Bash scripts (`build_all.sh`, `generate_proto.sh`, `release_this.sh`) for cross-platform build, protobuf generation, and release tagging.

- **Documentation & Metadata**:  
  - Markdown (`README.md`) for user guidance.  
  - GitHub Actions workflows for CI/CD, releases, and version marking.

---

## 2. Key Folder Structure & Their Responsibilities

### **`internal/`**  
- **`cmd/`**: Entry points for CLI commands, organized by functional areas (main, project, auth, ide, self, update, version).  
- **`components/`**: UI components, especially selection menus for companies, features, projects.  
- **`devplan/`**: Core logic for interacting with Devplan backend services, including API clients, domain paths, and domain-specific functions.  
- **`pb/`**: Generated protobuf code for configuration and CLI version info.  
- **`utils/`**: Utility functions for git operations, environment globals, IDE rules, and updater logic.

### **`proto/`**  
- Protocol buffer definitions for version info, configuration schemas, and other structured data.

### **`scripts/`**  
- Build, protobuf generation, release tagging, and testing scripts.

### **`.github/`**  
- CI/CD workflows for release, testing, and marking versions as production.

### **`.cursor/rules/`**  
- Markdown rule files for development workflows, insights collection, and repository overview.

### **`go.mod` & `Makefile`**  
- Dependency management and build orchestration.

---

## 3. Recommended Subfolder & File Responsibilities

### **`internal/cmd/`**  
- **Main CLI commands**: `main.go`, `project/`, `auth.go`, `ide.go`, `self.go`, `update.go`, `version.go`.  
- **Subcommands**: Focused on specific functionalities like project focus, authentication, IDE management, self-info, versioning, and updates.

### **`internal/components/selector/`**  
- UI components for selection menus (company, feature, project).  
- Use `bubbletea` and `list` for terminal-based interactive prompts.

### **`internal/devplan/`**  
- API client logic for communication with Devplan backend services.  
- Path helpers, domain URL management, and core business logic.

### **`internal/utils/`**  
- **`git/`**: Repository info, clone, update, and URL validation.  
- **`globals/`**: Persistent user preferences (last selected company/project).  
- **`ide/`**: IDE rule management, download, and configuration.  
- **`updater/`**: Self-update logic, version checks, binary downloads.

### **`proto/`**  
- Protocol buffer definitions for version info, configuration schemas, and CLI-specific data.

---

## 4. Code Organization & Best Practices Insights

- **Modular Structure**:  
  - Clear separation between CLI command handling (`cmd/`), core logic (`devplan/`), UI components (`components/`), and utilities (`utils/`).  
  - This promotes maintainability, testability, and ease of extension.

- **Configuration & State Management**:  
  - Use of `viper` for persistent user preferences (default IDE, last project).  
  - Environment variables and config files are externalized, enabling flexible deployment.

- **UI & User Interaction**:  
  - Terminal UI components leverage `bubbletea` for interactive prompts, such as IDE selection, project focus, and feature workflows.

- **Protobuf & Data Serialization**:  
  - Strongly typed data structures for versioning and configuration, ensuring backward compatibility and structured data exchange.

- **CI/CD & Release Automation**:  
  - GitHub workflows automate building, testing, protobuf generation, and release tagging.  
  - Scripts facilitate cross-platform build and release management.

- **Security & Best Practices**:  
  - `.gitignore` excludes sensitive or build artifacts.  
  - Self-updater fetches binaries securely from DigitalOcean Spaces, with version comparison to prevent regressions.

- **Extensibility & Documentation**:  
  - Markdown rule files (`devplan_flow.mdc`, `insights.mdc`, `repo_overview.mdc`) provide guidelines for development workflows and insights collection, fostering consistent practices.

---

**Summary**:  
The Devplan CLI repository is a well-structured, modular Go application designed for automation of development workflows. It combines CLI commands, terminal UI, backend API interactions, and self-updating capabilities, all orchestrated through clear folder organization, protobuf schemas, and CI/CD pipelines. The architecture emphasizes extensibility, maintainability, and user configurability, making it suitable for complex development environments with integrated IDE and repository management.