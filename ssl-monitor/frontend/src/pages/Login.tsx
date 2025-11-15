import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Container,
  Typography,
  Paper,
  Box,
  TextField,
  Button,
  Alert,
  Tabs,
  Tab,
  Fade,
  Zoom,
} from '@mui/material';
import {
  Security as SecurityIcon,
  Lock as LockIcon,
  Email as EmailIcon,
} from '@mui/icons-material';
import { useAuth } from '../hooks/useAuth';
import { useLanguage } from '../hooks/useLanguage';

const Login: React.FC = () => {
  const { t } = useLanguage();
  const navigate = useNavigate();
  const { login, register } = useAuth();
  
  const [tab, setTab] = useState(0);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [loading, setLoading] = useState(false);

  const handleTabChange = (_event: React.SyntheticEvent, newValue: number) => {
    setTab(newValue);
    setError('');
    setSuccess('');
  };

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setSuccess('');
    setLoading(true);

    try {
      await login(email, password);
      navigate('/dashboard');
    } catch (err: unknown) {
      if (err instanceof Error) {
        setError(err.message || 'Login failed. Please check your credentials.');
      } else {
        setError('Login failed. Please check your credentials.');
      }
    } finally {
      setLoading(false);
    }
  };

  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setSuccess('');
    setLoading(true);

    if (password !== confirmPassword) {
      setError('Passwords do not match');
      setLoading(false);
      return;
    }

    if (password.length < 8) {
      setError('Password must be at least 8 characters long');
      setLoading(false);
      return;
    }

    const hasUpperCase = /[A-Z]/.test(password);
    const hasLowerCase = /[a-z]/.test(password);
    const hasDigit = /[0-9]/.test(password);
    const hasSpecialChar = /[!@#$%^&*(),.?":{}|<>]/.test(password);

    if (!hasUpperCase) {
      setError('Password must contain at least one uppercase letter');
      setLoading(false);
      return;
    }

    if (!hasLowerCase) {
      setError('Password must contain at least one lowercase letter');
      setLoading(false);
      return;
    }

    if (!hasDigit) {
      setError('Password must contain at least one digit');
      setLoading(false);
      return;
    }

    if (!hasSpecialChar) {
      setError('Password must contain at least one special character (!@#$%^&*(),.?":{}|<>)');
      setLoading(false);
      return;
    }

    try {
      await register(email, password);
      setSuccess('Registration successful! Please login.');
      setEmail('');
      setPassword('');
      setConfirmPassword('');
      setTab(0);
    } catch (err: unknown) {
      if (err instanceof Error) {
        setError(err.message || 'Registration failed. Please try again.');
      } else {
        setError('Registration failed. Please try again.');
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <Box
      sx={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 50%, #ec4899 100%)',
        position: 'relative',
        overflow: 'hidden',
        '&::before': {
          content: '""',
          position: 'absolute',
          width: '100%',
          height: '100%',
          background: 'radial-gradient(circle at 20% 50%, rgba(255, 255, 255, 0.1) 0%, transparent 50%), radial-gradient(circle at 80% 80%, rgba(255, 255, 255, 0.1) 0%, transparent 50%)',
        }
      }}
    >
      <Container maxWidth="sm" sx={{ position: 'relative', zIndex: 1 }}>
        <Fade in={true} timeout={600}>
          <Paper 
            elevation={0}
            sx={{ 
              p: 5,
              borderRadius: 4,
              background: 'rgba(255, 255, 255, 0.95)',
              backdropFilter: 'blur(10px)',
              border: '1px solid rgba(255, 255, 255, 0.2)',
            }}
          >
            <Box textAlign="center" mb={4}>
              <Zoom in={true} timeout={400}>
                <Box
                  sx={{
                    display: 'inline-flex',
                    p: 2,
                    borderRadius: 3,
                    background: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)',
                    color: 'white',
                    mb: 2,
                  }}
                >
                  <SecurityIcon sx={{ fontSize: 40 }} />
                </Box>
              </Zoom>
              <Typography 
                variant="h3" 
                component="h1" 
                gutterBottom 
                fontWeight="bold"
                sx={{
                  background: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)',
                  WebkitBackgroundClip: 'text',
                  WebkitTextFillColor: 'transparent',
                  backgroundClip: 'text',
                }}
              >
                SSL Monitor
              </Typography>
              <Typography variant="body1" color="text.secondary">
                Secure SSL certificate monitoring platform
              </Typography>
            </Box>
            
            <Tabs 
              value={tab} 
              onChange={handleTabChange} 
              centered 
              sx={{ 
                mb: 4,
                '& .MuiTab-root': {
                  textTransform: 'none',
                  fontWeight: 600,
                  fontSize: '1rem',
                },
                '& .Mui-selected': {
                  color: '#6366f1',
                },
                '& .MuiTabs-indicator': {
                  background: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)',
                  height: 3,
                  borderRadius: 2,
                }
              }}
            >
              <Tab label={t('login')} />
              <Tab label="Register" />
            </Tabs>

            {error && (
              <Alert 
                severity="error" 
                sx={{ 
                  mb: 3,
                  borderRadius: 2,
                }}
                onClose={() => setError('')}
              >
                {error}
              </Alert>
            )}

            {success && (
              <Alert 
                severity="success" 
                sx={{ 
                  mb: 3,
                  borderRadius: 2,
                }}
                onClose={() => setSuccess('')}
              >
                {success}
              </Alert>
            )}

            {tab === 0 ? (
              <Box component="form" onSubmit={handleLogin}>
                <TextField
                  fullWidth
                  label="Email"
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  margin="normal"
                  required
                  autoFocus
                  InputProps={{
                    startAdornment: (
                      <EmailIcon sx={{ mr: 1, color: 'text.secondary' }} />
                    ),
                  }}
                  sx={{
                    '& .MuiOutlinedInput-root': {
                      borderRadius: 2,
                    }
                  }}
                />
                <TextField
                  fullWidth
                  label="Password"
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  margin="normal"
                  required
                  InputProps={{
                    startAdornment: (
                      <LockIcon sx={{ mr: 1, color: 'text.secondary' }} />
                    ),
                  }}
                  sx={{
                    '& .MuiOutlinedInput-root': {
                      borderRadius: 2,
                    }
                  }}
                />
                <Button
                  type="submit"
                  fullWidth
                  variant="contained"
                  size="large"
                  disabled={loading}
                  sx={{ 
                    mt: 3,
                    py: 1.5,
                    borderRadius: 2,
                    fontSize: '1rem',
                    fontWeight: 600,
                  }}
                >
                  {loading ? 'Logging in...' : t('login')}
                </Button>
              </Box>
            ) : (
              <Box component="form" onSubmit={handleRegister}>
                <TextField
                  fullWidth
                  label="Email"
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  margin="normal"
                  required
                  autoFocus
                  InputProps={{
                    startAdornment: (
                      <EmailIcon sx={{ mr: 1, color: 'text.secondary' }} />
                    ),
                  }}
                  sx={{
                    '& .MuiOutlinedInput-root': {
                      borderRadius: 2,
                    }
                  }}
                />
                <TextField
                  fullWidth
                  label="Password"
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  margin="normal"
                  required
                  InputProps={{
                    startAdornment: (
                      <LockIcon sx={{ mr: 1, color: 'text.secondary' }} />
                    ),
                  }}
                  sx={{
                    '& .MuiOutlinedInput-root': {
                      borderRadius: 2,
                    }
                  }}
                />
                <TextField
                  fullWidth
                  label="Confirm Password"
                  type="password"
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  margin="normal"
                  required
                  InputProps={{
                    startAdornment: (
                      <LockIcon sx={{ mr: 1, color: 'text.secondary' }} />
                    ),
                  }}
                  sx={{
                    '& .MuiOutlinedInput-root': {
                      borderRadius: 2,
                    }
                  }}
                />
                <Button
                  type="submit"
                  fullWidth
                  variant="contained"
                  size="large"
                  disabled={loading}
                  sx={{ 
                    mt: 3,
                    py: 1.5,
                    borderRadius: 2,
                    fontSize: '1rem',
                    fontWeight: 600,
                  }}
                >
                  {loading ? 'Registering...' : 'Register'}
                </Button>
              </Box>
            )}
          </Paper>
        </Fade>
      </Container>
    </Box>
  );
};

export default Login;
