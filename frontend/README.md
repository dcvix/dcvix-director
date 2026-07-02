<p align="center">
  <img src="../assets/dcvixLogoDarkBG.png" width="300" alt="Logo">
</p>

dcvix Director frontend
=======================
This is the admin frontend application for the dcvix Director

## Technology stack

- **React** as a GUI framework with modern hook and context
- **React Router** for navigation
- **Vite** as a build tool for fast development
- **Axios** for API communication
- **Bulma** for responsive styling
- **Font Awesome** for icons and controls 

## Getting Started

### Dev env setup on linux or WSL

- Download and install nvm:
```bash
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.5/install.sh | bash
```
- setup environment or just restart the shell
```bash
\. "$HOME/.nvm/nvm.sh"
```
- Download and install Node.js:
```bash
nvm install 24
```
- Verify the Node.js version:
```bash
node -v # Should print node version.
```
- Verify npm version:
```bash
npm -v # Should print npn version.
```

### Prerequisites

- Node.js (v22 or higher)
- npm

### Installation and Development

1.  **Navigate to the frontend directory**:
    ```bash
    cd frontend
    ```

2.  **Install dependencies**:
    ```bash
    npm install
    npm install axios react-router-dom bulma @fortawesome/fontawesome-svg-core @fortawesome/free-solid-svg-icons @fortawesome/react-fontawesome
    ```

3.  **Run the development server**:
    ```bash
    npm run dev
    ```
    The application will be available at `http://127.0.0.1:5173` (or the next available port).

## Available Scripts

- `npm run dev`: Starts the development server with Hot Module Replacement (HMR).
- `npm run build`: Compiles and bundles the application for production into the `dist` folder.
- `npm run preview`: Serves the production build locally to preview it.

## Note

- App scaffolding created with vite:
  ```bash
  npm create vite@latest frontend -- --template react-ts
  rm  frontend/App.css
  mkdir -p frontend/src/assets frontend/src/components frontend/src/contexts frontend/src/pages frontend/src/services
  ```
- keep dev tools open with "no cache" option selected on your browser!
