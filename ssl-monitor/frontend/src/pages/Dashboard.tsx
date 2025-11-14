import React, { useState, useEffect } from 'react';
import {
  Container,
  Typography,
  Paper,
  Grid,
  Card,
  CardContent,
  Box,
  Chip,
  CircularProgress,
  Alert,
} from '@mui/material';
import {
  CheckCircle as CheckCircleIcon,
  Error as ErrorIcon,
  Domain as DomainIcon,
} from '@mui/icons-material';
import { useLanguage } from '../hooks/useLanguage';
import { getStats, getHistory } from '../services/api';
import AlertsDisplay from '../components/AlertsDisplay';

interface Stats {
  total_checks: number;
  successful_checks: number;
  error_checks: number;
  unique_domains: number;
}

interface HistoryItem {
  id: number;
  domain: string;
  status: string;
  checked_at: string;
  ssl_status: string;
  data?: {
    ssl?: {
      daysUntilExpiration?: number;
    };
  };
}

const Dashboard: React.FC = () => {
  const { t } = useLanguage();
  const [stats, setStats] = useState<Stats | null>(null);
  const [recentChecks, setRecentChecks] = useState<HistoryItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        setError(null);
        
        // Fetch stats and history in parallel
        const [statsResponse, historyResponse] = await Promise.all([
          getStats(),
          getHistory(undefined, 5)
        ]);
        
        setStats(statsResponse.stats);
        // Ensure history is an array before setting state
        const history = historyResponse.history;
        setRecentChecks(Array.isArray(history) ? history : []);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to fetch data');
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  if (loading) {
    return (
      <Container maxWidth="lg" sx={{ mt: 4, mb: 4, display: 'flex', justifyContent: 'center' }}>
        <CircularProgress />
      </Container>
    );
  }

  if (error) {
    return (
      <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
        <Alert severity="error">{error}</Alert>
      </Container>
    );
  }

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
              <Typography variant="h4">{stats?.total_checks || 0}</Typography>
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
                {stats?.successful_checks || 0}
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
                {stats?.error_checks || 0}
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
                {stats?.unique_domains || 0}
              </Typography>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* Recent Checks */}
      <Grid container spacing={3}>
        <Grid item xs={12} md={6}>
          <AlertsDisplay unresolvedOnly={true} limit={10} />
        </Grid>
        <Grid item xs={12} md={6}>
          <Paper elevation={2} sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>
              Recent SSL Checks
            </Typography>
            {recentChecks.length === 0 ? (
              <Typography variant="body2" color="text.secondary">
                No checks found. Start by adding a domain to monitor.
              </Typography>
            ) : (
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
                              <Typography variant="h6">{check.domain || 'N/A'}</Typography>
                              <Typography variant="body2" color="text.secondary">
                                {t('checkedAt')}: {new Date(check.checked_at).toLocaleString()}
                              </Typography>
                            </Box>
                          </Box>
                          <Box display="flex" gap={1}>
                            <Chip
                              label={check.status === 'success' ? t('statusSuccess') : t('statusError')}
                              color={check.status === 'success' ? 'success' : 'error'}
                            />
                            {check.status === 'success' && check.data?.ssl?.daysUntilExpiration != null && (
                              <Chip
                                label={`${check.data.ssl.daysUntilExpiration} days`}
                                color={check.data.ssl.daysUntilExpiration < 30 ? 'warning' : 'default'}
                              />
                            )}
                          </Box>
                        </Box>
                      </CardContent>
                    </Card>
                  </Grid>
                ))}
              </Grid>
            )}
          </Paper>
        </Grid>
      </Grid>
    </Container>
  );
};

export default Dashboard;
