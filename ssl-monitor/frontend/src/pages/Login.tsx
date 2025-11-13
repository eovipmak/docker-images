import React from 'react';
import { Container, Typography, Paper, Box } from '@mui/material';
import { useLanguage } from '../hooks/useLanguage';

const Login: React.FC = () => {
  const { t } = useLanguage();

  return (
    <Container maxWidth="sm" sx={{ mt: 4 }}>
      <Paper elevation={2} sx={{ p: 4 }}>
        <Typography variant="h4" component="h1" gutterBottom align="center">
          {t('login')}
        </Typography>
        <Box sx={{ mt: 3 }}>
          <Typography variant="body1" color="text.secondary" align="center">
            Login page - Placeholder for future implementation
          </Typography>
        </Box>
      </Paper>
    </Container>
  );
};

export default Login;
