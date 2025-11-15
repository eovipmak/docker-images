import React from 'react';
import {
  AppBar,
  Toolbar,
  Typography,
  Button,
  Box,
  Container,
  Chip,
  Avatar,
} from '@mui/material';
import {
  Dashboard as DashboardIcon,
  AddCircle as AddCircleIcon,
  Login as LoginIcon,
  Logout as LogoutIcon,
  NotificationsActive as NotificationsActiveIcon,
  Security as SecurityIcon,
} from '@mui/icons-material';
import { useNavigate, useLocation } from 'react-router-dom';
import { useLanguage } from '../hooks/useLanguage';
import { useAuth } from '../hooks/useAuth';

const Navigation: React.FC = () => {
  const { t } = useLanguage();
  const { isAuthenticated, logout, user } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  const handleLogout = async () => {
    await logout();
    navigate('/login');
  };

  const isActive = (path: string) => location.pathname === path;

  return (
    <AppBar 
      position="sticky" 
      elevation={0}
      sx={{
        background: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)',
        backdropFilter: 'blur(10px)',
        borderBottom: '1px solid rgba(255, 255, 255, 0.1)',
      }}
    >
      <Container maxWidth="xl">
        <Toolbar disableGutters sx={{ py: 1 }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, flexGrow: 1 }}>
            <SecurityIcon sx={{ fontSize: 28 }} />
            <Typography 
              variant="h6" 
              component="div" 
              sx={{ 
                fontWeight: 700,
                letterSpacing: '-0.5px',
              }}
            >
              SSL Monitor
            </Typography>
          </Box>
          <Box sx={{ display: 'flex', gap: 0.5, alignItems: 'center' }}>
            {isAuthenticated ? (
              <>
                <Button
                  color="inherit"
                  startIcon={<DashboardIcon />}
                  onClick={() => navigate('/dashboard')}
                  sx={{
                    backgroundColor: isActive('/dashboard') ? 'rgba(255, 255, 255, 0.2)' : 'transparent',
                    borderRadius: 2,
                    px: 2,
                    py: 1,
                    '&:hover': {
                      backgroundColor: 'rgba(255, 255, 255, 0.15)',
                    },
                    transition: 'all 0.2s',
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
                    backgroundColor: isActive('/add-domain') ? 'rgba(255, 255, 255, 0.2)' : 'transparent',
                    borderRadius: 2,
                    px: 2,
                    py: 1,
                    '&:hover': {
                      backgroundColor: 'rgba(255, 255, 255, 0.15)',
                    },
                    transition: 'all 0.2s',
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
                    backgroundColor: isActive('/alert-settings') ? 'rgba(255, 255, 255, 0.2)' : 'transparent',
                    borderRadius: 2,
                    px: 2,
                    py: 1,
                    '&:hover': {
                      backgroundColor: 'rgba(255, 255, 255, 0.15)',
                    },
                    transition: 'all 0.2s',
                  }}
                  aria-label={t('alertSettings')}
                >
                  {t('alerts')}
                </Button>
                {user && (
                  <Chip
                    avatar={<Avatar sx={{ bgcolor: 'rgba(255, 255, 255, 0.2)' }}>{user.email?.[0]?.toUpperCase()}</Avatar>}
                    label={user.email}
                    sx={{
                      ml: 1,
                      mr: 1,
                      bgcolor: 'rgba(255, 255, 255, 0.1)',
                      color: 'white',
                      '& .MuiChip-label': {
                        fontWeight: 500,
                      },
                    }}
                  />
                )}
                <Button
                  color="inherit"
                  startIcon={<LogoutIcon />}
                  onClick={handleLogout}
                  sx={{
                    borderRadius: 2,
                    px: 2,
                    py: 1,
                    '&:hover': {
                      backgroundColor: 'rgba(255, 255, 255, 0.15)',
                    },
                    transition: 'all 0.2s',
                  }}
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
                  backgroundColor: isActive('/login') ? 'rgba(255, 255, 255, 0.2)' : 'transparent',
                  borderRadius: 2,
                  px: 2,
                  py: 1,
                  '&:hover': {
                    backgroundColor: 'rgba(255, 255, 255, 0.15)',
                  },
                  transition: 'all 0.2s',
                }}
                aria-label={t('login')}
              >
                {t('login')}
              </Button>
            )}
          </Box>
        </Toolbar>
      </Container>
    </AppBar>
  );
};

export default Navigation;
