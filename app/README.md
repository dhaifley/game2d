# game2d-app
Simple user interface for game2d.ai

This is a very basic React + Vite app used to implement a simple user interface
for the game2d.ai service.

It provides a basic web service navigation structure. It provides sign in and
user profile functionality and access to settings through a top nav bar.

It displays a list of available game2d game definitions provided by the
game2d-api back end service. It displays a longer description of the currently
selected game definition below the list view.

It centrally displays a large canvas for rendering the WASM version of the
game2d client, which is running the currently selected game definition.

It also displays a prompt window for interacting with a generative AI service.
The game2d-api back end can submit a prompt to an AI service, along with the
current game state, and information about the game engine, and have the AI
service generate new game state. The prompt window is used to collect user
prompts, and display dialog history for the currently selected game definition.

## Getting Started

### Requirements

- [node.js](https://nodejs.org/)
- [npm](https://www.npmjs.com/)
- [typescript](https://www.typescriptlang.org/)
- [react](https://react.dev/)
- [react router](https://reactrouter.com/)
- [vite](https://vite.dev/)

### Installation

1. Change into this app directory:
   ```sh
   cd game2d/app
   ```

2. Install dependencies:
   ```sh
   npm install
   ```

### Development

To run the application in development mode:

```sh
npm run dev
```

This will start the development server, usually at http://localhost:5173

### Building

To create a build:

```sh
npm run build
```

### Running

To preview the build locally:

```sh
npm run preview
```

### Connecting to game2d-api

Make sure the game2d-api backend service is running and properly configured to
connect with this frontend application.

