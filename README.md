# game2d: A 2D Gaming Framework

<div align="center">

![game2d](https://img.shields.io/badge/game2d-Framework-646cff)
![License](https://img.shields.io/badge/license-GPL--3.0-blue.svg)
![Status](https://img.shields.io/badge/status-active-success.svg)

</div>

## 📋 Overview

game2d is an open-source framework for 2D game development that combines Go,
the ebitengine game engine, Lua scripting, and a WebAssembly client. Built on a
declarative object schema, representable in JSON or YAML, it's designed for
experimenting with game code and assets created with generative AI.

### 🏗️ Architecture Overview

```mermaid
graph TD
    subgraph "API Layer"
        API["game2d-api Service"] --> Auth["Authentication & Authorization"]
        API --> GameManagement["Game Management"]
        API --> AIService["AI Integration Service"]
    end

    subgraph "Storage Layer"
        GameManagement --> MongoDB["MongoDB (Game Definitions)"]
        GameManagement --> Redis["Redis (Caching)"]
    end

    subgraph "Client Layer"
        Client["game2d WASM Client"] --> Ebitengine["Ebitengine (2D Game Engine)"]
        Client --> LuaInterpreter["Lua Interpreter"]
        Client --> API
    end
    
    subgraph "Frontend Layer"
        WebUI["game2d-app React Web UI"] --> API
        WebUI --> Client
    end

    AIService --> ExternalAI["External AI Services"]
```

## 🧩 Core Components

### 1. game2d Client

The main client and protocol, typically compiled into WebAssembly (WASM) for
browser integration, but also buildable for various native architectures.

- **Engine**: Combines Ebitengine 2D game engine with a Lua interpreter
- **State Management**: Shared game state between Lua and Go code
- **Game Loop**: Renders assets during draw phase, executes scripts during update phase
- **Game Schema**: Uses a declarative object schema for game definitions

### 2. game2d API Service

The distributed backend service providing REST API functionality:

- **Game Storage**: Persists game state definitions in MongoDB
- **Performance**: Implements Redis caching for fast access
- **AI Integration**: Interfaces with generative AI services to create game content
- **User Management**: Handles authentication, accounts, and profiles
- **Content Delivery**: Serves the web application UI

### 3. game2d App (Web UI)

A React/Vite application for interacting with the service:

- **Game Management**: Search, browse, edit, and save game definitions
- **Account Management**: Update profiles and account settings
- **AI Interaction**: Submit game definitions and prompts to AI services
- **Game Player**: Runs the WASM client directly in the browser

## 🚀 Getting Started

### Prerequisites

- [Go](https://go.dev/dl/) (1.18+)
- [Docker](https://docs.docker.com/get-docker/) & [Docker Compose](https://docs.docker.com/compose/install/)
- [Make](https://www.gnu.org/software/make/)
- [Node.js](https://nodejs.org/) (16+) & [npm](https://www.npmjs.com/)
- [TypeScript](https://www.typescriptlang.org/)
- [React](https://react.dev/)

## 🔧 Development Workflow

1. **Clone the repository**
   ```sh
   git clone https://github.com/dhaifley/game2d.git
   cd game2d
   ```

2. **Set up the development environment**
   ```sh
   cat <<EOF > .env
   LOG_LEVEL='debug'
   DB_CONNECTION='mongodb://game2d:mongodb@localhost:27017/game2d?authSource=admin' 
   CACHE_SERVERS='localhost:6379'
   SUPERUSER=admin
   SUPERUSER_PASSWORD=admin
   AI_API_KEY={your AI API key}
   EOF

   set -a
   . .env
   ```

   You can change any of these defaults, but these values will work with the
   defaults set in the Docker Compose configuration used to run and test
   the services.

3. **Run the services locally**
   ```sh
   make run
   ```

4. **Access the application**
   - Web UI: [http://localhost:8080/](http://localhost:8080/)
   - API docs: [http://localhost:8080/api/v1/docs](http://localhost:8080/api/v1/docs)

5. **Stop and cleanup the services**
   ```sh
   make stop
   ```
6. **Run all tests locally**
   ```sh
   make tests
   ```

## 📖 Documentation

While the service is running locally:

- **API Documentation**: Swagger UI at [http://localhost:8080/api/v1/docs](http://localhost:8080/api/v1/docs)
  
## 🎮 Game Definition Schema

game2d uses a declarative schema for defining games:

- **Game State**: Overall game properties and metadata
- **Assets**: Images and resources needed for rendering
- **Scripts**: Lua code executed during the game loop
- **Objects**: Compositions of assets, scripts, and data
- **Subject**: The special object representing the player

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the GPL-3.0 License - see the LICENSE file for
details.

## 📬 Contact

Project Link: [https://github.com/dhaifley/game2d](https://github.com/dhaifley/game2d)
