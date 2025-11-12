# SSL Checker Frontend

A modern React.js frontend application for SSL certificate checking, built with Vite, TypeScript, and Material-UI.

## Features

- ✅ React 19 with TypeScript for type safety
- ✅ Vite for fast development and optimized production builds
- ✅ Material-UI for consistent, responsive design
- ✅ React Router for client-side routing
- ✅ Axios for API calls
- ✅ Code splitting with React.lazy for optimization
- ✅ Multilingual support (English/Vietnamese)
- ✅ Accessibility-focused with ARIA labels
- ✅ Vitest for component testing

## Tech Stack

- **React**: ^19.2.0
- **TypeScript**: ~5.9.3
- **Vite**: ^7.2.2
- **Material-UI**: ^6.0.0
- **React Router**: ^7.9.5
- **Axios**: ^1.13.2
- **Vitest**: ^4.0.8

## Getting Started

### Prerequisites

- Node.js >= 20
- npm >= 10

### Installation

```bash
npm install
```

### Development

Start the development server:

```bash
npm run dev
```

The application will be available at http://localhost:3000

### Building

Build for production:

```bash
npm run build
```

Preview production build:

```bash
npm run preview
```

### Testing

Run tests:

```bash
npm run test
```

Run tests in watch mode:

```bash
npm run test:ui
```

### Linting

```bash
npm run lint
```

## Project Structure

```
frontend/
├── src/
│   ├── components/         # Reusable React components
│   │   ├── Navigation.tsx
│   │   ├── SSLCheckForm.tsx
│   │   └── SSLResultDisplay.tsx
│   ├── pages/             # Page components
│   │   ├── Login.tsx
│   │   ├── Dashboard.tsx
│   │   └── AddDomain.tsx
│   ├── hooks/             # Custom React hooks
│   │   ├── LanguageContext.tsx
│   │   └── useLanguage.tsx
│   ├── services/          # API services
│   │   └── api.ts
│   ├── types/             # TypeScript type definitions
│   │   └── index.ts
│   ├── utils/             # Utility functions
│   │   └── translations.ts
│   ├── tests/             # Test files
│   │   └── SSLCheckForm.test.tsx
│   ├── App.tsx            # Main application component
│   ├── main.tsx           # Application entry point
│   └── setupTests.ts      # Test setup
├── public/                # Static assets
├── Dockerfile            # Docker configuration
├── nginx.conf            # Nginx configuration for production
├── vite.config.ts        # Vite configuration
└── package.json          # Project dependencies and scripts
```

## Features Implementation

### Pages

1. **Login** - Placeholder page for authentication (future implementation)
2. **Dashboard** - Overview with statistics and recent checks (hardcoded data)
3. **Add Domain** - SSL certificate checking interface

### Components

- **Navigation**: Top navigation bar with routing and language toggle
- **SSLCheckForm**: Form for entering domain/IP to check
- **SSLResultDisplay**: Display SSL certificate information

### Accessibility

- ARIA labels on all interactive elements
- Semantic HTML structure
- Keyboard navigation support
- Screen reader friendly

### Internationalization

- English and Vietnamese language support
- Easy to add more languages
- Persistent language selection

## Docker Deployment

Build Docker image:

```bash
docker build -t ssl-checker-frontend .
```

Run container:

```bash
docker run -p 80:80 ssl-checker-frontend
```

## API Integration

The frontend proxies API requests to the SSL Checker backend at `/api`. Configure the backend URL in `vite.config.ts` (development) or `nginx.conf` (production).

## Browser Support

- Chrome (latest)
- Firefox (latest)
- Safari (latest)
- Edge (latest)

## License

This project is part of the docker-images repository.
