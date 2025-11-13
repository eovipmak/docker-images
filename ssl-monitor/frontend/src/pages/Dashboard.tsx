import React from 'react';
import {
  Container,
  Typography,
  Paper,
  Grid,
  Card,
  CardContent,
  Box,
  Chip,
} from '@mui/material';
import {
  CheckCircle as CheckCircleIcon,
  Error as ErrorIcon,
  Domain as DomainIcon,
} from '@mui/icons-material';
import { useLanguage } from '../hooks/useLanguage';

const Dashboard: React.FC = () => {
  const { t } = useLanguage();

  // Hardcoded data for demonstration
  const stats = {
    totalChecks: 125,
    successfulChecks: 98,
    errorChecks: 27,
    uniqueDomains: 42,
  };

  const recentChecks = [
    {
      id: 1,
      domain: 'example.com',
      status: 'success' as const,
      checkedAt: '2024-01-15T10:30:00Z',
      daysUntilExpiration: 45,
    },
    {
      id: 2,
      domain: 'google.com',
      status: 'success' as const,
      checkedAt: '2024-01-15T09:15:00Z',
      daysUntilExpiration: 89,
    },
    {
      id: 3,
      domain: 'expired-cert.com',
      status: 'error' as const,
      checkedAt: '2024-01-15T08:45:00Z',
      daysUntilExpiration: -5,
    },
  ];

  return (
    <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
      <Typography variant="h4" component="h1" gutterBottom>
        {t('dashboard')}
      </Typography>

      {/* Stats Cards */}
      <Grid container spacing={3} sx={{ mb: 4 }}>
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent>
              <Typography color="text.secondary" gutterBottom variant="body2">
                Total Checks
              </Typography>
              <Typography variant="h4">{stats.totalChecks}</Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent>
              <Box display="flex" alignItems="center" gap={1}>
                <CheckCircleIcon color="success" />
                <Typography color="text.secondary" gutterBottom variant="body2">
                  Successful
                </Typography>
              </Box>
              <Typography variant="h4" color="success.main">
                {stats.successfulChecks}
              </Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent>
              <Box display="flex" alignItems="center" gap={1}>
                <ErrorIcon color="error" />
                <Typography color="text.secondary" gutterBottom variant="body2">
                  Errors
                </Typography>
              </Box>
              <Typography variant="h4" color="error.main">
                {stats.errorChecks}
              </Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent>
              <Box display="flex" alignItems="center" gap={1}>
                <DomainIcon color="primary" />
                <Typography color="text.secondary" gutterBottom variant="body2">
                  Unique Domains
                </Typography>
              </Box>
              <Typography variant="h4" color="primary.main">
                {stats.uniqueDomains}
              </Typography>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* Recent Checks */}
      <Paper elevation={2} sx={{ p: 3 }}>
        <Typography variant="h6" gutterBottom>
          Recent SSL Checks
        </Typography>
        <Grid container spacing={2}>
          {recentChecks.map((check) => (
            <Grid item xs={12} key={check.id}>
              <Card variant="outlined">
                <CardContent>
                  <Box display="flex" justifyContent="space-between" alignItems="center">
                    <Box display="flex" alignItems="center" gap={2}>
                      {check.status === 'success' ? (
                        <CheckCircleIcon color="success" />
                      ) : (
                        <ErrorIcon color="error" />
                      )}
                      <Box>
                        <Typography variant="h6">{check.domain}</Typography>
                        <Typography variant="body2" color="text.secondary">
                          {t('checkedAt')}: {new Date(check.checkedAt).toLocaleString()}
                        </Typography>
                      </Box>
                    </Box>
                    <Box display="flex" gap={1}>
                      <Chip
                        label={check.status === 'success' ? t('statusSuccess') : t('statusError')}
                        color={check.status === 'success' ? 'success' : 'error'}
                      />
                      {check.status === 'success' && (
                        <Chip
                          label={`${check.daysUntilExpiration} days`}
                          color={check.daysUntilExpiration < 30 ? 'warning' : 'default'}
                        />
                      )}
                    </Box>
                  </Box>
                </CardContent>
              </Card>
            </Grid>
          ))}
        </Grid>
      </Paper>
    </Container>
  );
};

export default Dashboard;
