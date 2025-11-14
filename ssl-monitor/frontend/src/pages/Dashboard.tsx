import React, { useState, useEffect, useCallback, useRef } from 'react';
import {
  Container,
  Typography,
  Paper,
  Grid,
  Card,
  CardContent,
  CardActionArea,
  Box,
  Chip,
  CircularProgress,
  Alert,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Snackbar,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableRow,
  useTheme,
  useMediaQuery,
  IconButton,
  Tooltip,
  Menu,
  MenuItem,
  ListItemIcon,
  ListItemText,
  LinearProgress,
  Select,
  FormControl,
  InputLabel,
} from '@mui/material';
import {
  CheckCircle as CheckCircleIcon,
  Error as ErrorIcon,
  Domain as DomainIcon,
  AccessTime as AccessTimeIcon,
  Refresh as RefreshIcon,
  Delete as DeleteIcon,
  Settings as SettingsIcon,
  NotificationsOff as NotificationsOffIcon,
  Notifications as NotificationsIcon,
  BugReport as BugReportIcon,
} from '@mui/icons-material';
import { useLanguage } from '../hooks/useLanguage';
import { getStats, getHistory, deleteDomain, updateMonitor, testDomainAlert } from '../services/api';
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

interface SSLInfo {
  daysUntilExpiration?: number;
  notAfter?: string;
  notBefore?: string;
  issuer?: {
    commonName?: string;
    organizationName?: string;
    countryName?: string;
  };
  subject?: {
    commonName?: string;
    organizationName?: string;
  };
  serialNumber?: string;
  signatureAlgorithm?: string;
  tlsVersion?: string;
  cipherSuite?: string;
}

interface DomainStatus {
  domain: string;
  ip?: string;
  port: number;
  status: string;
  ssl_status: string;
  last_checked: string;
  ssl_info: SSLInfo;
  monitor?: {
    alerts_enabled: boolean;
    check_interval: number;
    webhook_url?: string;
  };
  uptime?: {
    uptime_percentage: number | null;
    total_checks: number;
    successful_checks: number;
    failed_checks: number;
    days_tracked: number;
  };
}

const Dashboard: React.FC = () => {
  const { t } = useLanguage();
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  
  const [stats, setStats] = useState<Stats | null>(null);
  const [recentChecks, setRecentChecks] = useState<HistoryItem[]>([]);
  const [domains, setDomains] = useState<DomainStatus[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedDomain, setSelectedDomain] = useState<DomainStatus | null>(null);
  const [snackbarOpen, setSnackbarOpen] = useState(false);
  const [snackbarMessage, setSnackbarMessage] = useState('');
  const [snackbarSeverity, setSnackbarSeverity] = useState<'success' | 'error' | 'warning'>('success');
  const [lastUpdate, setLastUpdate] = useState<Date>(new Date());
  const [menuAnchor, setMenuAnchor] = useState<{ element: HTMLElement; domain: DomainStatus } | null>(null);
  const [settingsDialogOpen, setSettingsDialogOpen] = useState(false);
  const [settingsDomain, setSettingsDomain] = useState<DomainStatus | null>(null);
  const [settingsCheckInterval, setSettingsCheckInterval] = useState<number>(3600);
  
  const wsRef = useRef<WebSocket | null>(null);
  const autoRefreshRef = useRef<number | null>(null);

  // Get color based on days until expiration
  const getStatusColor = (daysUntilExpiration?: number): 'success' | 'warning' | 'error' => {
    if (daysUntilExpiration === undefined || daysUntilExpiration === null) {
      return 'error';
    }
    if (daysUntilExpiration > 30) {
      return 'success'; // Green
    } else if (daysUntilExpiration >= 7) {
      return 'warning'; // Yellow
    } else {
      return 'error'; // Red
    }
  };

  // Format countdown display
  const formatCountdown = (daysUntilExpiration?: number): string => {
    if (daysUntilExpiration === undefined || daysUntilExpiration === null) {
      return 'Unknown';
    }
    if (daysUntilExpiration < 0) {
      return 'Expired';
    }
    if (daysUntilExpiration === 0) {
      return 'Expires today';
    }
    if (daysUntilExpiration === 1) {
      return '1 day';
    }
    return `${daysUntilExpiration} days`;
  };

  // Fetch domains with status
  const fetchDomainsStatus = useCallback(async () => {
    try {
      const token = localStorage.getItem('auth_token');
      if (!token) {
        return;
      }

      const response = await fetch('/api/domains/status', {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        if (data.domains) {
          setDomains(data.domains);
          setLastUpdate(new Date());
        }
      }
    } catch (err) {
      console.error('Failed to fetch domain status:', err);
    }
  }, []);

  // Fetch data function
  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      
      // Fetch stats, history, and domain status in parallel
      const [statsResponse, historyResponse] = await Promise.all([
        getStats(),
        getHistory(undefined, 5)
      ]);
      
      setStats(statsResponse.stats);
      const history = historyResponse.history;
      setRecentChecks(Array.isArray(history) ? history : []);
      
      // Fetch domain status
      await fetchDomainsStatus();
      
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to fetch data';
      setError(errorMsg);
      showSnackbar(errorMsg, 'error');
    } finally {
      setLoading(false);
    }
  }, [fetchDomainsStatus]);

  // Show snackbar notification
  const showSnackbar = (message: string, severity: 'success' | 'error' | 'warning' = 'success') => {
    setSnackbarMessage(message);
    setSnackbarSeverity(severity);
    setSnackbarOpen(true);
  };

  // Setup WebSocket connection
  useEffect(() => {
    const token = localStorage.getItem('auth_token');
    if (!token) {
      return;
    }

    // Create WebSocket connection
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws/domains`;
    
    try {
      const ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = () => {
        console.log('WebSocket connected');
        // Send authentication or heartbeat
        ws.send(JSON.stringify({ type: 'ping' }));
      };

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          if (data.type === 'update') {
            // Handle domain updates
            fetchDomainsStatus();
          }
        } catch (err) {
          console.error('Failed to parse WebSocket message:', err);
        }
      };

      ws.onerror = (error) => {
        console.error('WebSocket error:', error);
      };

      ws.onclose = () => {
        console.log('WebSocket disconnected');
      };
    } catch (err) {
      console.error('Failed to create WebSocket:', err);
    }

    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, [fetchDomainsStatus]);

  // Setup auto-refresh (every 30 seconds)
  useEffect(() => {
    // Initial fetch
    fetchData();

    // Setup auto-refresh
    autoRefreshRef.current = window.setInterval(() => {
      fetchDomainsStatus();
    }, 30000); // 30 seconds

    return () => {
      if (autoRefreshRef.current) {
        window.clearInterval(autoRefreshRef.current);
      }
    };
  }, [fetchData, fetchDomainsStatus]);

  // Handle domain card click
  const handleDomainClick = (domain: DomainStatus) => {
    setSelectedDomain(domain);
  };

  // Handle modal close
  const handleCloseModal = () => {
    setSelectedDomain(null);
  };

  // Handle alert toggle for domain
  const handleToggleAlerts = async (domainName: string, currentStatus: boolean) => {
    try {
      await updateMonitor(domainName, { alerts_enabled: !currentStatus });
      showSnackbar(
        `Alerts ${!currentStatus ? 'enabled' : 'disabled'} for ${domainName}`,
        'success'
      );
      // Refresh data
      await fetchDomainsStatus();
      setMenuAnchor(null);
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to update alert settings';
      showSnackbar(errorMsg, 'error');
    }
  };

  // Handle menu open
  const handleMenuOpen = (event: React.MouseEvent<HTMLElement>, domain: DomainStatus) => {
    event.stopPropagation();
    setMenuAnchor({ element: event.currentTarget, domain });
  };

  // Handle menu close
  const handleMenuClose = () => {
    setMenuAnchor(null);
  };

  // Handle delete from menu
  const handleDeleteFromMenu = async () => {
    if (menuAnchor) {
      await handleDeleteDomain(menuAnchor.domain.domain);
      setMenuAnchor(null);
    }
  };

  // Handle toggle alerts from menu
  const handleToggleAlertsFromMenu = async () => {
    if (menuAnchor) {
      const alertsEnabled = menuAnchor.domain.monitor?.alerts_enabled ?? true;
      await handleToggleAlerts(menuAnchor.domain.domain, alertsEnabled);
    }
  };

  // Handle test alert from menu
  const handleTestAlertFromMenu = async () => {
    if (!menuAnchor) return;
    
    const domainName = menuAnchor.domain.domain;
    setMenuAnchor(null);
    
    try {
      const result = await testDomainAlert(domainName);
      showSnackbar(result.message || `Test alert sent for ${domainName}`, 'success');
    } catch (err: unknown) {
      const errorMsg = err instanceof Error 
        ? err.message 
        : (err && typeof err === 'object' && 'response' in err && err.response && typeof err.response === 'object' && 'data' in err.response && err.response.data && typeof err.response.data === 'object' && 'detail' in err.response.data)
          ? String(err.response.data.detail)
          : 'Failed to send test alert';
      showSnackbar(errorMsg, 'error');
    }
  };

  // Handle settings from menu
  const handleSettingsFromMenu = () => {
    if (!menuAnchor) return;
    
    setSettingsDomain(menuAnchor.domain);
    setSettingsCheckInterval(menuAnchor.domain.monitor?.check_interval || 3600);
    setSettingsDialogOpen(true);
    setMenuAnchor(null);
  };

  // Handle close settings dialog
  const handleCloseSettingsDialog = () => {
    setSettingsDialogOpen(false);
    setSettingsDomain(null);
  };

  // Handle save settings
  const handleSaveSettings = async () => {
    if (!settingsDomain) return;
    
    try {
      await updateMonitor(settingsDomain.domain, { check_interval: settingsCheckInterval });
      showSnackbar(`Settings updated for ${settingsDomain.domain}`, 'success');
      setSettingsDialogOpen(false);
      setSettingsDomain(null);
      // Refresh data
      await fetchData();
    } catch (err: unknown) {
      const errorMsg = err instanceof Error 
        ? err.message 
        : (err && typeof err === 'object' && 'response' in err && err.response && typeof err.response === 'object' && 'data' in err.response && err.response.data && typeof err.response.data === 'object' && 'detail' in err.response.data)
          ? String(err.response.data.detail)
          : 'Failed to update settings';
      showSnackbar(errorMsg, 'error');
    }
  };

  // Handle domain delete
  const handleDeleteDomain = async (domainName: string) => {
    if (!window.confirm(`Are you sure you want to delete monitoring for ${domainName}?`)) {
      return;
    }

    try {
      await deleteDomain(domainName);
      showSnackbar(`Domain ${domainName} deleted successfully`, 'success');
      // Refresh data
      await fetchData();
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to delete domain';
      showSnackbar(errorMsg, 'error');
    }
  };

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
    <Container maxWidth="xl" sx={{ mt: 4, mb: 4 }}>
      {/* Header Section */}
      <Box 
        display="flex" 
        justifyContent="space-between" 
        alignItems="center" 
        mb={4}
        sx={{
          pb: 2,
          borderBottom: `2px solid ${theme.palette.divider}`,
        }}
      >
        <Box>
          <Typography variant="h4" component="h1" fontWeight="bold" gutterBottom>
            SSL Monitor Dashboard
          </Typography>
          <Typography variant="body2" color="text.secondary">
            Monitor SSL certificates for your domains in real-time
          </Typography>
        </Box>
        <Box display="flex" alignItems="center" gap={2}>
          <Chip
            icon={<AccessTimeIcon />}
            label={`Last updated ${lastUpdate.toLocaleTimeString()}`}
            size="small"
            variant="outlined"
          />
          <Button
            startIcon={<RefreshIcon />}
            onClick={fetchData}
            variant="outlined"
            size="medium"
          >
            Refresh Data
          </Button>
        </Box>
      </Box>

      {/* Stats Cards */}
      <Grid container spacing={3} sx={{ mb: 4 }}>
        <Grid item xs={12} sm={6} md={3}>
          <Card 
            elevation={2}
            sx={{ 
              transition: 'transform 0.2s, box-shadow 0.2s',
              '&:hover': {
                transform: 'translateY(-4px)',
                boxShadow: 4,
              }
            }}
          >
            <CardContent>
              <Typography color="text.secondary" gutterBottom variant="body2" fontWeight="medium">
                Total SSL Checks
              </Typography>
              <Typography variant="h3" fontWeight="bold">{stats?.total_checks || 0}</Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <Card
            elevation={2}
            sx={{ 
              transition: 'transform 0.2s, box-shadow 0.2s',
              '&:hover': {
                transform: 'translateY(-4px)',
                boxShadow: 4,
              }
            }}
          >
            <CardContent>
              <Box display="flex" alignItems="center" gap={1} mb={1}>
                <CheckCircleIcon color="success" />
                <Typography color="text.secondary" variant="body2" fontWeight="medium">
                  Valid Certificates
                </Typography>
              </Box>
              <Typography variant="h3" color="success.main" fontWeight="bold">
                {stats?.successful_checks || 0}
              </Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <Card
            elevation={2}
            sx={{ 
              transition: 'transform 0.2s, box-shadow 0.2s',
              '&:hover': {
                transform: 'translateY(-4px)',
                boxShadow: 4,
              }
            }}
          >
            <CardContent>
              <Box display="flex" alignItems="center" gap={1} mb={1}>
                <ErrorIcon color="error" />
                <Typography color="text.secondary" variant="body2" fontWeight="medium">
                  Certificate Errors
                </Typography>
              </Box>
              <Typography variant="h3" color="error.main" fontWeight="bold">
                {stats?.error_checks || 0}
              </Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <Card
            elevation={2}
            sx={{ 
              transition: 'transform 0.2s, box-shadow 0.2s',
              '&:hover': {
                transform: 'translateY(-4px)',
                boxShadow: 4,
              }
            }}
          >
            <CardContent>
              <Box display="flex" alignItems="center" gap={1} mb={1}>
                <DomainIcon color="primary" />
                <Typography color="text.secondary" variant="body2" fontWeight="medium">
                  Monitored Domains
                </Typography>
              </Box>
              <Typography variant="h3" color="primary.main" fontWeight="bold">
                {stats?.unique_domains || 0}
              </Typography>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* Monitored Domains Section */}
      <Box mb={3}>
        <Typography variant="h5" gutterBottom fontWeight="bold">
          Monitored Domains
        </Typography>
        <Typography variant="body2" color="text.secondary">
          Real-time SSL certificate monitoring for your domains
        </Typography>
      </Box>
      
      {domains.length === 0 ? (
        <Alert severity="info" sx={{ mb: 4 }}>
          No domains are being monitored. Add a domain to get started.
        </Alert>
      ) : (
        <Grid 
          container 
          spacing={3} 
          sx={{ 
            mb: 4,
            // Responsive grid breakpoints
            '@media (max-width: 480px)': {
              spacing: 2,
            },
          }}
        >
          {domains.map((domain, index) => {
            const statusColor = getStatusColor(domain.ssl_info?.daysUntilExpiration);
            const countdown = formatCountdown(domain.ssl_info?.daysUntilExpiration);
            
            return (
              <Grid 
                item 
                xs={12} 
                sm={6} 
                md={4} 
                lg={3}
                key={`${domain.domain}-${index}`}
              >
                <Card 
                  elevation={2}
                  sx={{ 
                    height: '100%',
                    borderLeft: `4px solid ${
                      statusColor === 'success' 
                        ? theme.palette.success.main 
                        : statusColor === 'warning' 
                        ? theme.palette.warning.main 
                        : theme.palette.error.main
                    }`,
                    transition: 'transform 0.2s, box-shadow 0.2s',
                    '&:hover': {
                      transform: 'translateY(-4px)',
                      boxShadow: 6,
                    }
                  }}
                >
                  <CardActionArea onClick={() => handleDomainClick(domain)}>
                    <CardContent>
                      <Box display="flex" alignItems="flex-start" justifyContent="space-between" mb={1}>
                        <Box display="flex" alignItems="center" gap={1} flex={1} minWidth={0}>
                          <DomainIcon color={statusColor} />
                          <Typography 
                            variant="h6" 
                            component="div"
                            sx={{
                              overflow: 'hidden',
                              textOverflow: 'ellipsis',
                              whiteSpace: 'nowrap',
                            }}
                          >
                            {domain.domain}
                          </Typography>
                        </Box>
                        <Tooltip title="Actions">
                          <IconButton
                            size="small"
                            onClick={(e) => handleMenuOpen(e, domain)}
                            sx={{ ml: 1 }}
                          >
                            <SettingsIcon fontSize="small" />
                          </IconButton>
                        </Tooltip>
                      </Box>
                      
                      <Box display="flex" alignItems="center" gap={1} mb={1}>
                        <AccessTimeIcon fontSize="small" color="action" />
                        <Typography variant="body2" color="text.secondary">
                          {countdown}
                        </Typography>
                      </Box>
                      
                      <Box display="flex" gap={1} flexWrap="wrap" mt={2}>
                        <Chip
                          label={domain.ssl_status === 'success' ? 'Valid SSL' : 'SSL Error'}
                          color={domain.ssl_status === 'success' ? 'success' : 'error'}
                          size="small"
                        />
                        {domain.ssl_info?.daysUntilExpiration !== undefined && (
                          <Chip
                            label={`${domain.ssl_info.daysUntilExpiration} days left`}
                            color={statusColor}
                            size="small"
                          />
                        )}
                        {domain.monitor?.alerts_enabled === false && (
                          <Chip
                            icon={<NotificationsOffIcon />}
                            label="Alerts Disabled"
                            size="small"
                            variant="outlined"
                            color="warning"
                          />
                        )}
                      </Box>
                      
                      {domain.ssl_info?.issuer?.organizationName && (
                        <Typography variant="caption" color="text.secondary" display="block" mt={1}>
                          Certificate Issuer: {domain.ssl_info.issuer.organizationName}
                        </Typography>
                      )}
                      
                      {/* Uptime Bar */}
                      {domain.uptime && domain.uptime.uptime_percentage !== null && (
                        <Box mt={2}>
                          <Box display="flex" justifyContent="space-between" alignItems="center" mb={0.5}>
                            <Typography variant="caption" color="text.secondary" fontWeight="medium">
                              30-Day Uptime
                            </Typography>
                            <Typography 
                              variant="caption" 
                              fontWeight="bold"
                              color={
                                domain.uptime.uptime_percentage >= 99 
                                  ? 'success.main' 
                                  : domain.uptime.uptime_percentage >= 95 
                                  ? 'warning.main' 
                                  : 'error.main'
                              }
                            >
                              {domain.uptime.uptime_percentage.toFixed(2)}%
                            </Typography>
                          </Box>
                          <LinearProgress 
                            variant="determinate" 
                            value={domain.uptime.uptime_percentage} 
                            sx={{
                              height: 6,
                              borderRadius: 3,
                              backgroundColor: 'rgba(0, 0, 0, 0.1)',
                              '& .MuiLinearProgress-bar': {
                                borderRadius: 3,
                                backgroundColor: 
                                  domain.uptime.uptime_percentage >= 99 
                                    ? theme.palette.success.main 
                                    : domain.uptime.uptime_percentage >= 95 
                                    ? theme.palette.warning.main 
                                    : theme.palette.error.main
                              }
                            }}
                          />
                          <Typography variant="caption" color="text.secondary" display="block" mt={0.5}>
                            {domain.uptime.successful_checks} of {domain.uptime.total_checks} checks successful
                          </Typography>
                        </Box>
                      )}
                    </CardContent>
                  </CardActionArea>
                </Card>
              </Grid>
            );
          })}
        </Grid>
      )}

      {/* Recent Checks and Alerts */}
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
                        <Box display="flex" justifyContent="space-between" alignItems="center" flexWrap="wrap" gap={1}>
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
                          <Box display="flex" gap={1} flexWrap="wrap">
                            <Chip
                              label={check.status === 'success' ? t('statusSuccess') : t('statusError')}
                              color={check.status === 'success' ? 'success' : 'error'}
                              size="small"
                            />
                            {check.status === 'success' && check.data?.ssl?.daysUntilExpiration != null && (
                              <Chip
                                label={`${check.data.ssl.daysUntilExpiration} days`}
                                color={getStatusColor(check.data.ssl.daysUntilExpiration)}
                                size="small"
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

      {/* Certificate Details Modal */}
      <Dialog 
        open={!!selectedDomain} 
        onClose={handleCloseModal}
        maxWidth="md"
        fullWidth
        fullScreen={isMobile}
      >
        <DialogTitle>
          <Box display="flex" alignItems="center" gap={1}>
            <DomainIcon color="primary" />
            <Box>
              <Typography variant="h6" fontWeight="bold">
                SSL Certificate Details
              </Typography>
              <Typography variant="body2" color="text.secondary">
                {selectedDomain?.domain}
              </Typography>
            </Box>
          </Box>
        </DialogTitle>
        <DialogContent>
          {selectedDomain && (
            <TableContainer>
              <Table>
                <TableBody>
                  <TableRow>
                    <TableCell component="th" scope="row">
                      <strong>Certificate Status</strong>
                    </TableCell>
                    <TableCell>
                      <Chip
                        label={selectedDomain.ssl_status === 'success' ? 'Valid' : 'Error'}
                        color={selectedDomain.ssl_status === 'success' ? 'success' : 'error'}
                        size="small"
                      />
                    </TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell component="th" scope="row">
                      <strong>Domain</strong>
                    </TableCell>
                    <TableCell>{selectedDomain.domain}</TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell component="th" scope="row">
                      <strong>IP Address</strong>
                    </TableCell>
                    <TableCell>{selectedDomain.ip || 'N/A'}</TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell component="th" scope="row">
                      <strong>Port</strong>
                    </TableCell>
                    <TableCell>{selectedDomain.port}</TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell component="th" scope="row">
                      <strong>Days Until Expiration</strong>
                    </TableCell>
                    <TableCell>
                      <Box display="flex" alignItems="center" gap={1}>
                        <Typography>
                          {formatCountdown(selectedDomain.ssl_info?.daysUntilExpiration)}
                        </Typography>
                        {selectedDomain.ssl_info?.daysUntilExpiration !== undefined && (
                          <Chip
                            label={`${selectedDomain.ssl_info.daysUntilExpiration}d`}
                            color={getStatusColor(selectedDomain.ssl_info.daysUntilExpiration)}
                            size="small"
                          />
                        )}
                      </Box>
                    </TableCell>
                  </TableRow>
                  {selectedDomain.ssl_info?.notAfter && (
                    <TableRow>
                      <TableCell component="th" scope="row">
                        <strong>Expires On</strong>
                      </TableCell>
                      <TableCell>
                        {new Date(selectedDomain.ssl_info.notAfter).toLocaleString()}
                      </TableCell>
                    </TableRow>
                  )}
                  {selectedDomain.ssl_info?.notBefore && (
                    <TableRow>
                      <TableCell component="th" scope="row">
                        <strong>Valid From</strong>
                      </TableCell>
                      <TableCell>
                        {new Date(selectedDomain.ssl_info.notBefore).toLocaleString()}
                      </TableCell>
                    </TableRow>
                  )}
                  {selectedDomain.ssl_info?.issuer?.commonName && (
                    <TableRow>
                      <TableCell component="th" scope="row">
                        <strong>Issuer</strong>
                      </TableCell>
                      <TableCell>
                        {selectedDomain.ssl_info.issuer.organizationName || selectedDomain.ssl_info.issuer.commonName}
                      </TableCell>
                    </TableRow>
                  )}
                  {selectedDomain.ssl_info?.subject?.commonName && (
                    <TableRow>
                      <TableCell component="th" scope="row">
                        <strong>Subject</strong>
                      </TableCell>
                      <TableCell>{selectedDomain.ssl_info.subject.commonName}</TableCell>
                    </TableRow>
                  )}
                  {selectedDomain.ssl_info?.serialNumber && (
                    <TableRow>
                      <TableCell component="th" scope="row">
                        <strong>Serial Number</strong>
                      </TableCell>
                      <TableCell sx={{ wordBreak: 'break-all' }}>
                        {selectedDomain.ssl_info.serialNumber}
                      </TableCell>
                    </TableRow>
                  )}
                  {selectedDomain.ssl_info?.signatureAlgorithm && (
                    <TableRow>
                      <TableCell component="th" scope="row">
                        <strong>Signature Algorithm</strong>
                      </TableCell>
                      <TableCell>{selectedDomain.ssl_info.signatureAlgorithm}</TableCell>
                    </TableRow>
                  )}
                  {selectedDomain.ssl_info?.tlsVersion && (
                    <TableRow>
                      <TableCell component="th" scope="row">
                        <strong>TLS Version</strong>
                      </TableCell>
                      <TableCell>{selectedDomain.ssl_info.tlsVersion}</TableCell>
                    </TableRow>
                  )}
                  {selectedDomain.ssl_info?.cipherSuite && (
                    <TableRow>
                      <TableCell component="th" scope="row">
                        <strong>Cipher Suite</strong>
                      </TableCell>
                      <TableCell sx={{ wordBreak: 'break-all' }}>
                        {selectedDomain.ssl_info.cipherSuite}
                      </TableCell>
                    </TableRow>
                  )}
                  <TableRow>
                    <TableCell component="th" scope="row">
                      <strong>Last Checked</strong>
                    </TableCell>
                    <TableCell>
                      {new Date(selectedDomain.last_checked).toLocaleString()}
                    </TableCell>
                  </TableRow>
                </TableBody>
              </Table>
            </TableContainer>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseModal}>Close</Button>
        </DialogActions>
      </Dialog>

      {/* Monitor Settings Dialog */}
      <Dialog
        open={settingsDialogOpen}
        onClose={handleCloseSettingsDialog}
        maxWidth="sm"
        fullWidth
      >
        <DialogTitle>
          Monitor Settings
          {settingsDomain && (
            <Typography variant="body2" color="text.secondary">
              {settingsDomain.domain}
            </Typography>
          )}
        </DialogTitle>
        <DialogContent>
          <Box sx={{ pt: 2 }}>
            <FormControl fullWidth>
              <InputLabel id="check-interval-label">Check Interval</InputLabel>
              <Select
                labelId="check-interval-label"
                value={settingsCheckInterval}
                label="Check Interval"
                onChange={(e) => setSettingsCheckInterval(Number(e.target.value))}
              >
                <MenuItem value={3600}>1 hour</MenuItem>
                <MenuItem value={10800}>3 hours</MenuItem>
                <MenuItem value={43200}>12 hours</MenuItem>
                <MenuItem value={86400}>24 hours (1 day)</MenuItem>
              </Select>
            </FormControl>
            <Typography variant="caption" color="text.secondary" sx={{ mt: 1, display: 'block' }}>
              How often to check the SSL certificate status
            </Typography>
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseSettingsDialog}>Cancel</Button>
          <Button onClick={handleSaveSettings} variant="contained" color="primary">
            Save
          </Button>
        </DialogActions>
      </Dialog>

      {/* Snackbar for notifications */}
      <Snackbar
        open={snackbarOpen}
        autoHideDuration={6000}
        onClose={() => setSnackbarOpen(false)}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
      >
        <Alert 
          onClose={() => setSnackbarOpen(false)} 
          severity={snackbarSeverity}
          sx={{ width: '100%' }}
        >
          {snackbarMessage}
        </Alert>
      </Snackbar>

      {/* Domain Actions Menu */}
      <Menu
        anchorEl={menuAnchor?.element}
        open={Boolean(menuAnchor)}
        onClose={handleMenuClose}
      >
        <MenuItem onClick={handleToggleAlertsFromMenu}>
          <ListItemIcon>
            {menuAnchor?.domain.monitor?.alerts_enabled === false ? (
              <NotificationsIcon fontSize="small" />
            ) : (
              <NotificationsOffIcon fontSize="small" />
            )}
          </ListItemIcon>
          <ListItemText>
            {menuAnchor?.domain.monitor?.alerts_enabled === false
              ? 'Enable Alerts'
              : 'Disable Alerts'}
          </ListItemText>
        </MenuItem>
        <MenuItem onClick={handleSettingsFromMenu}>
          <ListItemIcon>
            <SettingsIcon fontSize="small" />
          </ListItemIcon>
          <ListItemText>Monitor Settings</ListItemText>
        </MenuItem>
        <MenuItem onClick={handleTestAlertFromMenu}>
          <ListItemIcon>
            <BugReportIcon fontSize="small" color="primary" />
          </ListItemIcon>
          <ListItemText>Send Test Alert</ListItemText>
        </MenuItem>
        <MenuItem onClick={handleDeleteFromMenu}>
          <ListItemIcon>
            <DeleteIcon fontSize="small" color="error" />
          </ListItemIcon>
          <ListItemText>Delete Domain</ListItemText>
        </MenuItem>
      </Menu>
    </Container>
  );
};

export default Dashboard;
