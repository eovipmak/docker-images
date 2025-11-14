import React from 'react';
import {
  AppBar,
  Toolbar,
  Typography,
  Button,
  IconButton,
  Box,
  Container,
} from '@mui/material';
import {
  Language as LanguageIcon,
  Dashboard as DashboardIcon,
  AddCircle as AddCircleIcon,
  Login as LoginIcon,
  Logout as LogoutIcon,
  NotificationsActive as NotificationsActiveIcon,
} from '@mui/icons-material';
import { useNavigate, useLocation } from 'react-router-dom';
import { useLanguage } from '../hooks/useLanguage';
import { useAuth } from '../hooks/useAuth';

const Navigation: React.FC = () => {
  const { language, setLanguage, t } = useLanguage();
  const { isAuthenticated, logout, user } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  const toggleLanguage = () => {
    setLanguage(language === 'en' ? 'vi' : 'en');
  };

  const handleLogout = async () => {
    await logout();
    navigate('/login');
  };

  const isActive = (path: string) => location.pathname === path;

  return (
    <AppBar position="static">
      <Container maxWidth="lg">
        <Toolbar disableGutters>
          <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
            {t('title')}
          </Typography>
          <Box sx={{ display: 'flex', gap: 1, alignItems: 'center' }}>
            {isAuthenticated ? (
              <>
                {user && (
                  <Typography variant="body2" sx={{ mr: 1 }}>
                    {user.email}
                  </Typography>
                )}
                <Button
                  color="inherit"
                  startIcon={<DashboardIcon />}
                  onClick={() => navigate('/dashboard')}
                  sx={{
                    backgroundColor: isActive('/dashboard') ? 'rgba(255, 255, 255, 0.1)' : 'transparent',
                  }}
                  aria-label={t('dashboard')}
                >
                  {t('dashboard')}
                </Button>
                <Button
                  color="inherit"
                  startIcon={<AddCircleIcon />}
                  onClick={() => navigate('/add-domain')}
                  sx={{
                    backgroundColor: isActive('/add-domain') ? 'rgba(255, 255, 255, 0.1)' : 'transparent',
                  }}
                  aria-label={t('addDomain')}
                >
                  {t('addDomain')}
                </Button>
                <Button
                  color="inherit"
                  startIcon={<NotificationsActiveIcon />}
                  onClick={() => navigate('/alert-settings')}
                  sx={{
                    backgroundColor: isActive('/alert-settings') ? 'rgba(255, 255, 255, 0.1)' : 'transparent',
                  }}
                  aria-label="Alert Settings"
                >
                  Alerts
                </Button>
                <Button
                  color="inherit"
                  startIcon={<LogoutIcon />}
                  onClick={handleLogout}
                  aria-label="Logout"
                >
                  Logout
                </Button>
              </>
            ) : (
              <Button
                color="inherit"
                startIcon={<LoginIcon />}
                onClick={() => navigate('/login')}
                sx={{
                  backgroundColor: isActive('/login') ? 'rgba(255, 255, 255, 0.1)' : 'transparent',
                }}
                aria-label={t('login')}
              >
                {t('login')}
              </Button>
            )}
            <IconButton
              color="inherit"
              onClick={toggleLanguage}
              aria-label={`Switch to ${language === 'en' ? 'Vietnamese' : 'English'}`}
            >
              <LanguageIcon />
            </IconButton>
          </Box>
        </Toolbar>
      </Container>
    </AppBar>
  );
};

export default Navigation;
